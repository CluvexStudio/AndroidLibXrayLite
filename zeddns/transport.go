package zeddns

import (
	"context"
	"encoding/binary"
	"errors"
	"io"
	"net"
	"sync"
	"time"

	utls "github.com/refraction-networking/utls"
)

type udpConn struct {
	up      upstream
	timeout time.Duration
}

func newUDPConn(u upstream, cfg poolConfig) queryConn {
	return &udpConn{up: u, timeout: cfg.timeout}
}

func (c *udpConn) close() {}

func (c *udpConn) query(ctx context.Context, msg []byte) ([]byte, error) {
	deadline := time.Now().Add(c.timeout)
	if d, ok := ctx.Deadline(); ok && d.Before(deadline) {
		deadline = d
	}
	conn, err := net.DialTimeout("udp", c.up.addr, time.Until(deadline))
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	if err := conn.SetDeadline(deadline); err != nil {
		return nil, err
	}
	if _, err := conn.Write(msg); err != nil {
		return nil, err
	}
	buf := make([]byte, maxDNSMessage)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			return nil, err
		}
		if n >= 12 && msgID(buf[:n]) == msgID(msg) {
			reply := make([]byte, n)
			copy(reply, buf[:n])
			return reply, nil
		}
	}
}

type streamConn struct {
	up  upstream
	cfg poolConfig

	mu  sync.Mutex
	cur *streamSession
}

func newStreamConn(u upstream, cfg poolConfig) queryConn {
	return &streamConn{up: u, cfg: cfg}
}

func (c *streamConn) close() {
	c.mu.Lock()
	s := c.cur
	c.cur = nil
	c.mu.Unlock()
	if s != nil {
		s.shutdown()
	}
}

func (c *streamConn) query(ctx context.Context, msg []byte) ([]byte, error) {
	s, err := c.session(ctx)
	if err != nil {
		return nil, err
	}

	id, ch, err := s.register()
	if err != nil {
		return nil, err
	}
	defer s.unregister(id)

	out := make([]byte, len(msg))
	copy(out, msg)
	setMsgID(out, id)

	if err := s.write(ctx, out); err != nil {
		c.drop(s)
		return nil, err
	}

	select {
	case reply := <-ch:
		setMsgID(reply, msgID(msg))
		return reply, nil
	case <-s.dead:
		return nil, errors.New("resolver connection closed")
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (c *streamConn) session(ctx context.Context) (*streamSession, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cur != nil && !c.cur.isDead() {
		return c.cur, nil
	}
	conn, err := c.dial(ctx)
	if err != nil {
		return nil, err
	}
	c.cur = newStreamSession(conn)
	return c.cur, nil
}

func (c *streamConn) drop(s *streamSession) {
	c.mu.Lock()
	if c.cur == s {
		c.cur = nil
	}
	c.mu.Unlock()
	s.shutdown()
}

func (c *streamConn) dial(ctx context.Context) (net.Conn, error) {
	conn, err := c.cfg.dial(ctx, "tcp", c.up.addr)
	if err != nil {
		return nil, err
	}
	if c.up.kind != transportDoT {
		return conn, nil
	}

	host := hostOf(c.up.addr)
	cfg := &utls.Config{ServerName: host}
	if net.ParseIP(host) != nil {
		cfg.ServerName = ""
		cfg.InsecureSkipVerify = true
	}

	hello := utls.HelloChrome_Auto
	if c.cfg.helloID != nil {
		hello = *c.cfg.helloID
	}
	uconn := utls.UClient(conn, cfg, hello)
	if err := uconn.HandshakeContext(ctx); err != nil {
		conn.Close()
		return nil, err
	}
	return uconn, nil
}

type streamSession struct {
	conn net.Conn
	dead chan struct{}

	wmu sync.Mutex

	mu      sync.Mutex
	pending map[uint16]chan []byte
	nextID  uint16
	closed  bool
}

func newStreamSession(conn net.Conn) *streamSession {
	s := &streamSession{
		conn:    conn,
		dead:    make(chan struct{}),
		pending: make(map[uint16]chan []byte),
	}
	go s.read()
	return s
}

func (s *streamSession) isDead() bool {
	select {
	case <-s.dead:
		return true
	default:
		return false
	}
}

func (s *streamSession) shutdown() {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return
	}
	s.closed = true
	close(s.dead)
	s.mu.Unlock()
	_ = s.conn.Close()
}

func (s *streamSession) register() (uint16, chan []byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return 0, nil, errors.New("resolver connection closed")
	}
	for i := 0; i <= 0xFFFF; i++ {
		s.nextID++
		if _, taken := s.pending[s.nextID]; !taken {
			ch := make(chan []byte, 1)
			s.pending[s.nextID] = ch
			return s.nextID, ch, nil
		}
	}
	return 0, nil, errors.New("no free transaction id")
}

func (s *streamSession) unregister(id uint16) {
	s.mu.Lock()
	delete(s.pending, id)
	s.mu.Unlock()
}

func (s *streamSession) write(ctx context.Context, msg []byte) error {
	if len(msg) > 0xFFFF {
		return errors.New("query too long for a stream resolver")
	}
	frame := make([]byte, 2+len(msg))
	binary.BigEndian.PutUint16(frame[:2], uint16(len(msg)))
	copy(frame[2:], msg)

	deadline, ok := ctx.Deadline()
	if !ok {
		deadline = time.Now().Add(10 * time.Second)
	}

	s.wmu.Lock()
	defer s.wmu.Unlock()
	if err := s.conn.SetWriteDeadline(deadline); err != nil {
		return err
	}
	_, err := s.conn.Write(frame)
	return err
}

func (s *streamSession) read() {
	defer contain("resolver connection")
	defer s.shutdown()
	var hdr [2]byte
	for {
		if _, err := io.ReadFull(s.conn, hdr[:]); err != nil {
			return
		}
		n := int(binary.BigEndian.Uint16(hdr[:]))
		if n == 0 || n > maxDNSMessage {
			return
		}
		buf := make([]byte, n)
		if _, err := io.ReadFull(s.conn, buf); err != nil {
			return
		}
		if n < 12 {
			continue
		}
		s.deliver(msgID(buf), buf)
	}
}

func (s *streamSession) deliver(id uint16, reply []byte) {
	s.mu.Lock()
	ch, ok := s.pending[id]
	if ok {
		delete(s.pending, id)
	}
	s.mu.Unlock()
	if ok {
		ch <- reply
	}
}
