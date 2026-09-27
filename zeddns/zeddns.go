package zeddns

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"runtime/debug"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/net2share/vaydns/client"
	"github.com/net2share/vaydns/dns"
	"github.com/net2share/vaydns/turbotunnel"
	utls "github.com/refraction-networking/utls"
	log "github.com/sirupsen/logrus"
)

const (
	linkAttempts      = 3
	reconnectMinDelay = 500 * time.Millisecond
	reconnectMaxDelay = 4 * time.Second

	firstReplyTimeout = 20 * time.Second
)

const (
	wireVayDNS = "vaydns"
	wireDnstt  = "dnstt"
)

type config struct {
	Listen            string      `json:"listen"`
	Domain            string      `json:"domain"`
	ServerKey         string      `json:"serverKey"`
	Resolvers         []string    `json:"resolvers"`
	Strategy          string      `json:"strategy"`
	PerQuery          int         `json:"perQuery"`
	Wire              string      `json:"wire"`
	Authoritative     bool        `json:"authoritative"`
	PayloadLimit      int         `json:"payloadLimit"`
	Record            string      `json:"record"`
	NameLimit         int         `json:"nameLimit"`
	LabelLimit        int         `json:"labelLimit"`
	IDBytes           int         `json:"idBytes"`
	QueriesPerSecond  float64     `json:"queriesPerSecond"`
	IdleSeconds       int         `json:"idleSeconds"`
	KeepaliveSeconds  int         `json:"keepaliveSeconds"`
	ResolverTimeoutMs int         `json:"resolverTimeoutMs"`
	TLSHello          string      `json:"tlsHello"`
	ExitAuth          *credential `json:"exitAuth,omitempty"`
	Via               *hop        `json:"via,omitempty"`
	LogLevel          string      `json:"logLevel"`
}

type credential struct {
	User string `json:"user"`
	Pass string `json:"pass"`
}

type hop struct {
	Address string `json:"address"`
	User    string `json:"user"`
	Pass    string `json:"pass"`
}

func parseConfig(document string) (config, error) {
	var cfg config
	dec := json.NewDecoder(bytes.NewReader([]byte(document)))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&cfg); err != nil {
		return cfg, fmt.Errorf("tunnel configuration: %w", err)
	}
	cfg.Domain = strings.TrimSpace(cfg.Domain)
	cfg.ServerKey = strings.TrimSpace(cfg.ServerKey)
	switch {
	case cfg.Listen == "":
		return cfg, errors.New("tunnel configuration: listen address is required")
	case cfg.Domain == "" || cfg.ServerKey == "":
		return cfg, errors.New("domain and server key are required")
	}
	switch cfg.Wire {
	case "":
		cfg.Wire = wireVayDNS
	case wireVayDNS, wireDnstt:
	default:
		return cfg, fmt.Errorf("tunnel configuration: unknown wire format %q", cfg.Wire)
	}
	switch cfg.Strategy {
	case "":
		cfg.Strategy = strategyAll
	case strategyAll, strategyRotate:
	default:
		return cfg, fmt.Errorf("tunnel configuration: unknown resolver strategy %q", cfg.Strategy)
	}
	if cfg.Record == "" {
		cfg.Record = "txt"
	}
	return cfg, nil
}

type Session struct {
	cfg config

	mu       sync.Mutex
	running  bool
	listener net.Listener
	pool     *resolverPool
	cancel   context.CancelFunc
	wg       sync.WaitGroup

	lmu     sync.Mutex
	tunnel  *client.Tunnel
	linkUp  bool
	linkGen uint64

	failure atomic.Value
}

func Open(configJSON string) (session *Session, err error) {
	defer func() {
		if r := recover(); r != nil {
			log.Errorf("zeddns: Open panicked: %v\n%s", r, debug.Stack())
			session, err = nil, fmt.Errorf("zeddns: internal error: %v", r)
		}
	}()
	cfg, err := parseConfig(configJSON)
	if err != nil {
		return nil, err
	}
	if cfg.LogLevel != "" {
		if lvl, err := log.ParseLevel(cfg.LogLevel); err == nil {
			log.SetLevel(lvl)
		}
	}

	s := &Session{cfg: cfg}
	s.failure.Store("")

	tunnel, pool, err := s.build()
	if err != nil {
		return nil, err
	}
	ln, err := net.Listen("tcp", cfg.Listen)
	if err != nil {
		if pool != nil {
			pool.close()
		}
		return nil, fmt.Errorf("listening on %s: %w", cfg.Listen, err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	s.tunnel, s.pool, s.listener, s.cancel, s.running = tunnel, pool, ln, cancel, true

	s.wg.Add(2)
	go s.serve(ctx, ln)
	go s.warmUp(ctx)
	return s, nil
}

func (s *Session) Close() {
	defer contain("close")
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return
	}
	s.running = false
	cancel, ln, pool := s.cancel, s.listener, s.pool
	s.cancel, s.listener, s.pool = nil, nil, nil
	s.mu.Unlock()

	cancel()
	_ = ln.Close()

	s.lmu.Lock()
	s.teardownLink()
	s.tunnel = nil
	s.lmu.Unlock()

	if pool != nil {
		pool.close()
	}
	s.wg.Wait()
}

func (s *Session) Active() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.running
}

func (s *Session) Failure() string {
	v, _ := s.failure.Load().(string)
	return v
}

func (s *Session) build() (*client.Tunnel, *resolverPool, error) {
	cfg := s.cfg
	ups, err := parseResolvers(cfg.Resolvers)
	if err != nil {
		return nil, nil, err
	}

	var dial dialFunc
	if cfg.Via != nil && cfg.Via.Address != "" {
		dial = socks5Dialer{proxy: cfg.Via.Address, username: cfg.Via.User, password: cfg.Via.Pass}.DialContext
		for i := range ups {
			if ups[i].kind == transportUDP {
				ups[i].kind = transportTCP
			}
		}
	}

	var hello *utls.ClientHelloID
	if spec := strings.TrimSpace(cfg.TLSHello); spec != "" {
		if id, err := client.SampleUTLSDistribution(spec); err == nil {
			hello = id
		} else {
			log.Warnf("zeddns: ignoring TLS hello %q: %v", spec, err)
		}
	}

	server, err := s.tunnelServer()
	if err != nil {
		return nil, nil, err
	}
	resolver, pool, err := s.resolver(ups, hello, dial)
	if err != nil {
		return nil, nil, err
	}
	tunnel, err := client.NewTunnel(resolver, server)
	if err != nil {
		if pool != nil {
			pool.close()
		}
		return nil, nil, err
	}
	tunnel.IdleTimeout = time.Duration(cfg.IdleSeconds) * time.Second
	tunnel.KeepAlive = time.Duration(cfg.KeepaliveSeconds) * time.Second
	return tunnel, pool, nil
}

func (s *Session) resolver(ups []upstream, hello *utls.ClientHelloID, dial dialFunc) (client.Resolver, *resolverPool, error) {
	cfg := s.cfg
	timeout := time.Duration(cfg.ResolverTimeoutMs) * time.Millisecond
	native := func(kind client.ResolverType, addr string) (client.Resolver, error) {
		r, err := client.NewResolver(kind, addr)
		if err != nil {
			return r, err
		}
		r.UTLSClientHelloID = hello
		r.UDPTimeout = timeout
		return r, nil
	}

	if ups[0].kind == transportDoH {
		if len(ups) > 1 {
			log.Warnf("zeddns: using the first DoH endpoint, ignoring %d more", len(ups)-1)
		}
		r, err := native(client.ResolverTypeDOH, ups[0].addr)
		if err != nil {
			return r, nil, err
		}
		if dial != nil {
			r.RoundTripper = newProxiedDoHTransport(dial)
		}
		return r, nil, nil
	}

	if len(ups) == 1 && ups[0].kind == transportUDP && dial == nil && !cfg.Authoritative {
		r, err := native(client.ResolverTypeUDP, ups[0].addr)
		return r, nil, err
	}

	pool, err := newResolverPool(poolConfig{
		upstreams:     ups,
		strategy:      cfg.Strategy,
		perQuery:      cfg.PerQuery,
		timeout:       timeout,
		helloID:       hello,
		dial:          dial,
		authoritative: cfg.Authoritative,
	})
	if err != nil {
		return client.Resolver{}, nil, err
	}
	r, err := native(client.ResolverTypeUDP, pool.addr())
	if err != nil {
		pool.close()
		return r, nil, err
	}
	return r, pool, nil
}

func (s *Session) tunnelServer() (client.TunnelServer, error) {
	cfg := s.cfg
	server, err := client.NewTunnelServer(cfg.Domain, cfg.ServerKey)
	if err != nil {
		return server, err
	}
	compat := cfg.Wire == wireDnstt
	server.DnsttCompat = compat
	server.ClientIDSize = cfg.IDBytes
	server.MaxNumLabels = cfg.LabelLimit
	server.RecordType = cfg.Record
	server.RPS = cfg.QueriesPerSecond

	switch {
	case cfg.NameLimit > 0:
		server.MaxQnameLen = cfg.NameLimit
	case cfg.PayloadLimit > 0:
		server.MaxQnameLen = nameLimitForPayload(server.Addr, cfg.PayloadLimit, cfg.LabelLimit, cfg.IDBytes, compat)
	}
	return server, nil
}

func nameLimitForPayload(domain dns.Name, payload, labelLimit, idBytes int, compat bool) int {
	wire := turbotunnel.WireConfig{ClientIDSize: idBytes, Compat: compat}
	if compat {
		wire.ClientIDSize = 8
	} else if wire.ClientIDSize <= 0 {
		wire.ClientIDSize = 2
	}
	overhead := wire.DataOverhead()

	best := 0
	for length := 1; length <= 253; length++ {
		if client.DNSNameCapacity(domain, length, labelLimit)-overhead > payload {
			break
		}
		best = length
	}
	return best
}

func (s *Session) serve(ctx context.Context, ln net.Listener) {
	defer s.wg.Done()
	for {
		local, err := ln.Accept()
		if err != nil {
			return
		}
		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			s.handle(ctx, local)
		}()
	}
}

func (s *Session) warmUp(ctx context.Context) {
	defer contain("first link")
	defer s.wg.Done()
	s.lmu.Lock()
	defer s.lmu.Unlock()
	if s.tunnel == nil || s.linkUp {
		return
	}
	if err := s.buildLink(ctx); err != nil && ctx.Err() == nil {
		log.Warnf("zeddns: first link failed: %v", err)
	}
}

func (s *Session) handle(ctx context.Context, local net.Conn) {
	defer contain("connection")
	defer local.Close()

	remote, gen, err := s.stream(ctx)
	if err != nil {
		if ctx.Err() == nil {
			s.fail(err)
		}
		return
	}
	defer remote.Close()
	watch := &replyWatch{conn: remote, timeout: firstReplyTimeout}

	if auth := s.cfg.ExitAuth; auth != nil && auth.User != "" {
		relay := credentialRelay{username: auth.User, password: auth.Pass}
		_ = remote.SetDeadline(time.Now().Add(firstReplyTimeout))
		err := relay.negotiate(local, remote)
		_ = remote.SetDeadline(time.Time{})
		if err != nil {
			if isTimeout(err) {
				s.dropLink(gen, errors.New("the tunnel exit stopped answering"))
				return
			}
			s.fail(fmt.Errorf("SOCKS5 authentication at the tunnel exit: %w", err))
			return
		}
		watch.replied()
	}

	var stalled atomic.Bool
	done := make(chan struct{}, 2)
	go func() {
		buf := make([]byte, 32<<10)
		for {
			n, err := local.Read(buf)
			if n > 0 {
				watch.sent()
				if _, werr := remote.Write(buf[:n]); werr != nil {
					break
				}
			}
			if err != nil {
				break
			}
		}
		done <- struct{}{}
	}()
	go func() {
		buf := make([]byte, 32<<10)
		for {
			n, err := remote.Read(buf)
			if n > 0 {
				watch.replied()
				if _, werr := local.Write(buf[:n]); werr != nil {
					break
				}
			}
			if err != nil {
				if watch.stalled(err) {
					stalled.Store(true)
				}
				break
			}
		}
		done <- struct{}{}
	}()

	select {
	case <-done:
	case <-ctx.Done():
	}
	if stalled.Load() {
		s.dropLink(gen, errors.New("the tunnel stopped answering"))
	}
}

type replyWatch struct {
	mu      sync.Mutex
	conn    net.Conn
	timeout time.Duration
	armed   bool
	heard   bool
}

func (w *replyWatch) sent() {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.heard || w.armed {
		return
	}
	w.armed = true
	_ = w.conn.SetReadDeadline(time.Now().Add(w.timeout))
}

func (w *replyWatch) replied() {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.heard {
		return
	}
	w.heard = true
	if w.armed {
		_ = w.conn.SetReadDeadline(time.Time{})
	}
}

func (w *replyWatch) stalled(err error) bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.armed && !w.heard && isTimeout(err)
}

func isTimeout(err error) bool {
	var ne net.Error
	return errors.As(err, &ne) && ne.Timeout()
}

func (s *Session) dropLink(gen uint64, reason error) {
	s.lmu.Lock()
	if s.linkUp && s.linkGen == gen {
		s.teardownLink()
	}
	s.lmu.Unlock()
	s.fail(reason)
}

func (s *Session) stream(ctx context.Context) (net.Conn, uint64, error) {
	var lastErr error
	for attempt := 0; attempt < 2; attempt++ {
		var (
			stream net.Conn
			gen    uint64
			err    error
			final  bool
		)
		func() {
			s.lmu.Lock()
			defer s.lmu.Unlock()
			if s.tunnel == nil {
				err, final = errors.New("tunnel is closed"), true
				return
			}
			if !s.linkUp {
				if err = s.buildLink(ctx); err != nil {
					final = true
					return
				}
			}
			gen = s.linkGen
			stream, err = s.tunnel.OpenStream()
		}()
		if err == nil {
			return stream, gen, nil
		}
		if final {
			return nil, 0, err
		}
		lastErr = err

		func() {
			s.lmu.Lock()
			defer s.lmu.Unlock()
			if s.linkGen == gen {
				s.teardownLink()
			}
		}()
	}
	return nil, 0, lastErr
}

func (s *Session) buildLink(ctx context.Context) error {
	delay := reconnectMinDelay
	var err error
	for attempt := 0; attempt < linkAttempts; attempt++ {
		if err = s.dialLink(); err == nil {
			s.linkUp = true
			s.linkGen++
			return nil
		}
		_ = s.tunnel.Close()
		if attempt+1 == linkAttempts {
			break
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}
		delay *= 2
		if delay > reconnectMaxDelay {
			delay = reconnectMaxDelay
		}
	}
	s.fail(err)
	return err
}

func (s *Session) dialLink() error {
	t := s.tunnel
	if err := t.InitiateResolverConnection(); err != nil {
		return fmt.Errorf("resolver connection: %w", err)
	}
	if err := t.InitiateDNSPacketConn(t.TunnelServer.Addr); err != nil {
		return fmt.Errorf("DNS packet connection: %w", err)
	}
	if err := t.InitiateKCPConn(0); err != nil {
		return fmt.Errorf("tunnel connection: %w", err)
	}
	if err := t.InitiateNoiseChannel(); err != nil {
		return fmt.Errorf("handshake with the server: %w", err)
	}
	if err := t.InitiateSmuxSession(); err != nil {
		return fmt.Errorf("multiplexed session: %w", err)
	}
	return nil
}

func (s *Session) teardownLink() {
	if s.tunnel != nil {
		_ = s.tunnel.Close()
	}
	s.linkUp = false
}

func (s *Session) fail(err error) {
	if err == nil {
		return
	}
	s.failure.Store(err.Error())
	log.Warnf("zeddns: %v", err)
}

func contain(where string) {
	if r := recover(); r != nil {
		log.Errorf("zeddns: %s panicked: %v\n%s", where, r, debug.Stack())
	}
}
