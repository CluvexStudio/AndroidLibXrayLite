package zeddns

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"

	utls "github.com/refraction-networking/utls"
	log "github.com/sirupsen/logrus"
)

const (
	strategyAll    = "all"
	strategyRotate = "rotate"
)

const (
	maxDNSMessage    = 8192
	poolSocketBuffer = 1 << 20
)

type dialFunc func(ctx context.Context, network, addr string) (net.Conn, error)

type queryConn interface {
	query(ctx context.Context, msg []byte) ([]byte, error)
	close()
}

type poolConfig struct {
	upstreams     []upstream
	strategy      string
	perQuery      int
	timeout       time.Duration
	helloID       *utls.ClientHelloID
	dial          dialFunc
	authoritative bool
}

type resolverPool struct {
	cfg   poolConfig
	pc    *net.UDPConn
	conns []queryConn
	rr    atomic.Uint64

	closeOnce sync.Once
	done      chan struct{}
	wg        sync.WaitGroup
}

func newResolverPool(cfg poolConfig) (*resolverPool, error) {
	if len(cfg.upstreams) == 0 {
		return nil, errors.New("resolver pool needs at least one resolver")
	}
	if cfg.timeout <= 0 {
		cfg.timeout = 5 * time.Second
	}
	if cfg.perQuery <= 0 {
		cfg.perQuery = 1
	}
	if cfg.perQuery > len(cfg.upstreams) {
		cfg.perQuery = len(cfg.upstreams)
	}
	if cfg.dial == nil {
		cfg.dial = (&net.Dialer{Timeout: 10 * time.Second}).DialContext
	}

	pc, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1)})
	if err != nil {
		return nil, fmt.Errorf("resolver pool listener: %w", err)
	}
	_ = pc.SetReadBuffer(poolSocketBuffer)
	_ = pc.SetWriteBuffer(poolSocketBuffer)

	p := &resolverPool{cfg: cfg, pc: pc, done: make(chan struct{})}
	for _, u := range cfg.upstreams {
		p.conns = append(p.conns, newQueryConn(u, cfg))
	}

	p.wg.Add(1)
	go p.serve()
	return p, nil
}

func newQueryConn(u upstream, cfg poolConfig) queryConn {
	if u.kind.stream() {
		return newStreamConn(u, cfg)
	}
	return newUDPConn(u, cfg)
}

func (p *resolverPool) addr() string { return p.pc.LocalAddr().String() }

func (p *resolverPool) close() {
	p.closeOnce.Do(func() {
		close(p.done)
		_ = p.pc.Close()
		for _, c := range p.conns {
			c.close()
		}
	})
	p.wg.Wait()
}

func (p *resolverPool) serve() {
	defer p.wg.Done()
	buf := make([]byte, maxDNSMessage)
	for {
		n, from, err := p.pc.ReadFromUDP(buf)
		if err != nil {
			return
		}
		if n < 12 {
			continue
		}
		msg := make([]byte, n)
		copy(msg, buf[:n])
		p.wg.Add(1)
		go func() {
			defer contain("resolver pool")
			defer p.wg.Done()
			p.dispatch(msg, from)
		}()
	}
}

func (p *resolverPool) dispatch(msg []byte, from *net.UDPAddr) {
	targets := p.pick()
	if len(targets) == 0 {
		return
	}

	if p.cfg.authoritative {
		msg[2] &^= 0x01
	}

	ctx, cancel := context.WithTimeout(context.Background(), p.cfg.timeout)
	defer cancel()

	replies := make(chan []byte, len(targets))
	for _, c := range targets {
		go func(c queryConn) {
			defer func() {
				if r := recover(); r != nil {
					contained("resolver query", r)
					replies <- nil
				}
			}()
			reply, err := c.query(ctx, msg)
			if err != nil || !usableReply(msg, reply) {
				replies <- nil
				return
			}
			replies <- reply
		}(c)
	}

	for range targets {
		select {
		case reply := <-replies:
			if reply != nil {
				_, _ = p.pc.WriteToUDP(reply, from)
				return
			}
		case <-ctx.Done():
			return
		case <-p.done:
			return
		}
	}
}

func (p *resolverPool) pick() []queryConn {
	if p.cfg.strategy != strategyRotate || p.cfg.perQuery >= len(p.conns) {
		return p.conns
	}
	start := int(p.rr.Add(1)-1) % len(p.conns)
	out := make([]queryConn, 0, p.cfg.perQuery)
	for i := 0; i < p.cfg.perQuery; i++ {
		out = append(out, p.conns[(start+i)%len(p.conns)])
	}
	return out
}

func usableReply(query, reply []byte) bool {
	if len(reply) < 12 || len(query) < 12 {
		return false
	}
	if reply[0] != query[0] || reply[1] != query[1] {
		return false
	}
	if reply[2]&0x80 == 0 {
		return false
	}
	return reply[3]&0x0f == 0
}

func msgID(msg []byte) uint16 { return binary.BigEndian.Uint16(msg[:2]) }

func setMsgID(msg []byte, id uint16) { binary.BigEndian.PutUint16(msg[:2], id) }

func contained(where string, r any) {
	log.Errorf("zeddns: %s panicked: %v\n%s", where, r, debug.Stack())
}
