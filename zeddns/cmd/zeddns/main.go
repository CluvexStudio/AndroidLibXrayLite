package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"zeddns"
)

func main() {
	var (
		resolvers     = flag.String("dns-addr", "", "resolvers: host:port, or tcp://, tls://, https:// entries, comma separated")
		domain        = flag.String("domain", "", "tunnel domain configured on the server")
		serverKey     = flag.String("pubkey", "", "server public key, hex encoded")
		listen        = flag.String("listen", "127.0.0.1:7000", "local address to serve")
		strategy      = flag.String("resolver-mode", "all", "resolver strategy: all or rotate (fanout and roundrobin are accepted too)")
		perQuery      = flag.Int("rr-spread", 1, "resolvers per query when rotating")
		dnsttWire     = flag.Bool("dnstt-compat", false, "speak the original dnstt wire format")
		authoritative = flag.Bool("authoritative", false, "the resolvers are the domain's own name servers")
		payloadLimit  = flag.Int("max-payload", 0, "cap the data one query carries, in bytes (0 = no cap)")
		record        = flag.String("record-type", "txt", "record carrying downstream data")
		nameLimit     = flag.Int("max-qname-len", 0, "cap the query name length in bytes (0 = engine default)")
		labelLimit    = flag.Int("max-num-labels", 0, "cap the number of data labels (0 = no cap)")
		idBytes       = flag.Int("clientid-size", 0, "client id width in bytes (0 = engine default)")
		rps           = flag.Float64("rps", 0, "limit queries per second (0 = unlimited)")
		idleSeconds   = flag.Int("idle-timeout", 0, "link idle timeout in seconds (0 = engine default)")
		keepalive     = flag.Int("keepalive", 0, "keepalive interval in seconds (0 = engine default)")
		resolverMs    = flag.Int("udp-timeout", 0, "per-resolver response timeout in milliseconds")
		tlsHello      = flag.String("utls", "", "TLS hello distribution for DoT and DoH, e.g. 3*Firefox,2*Chrome")
		exitUser      = flag.String("socks-user", "", "username the SOCKS5 server at the tunnel exit expects")
		exitPass      = flag.String("socks-pass", "", "password the SOCKS5 server at the tunnel exit expects")
		viaAddr       = flag.String("proxy", "", "carry this tunnel's own resolver traffic through a local SOCKS5 proxy")
		viaUser       = flag.String("proxy-user", "", "username for -proxy")
		viaPass       = flag.String("proxy-pass", "", "password for -proxy")
		logLevel      = flag.String("log-level", "info", "debug, info, warning or error")
	)
	flag.Parse()

	wire := "vaydns"
	if *dnsttWire {
		wire = "dnstt"
	}
	cfg := map[string]any{
		"listen":            *listen,
		"domain":            *domain,
		"serverKey":         *serverKey,
		"resolvers":         strings.Split(*resolvers, ","),
		"strategy":          strategyName(*strategy),
		"perQuery":          *perQuery,
		"wire":              wire,
		"authoritative":     *authoritative,
		"payloadLimit":      *payloadLimit,
		"record":            *record,
		"nameLimit":         *nameLimit,
		"labelLimit":        *labelLimit,
		"idBytes":           *idBytes,
		"queriesPerSecond":  *rps,
		"idleSeconds":       *idleSeconds,
		"keepaliveSeconds":  *keepalive,
		"resolverTimeoutMs": *resolverMs,
		"tlsHello":          *tlsHello,
		"logLevel":          *logLevel,
	}
	if *exitUser != "" {
		cfg["exitAuth"] = map[string]string{"user": *exitUser, "pass": *exitPass}
	}
	if *viaAddr != "" {
		cfg["via"] = map[string]string{"address": *viaAddr, "user": *viaUser, "pass": *viaPass}
	}
	document, err := json.Marshal(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "zeddns: %v\n", err)
		os.Exit(2)
	}

	session, err := zeddns.Open(string(document))
	if err != nil {
		fmt.Fprintf(os.Stderr, "zeddns: %v\n", err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "zeddns: serving %s for %s\n", *listen, *domain)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	session.Close()
}

func strategyName(s string) string {
	switch s {
	case "rotate", "roundrobin":
		return "rotate"
	default:
		return "all"
	}
}
