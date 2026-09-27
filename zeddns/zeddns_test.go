package zeddns

import (
	"context"
	"encoding/binary"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type fakeResolver struct {
	marker byte
	rcode  byte
	calls  atomic.Int64
}

func reply(q []byte, rcode, marker byte) []byte {
	r := make([]byte, len(q)+1)
	copy(r, q)
	r[2] |= 0x80
	r[3] = (r[3] &^ 0x0f) | rcode
	r[len(q)] = marker
	return r
}

func (f *fakeResolver) serveUDP(t *testing.T) string {
	t.Helper()
	pc, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { pc.Close() })
	go func() {
		buf := make([]byte, 4096)
		for {
			n, from, err := pc.ReadFrom(buf)
			if err != nil {
				return
			}
			f.calls.Add(1)
			_, _ = pc.WriteTo(reply(buf[:n], f.rcode, f.marker), from)
		}
	}()
	return pc.LocalAddr().String()
}

func (f *fakeResolver) serveTCP(t *testing.T, delay func(q []byte) time.Duration) string {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { ln.Close() })
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go func() {
				defer conn.Close()
				var wmu sync.Mutex
				var hdr [2]byte
				for {
					if _, err := io.ReadFull(conn, hdr[:]); err != nil {
						return
					}
					q := make([]byte, binary.BigEndian.Uint16(hdr[:]))
					if _, err := io.ReadFull(conn, q); err != nil {
						return
					}
					f.calls.Add(1)
					go func(q []byte) {
						if delay != nil {
							time.Sleep(delay(q))
						}
						r := reply(q, f.rcode, f.marker)
						frame := make([]byte, 2+len(r))
						binary.BigEndian.PutUint16(frame[:2], uint16(len(r)))
						copy(frame[2:], r)
						wmu.Lock()
						_, _ = conn.Write(frame)
						wmu.Unlock()
					}(q)
				}
			}()
		}
	}()
	return ln.Addr().String()
}

func query(id uint16, tag byte) []byte {
	q := make([]byte, 20)
	binary.BigEndian.PutUint16(q[:2], id)
	q[2] = 0x01
	q[5] = 0x01
	q[19] = tag
	return q
}

func ask(t *testing.T, p *resolverPool, q []byte) []byte {
	t.Helper()
	conn, err := net.Dial("udp", p.addr())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	if err := conn.SetDeadline(time.Now().Add(5 * time.Second)); err != nil {
		t.Fatal(err)
	}
	if _, err := conn.Write(q); err != nil {
		t.Fatal(err)
	}
	buf := make([]byte, 4096)
	n, err := conn.Read(buf)
	if err != nil {
		t.Fatalf("no answer: %v", err)
	}
	return buf[:n]
}

func TestParseUpstreams(t *testing.T) {
	got, err := parseResolvers([]string{"1.1.1.1", " tcp://8.8.8.8:5353", "tls://9.9.9.9", "https://dns.example/dns-query"})
	if err != nil {
		t.Fatal(err)
	}
	want := []upstream{
		{transportUDP, "1.1.1.1:53"},
		{transportTCP, "8.8.8.8:5353"},
		{transportDoT, "9.9.9.9:853"},
		{transportDoH, "https://dns.example/dns-query"},
	}
	if len(got) != len(want) {
		t.Fatalf("got %d resolvers, want %d: %v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("resolver %d = %v, want %v", i, got[i], want[i])
		}
	}
	if _, err := parseResolvers([]string{"  ", ""}); err == nil {
		t.Error("an empty list should be rejected")
	}
}

func TestParseUpstreamsIPv6(t *testing.T) {
	got, err := parseResolvers([]string{"2001:4860:4860::8888", "[2606:4700:4700::1111]:5353"})
	if err != nil {
		t.Fatal(err)
	}
	if got[0].addr != "[2001:4860:4860::8888]:53" {
		t.Errorf("bare IPv6 = %q", got[0].addr)
	}
	if got[1].addr != "[2606:4700:4700::1111]:5353" {
		t.Errorf("bracketed IPv6 with port = %q", got[1].addr)
	}
}

func TestPoolSkipsRefusedAnswers(t *testing.T) {
	bad := &fakeResolver{marker: 'B', rcode: 5}
	good := &fakeResolver{marker: 'G', rcode: 0}
	p, err := newResolverPool(poolConfig{
		upstreams: []upstream{{transportUDP, bad.serveUDP(t)}, {transportUDP, good.serveUDP(t)}},
		strategy:  strategyAll,
		timeout:   2 * time.Second,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer p.close()

	q := query(0x1234, 'q')
	got := ask(t, p, q)
	if got[len(got)-1] != 'G' {
		t.Fatalf("answer came from %q, want the resolver that said NOERROR", got[len(got)-1])
	}
	if msgID(got) != 0x1234 {
		t.Errorf("transaction id = %#x, want 0x1234", msgID(got))
	}
}

func TestPoolRoundRobinSpread(t *testing.T) {
	a := &fakeResolver{marker: 'A'}
	b := &fakeResolver{marker: 'B'}
	c := &fakeResolver{marker: 'C'}
	p, err := newResolverPool(poolConfig{
		upstreams: []upstream{{transportUDP, a.serveUDP(t)}, {transportUDP, b.serveUDP(t)}, {transportUDP, c.serveUDP(t)}},
		strategy:  strategyRotate,
		perQuery:  1,
		timeout:   2 * time.Second,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer p.close()

	for i := 0; i < 9; i++ {
		ask(t, p, query(uint16(i), 'q'))
	}
	for name, r := range map[string]*fakeResolver{"A": a, "B": b, "C": c} {
		if n := r.calls.Load(); n != 3 {
			t.Errorf("resolver %s saw %d queries, want 3", name, n)
		}
	}
}

func TestPoolFanoutAsksEveryResolver(t *testing.T) {
	a := &fakeResolver{marker: 'A'}
	b := &fakeResolver{marker: 'B'}
	p, err := newResolverPool(poolConfig{
		upstreams: []upstream{{transportUDP, a.serveUDP(t)}, {transportUDP, b.serveUDP(t)}},
		strategy:  strategyAll,
		timeout:   2 * time.Second,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer p.close()

	ask(t, p, query(7, 'q'))
	time.Sleep(200 * time.Millisecond)
	if a.calls.Load() != 1 || b.calls.Load() != 1 {
		t.Errorf("fanout reached %d and %d resolvers, want both", a.calls.Load(), b.calls.Load())
	}
}

func TestStreamConnSeparatesCollidingIDs(t *testing.T) {
	srv := &fakeResolver{marker: 'T'}
	addr := srv.serveTCP(t, func(q []byte) time.Duration {
		if q[len(q)-1] == '1' {
			return 300 * time.Millisecond
		}
		return 0
	})

	c := newStreamConn(upstream{transportTCP, addr}, poolConfig{
		timeout: 5 * time.Second,
		dial:    (&net.Dialer{}).DialContext,
	})
	defer c.close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var wg sync.WaitGroup
	results := make([]byte, 2)
	for i, tag := range []byte{'1', '2'} {
		wg.Add(1)
		go func(i int, tag byte) {
			defer wg.Done()
			r, err := c.query(ctx, query(0x4242, tag))
			if err != nil {
				t.Errorf("query %c: %v", tag, err)
				return
			}
			if msgID(r) != 0x4242 {
				t.Errorf("query %c got id %#x back, want the caller's 0x4242", tag, msgID(r))
			}
			results[i] = r[len(r)-2]
		}(i, tag)
	}
	wg.Wait()

	if results[0] != '1' || results[1] != '2' {
		t.Fatalf("replies were crossed: got %q and %q, want '1' and '2'", results[0], results[1])
	}
}

func TestStreamConnReusesOneConnection(t *testing.T) {
	srv := &fakeResolver{marker: 'T'}
	var conns atomic.Int64
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			conns.Add(1)
			go func() {
				defer conn.Close()
				var hdr [2]byte
				for {
					if _, err := io.ReadFull(conn, hdr[:]); err != nil {
						return
					}
					q := make([]byte, binary.BigEndian.Uint16(hdr[:]))
					if _, err := io.ReadFull(conn, q); err != nil {
						return
					}
					r := reply(q, srv.rcode, srv.marker)
					frame := make([]byte, 2+len(r))
					binary.BigEndian.PutUint16(frame[:2], uint16(len(r)))
					copy(frame[2:], r)
					if _, err := conn.Write(frame); err != nil {
						return
					}
				}
			}()
		}
	}()

	c := newStreamConn(upstream{transportTCP, ln.Addr().String()}, poolConfig{
		timeout: 5 * time.Second,
		dial:    (&net.Dialer{}).DialContext,
	})
	defer c.close()

	ctx := context.Background()
	for i := 0; i < 5; i++ {
		if _, err := c.query(ctx, query(uint16(i), 'q')); err != nil {
			t.Fatalf("query %d: %v", i, err)
		}
	}
	if n := conns.Load(); n != 1 {
		t.Errorf("opened %d connections for 5 queries, want 1", n)
	}
}

func TestCredentialRelay(t *testing.T) {
	localApp, localRelay := net.Pipe()
	relayServer, farEnd := net.Pipe()
	defer localApp.Close()
	defer farEnd.Close()

	done := make(chan error, 1)
	go func() {
		done <- credentialRelay{username: "u", password: "p"}.negotiate(localRelay, relayServer)
	}()

	go func() {
		_, _ = localApp.Write([]byte{socksVersion, 0x01, socksAuthNone})
		var choice [2]byte
		_, _ = io.ReadFull(localApp, choice[:])
		if choice[1] != socksAuthNone {
			t.Errorf("relay told the local side to authenticate (%#x)", choice[1])
		}
	}()

	var greeting [3]byte
	if _, err := io.ReadFull(farEnd, greeting[:]); err != nil {
		t.Fatal(err)
	}
	if greeting[2] != socksAuthUser {
		t.Fatalf("relay offered method %#x upstream, want username/password", greeting[2])
	}
	if _, err := farEnd.Write([]byte{socksVersion, socksAuthUser}); err != nil {
		t.Fatal(err)
	}

	var head [2]byte
	if _, err := io.ReadFull(farEnd, head[:]); err != nil {
		t.Fatal(err)
	}
	user := make([]byte, head[1])
	if _, err := io.ReadFull(farEnd, user); err != nil {
		t.Fatal(err)
	}
	var plen [1]byte
	if _, err := io.ReadFull(farEnd, plen[:]); err != nil {
		t.Fatal(err)
	}
	pass := make([]byte, plen[0])
	if _, err := io.ReadFull(farEnd, pass); err != nil {
		t.Fatal(err)
	}
	if string(user) != "u" || string(pass) != "p" {
		t.Errorf("relay sent %q/%q upstream", user, pass)
	}
	if _, err := farEnd.Write([]byte{socksUserAuthVersion, 0x00}); err != nil {
		t.Fatal(err)
	}

	if err := <-done; err != nil {
		t.Fatalf("negotiate: %v", err)
	}
}

func TestNameLimitShrinksWithPayload(t *testing.T) {
	server, err := client_NewTunnelServerForTest("t.example.com")
	if err != nil {
		t.Fatal(err)
	}
	full := nameLimitForPayload(server, 0xFFFF, 0, 2, false)
	small := nameLimitForPayload(server, 60, 0, 2, false)
	if full != 253 {
		t.Errorf("an unbounded payload should allow the longest name, got %d", full)
	}
	if small <= 0 || small >= full {
		t.Errorf("a 60-byte payload gave name length %d, want between 1 and %d", small, full)
	}
}

func TestParseConfig(t *testing.T) {
	cfg, err := parseConfig(`{"listen":"127.0.0.1:1","domain":" t.example.com ","serverKey":"ab","resolvers":["1.1.1.1"]}`)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Wire != wireVayDNS || cfg.Strategy != strategyAll || cfg.Record != "txt" || cfg.Domain != "t.example.com" {
		t.Errorf("defaults not applied: %+v", cfg)
	}

	bad := map[string]string{
		"unknown field":    `{"listen":"a","domain":"d","serverKey":"k","resolverz":[]}`,
		"no listen":        `{"domain":"d","serverKey":"k"}`,
		"no server key":    `{"listen":"a","domain":"d"}`,
		"unknown wire":     `{"listen":"a","domain":"d","serverKey":"k","wire":"iodine"}`,
		"unknown strategy": `{"listen":"a","domain":"d","serverKey":"k","strategy":"random"}`,
		"not json":         `listen=a`,
	}
	for name, doc := range bad {
		if _, err := parseConfig(doc); err == nil {
			t.Errorf("%s: accepted %s", name, doc)
		}
	}
}
