package zeddns

import (
	"fmt"
	"net"
	"strings"
)

type transport int

const (
	transportUDP transport = iota
	transportTCP
	transportDoT
	transportDoH
)

func (t transport) String() string {
	switch t {
	case transportTCP:
		return "tcp"
	case transportDoT:
		return "dot"
	case transportDoH:
		return "doh"
	default:
		return "udp"
	}
}

func (t transport) stream() bool { return t == transportTCP || t == transportDoT }

type upstream struct {
	kind transport
	addr string
}

func (u upstream) String() string { return u.kind.String() + "://" + u.addr }

func parseResolvers(entries []string) ([]upstream, error) {
	var out []upstream
	for _, raw := range entries {
		entry := strings.TrimSpace(raw)
		if entry == "" {
			continue
		}
		var u upstream
		switch {
		case strings.HasPrefix(entry, "https://"), strings.HasPrefix(entry, "http://"):
			u = upstream{kind: transportDoH, addr: entry}
		case strings.HasPrefix(entry, "tls://"):
			u = upstream{kind: transportDoT, addr: withDefaultPort(entry[len("tls://"):], "853")}
		case strings.HasPrefix(entry, "tcp://"):
			u = upstream{kind: transportTCP, addr: withDefaultPort(entry[len("tcp://"):], "53")}
		case strings.HasPrefix(entry, "udp://"):
			u = upstream{kind: transportUDP, addr: withDefaultPort(entry[len("udp://"):], "53")}
		default:
			u = upstream{kind: transportUDP, addr: withDefaultPort(entry, "53")}
		}
		if u.addr == "" {
			return nil, fmt.Errorf("empty resolver in %q", entry)
		}
		out = append(out, u)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("no resolvers configured")
	}
	return out, nil
}

func withDefaultPort(hostPort, port string) string {
	hostPort = strings.TrimSpace(strings.TrimSuffix(hostPort, "/"))
	if hostPort == "" {
		return ""
	}
	if _, _, err := net.SplitHostPort(hostPort); err == nil {
		return hostPort
	}
	if strings.HasPrefix(hostPort, "[") && strings.HasSuffix(hostPort, "]") {
		return hostPort + ":" + port
	}
	if strings.Count(hostPort, ":") > 1 {
		return "[" + hostPort + "]:" + port
	}
	return hostPort + ":" + port
}

func hostOf(hostPort string) string {
	if h, _, err := net.SplitHostPort(hostPort); err == nil {
		return h
	}
	return hostPort
}
