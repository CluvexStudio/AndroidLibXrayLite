module github.com/2dust/AndroidLibXrayLite

go 1.26.0

// Point at our patched fork instead of the upstream module. The only diff is
// infra/conf/xray.go: validateOutboundTransportSecurity is disabled, mirroring
// XTLS/Xray-core@f8d54a0 ("Allow unencrypted outbounds on public Internet for
// VLESS and Trojan"), which had not shipped in a tagged release yet when this
// build was pinned. See vendor/xray-core's git history for the exact diff.
replace github.com/xtls/xray-core => ../xray-core

require (
	github.com/apernet/quic-go v0.61.1-0.20260806010916-184d081eef3e
	github.com/xtls/xray-core v1.260327.1-0.20260728075948-5ca6f4b7d4dc
	golang.org/x/mobile v0.0.0-20260709172247-6129f5bee9d5
)

// gomobile binds these as separate roots; they are NOT imported by this module's own Go code, so
// `go mod tidy` prunes them and the bind then fails with "no exported names". Keep them here
// explicitly and do NOT run `go mod tidy` on this module.
require (
	github.com/Psiphon-Labs/psiphon-tunnel-core v1.0.11-0.20250319154633-ceb78316d06e // indirect
	// sing-box: bound as its own root (experimental/libbox, the full engine); Xray also links it for
	// the "singbox" outbound.
	github.com/sagernet/sing-box v1.14.0
	masterdnsvpn-go v0.0.0
	noizdns v0.0.0
	vaydns-mobile v0.0.0
)

require (
	github.com/amnezia-vpn/amneziawg-go/v3 v3.0.0-00010101000000-000000000000 // indirect
	github.com/andybalholm/brotli v1.2.0 // indirect
	github.com/caddyserver/certmagic v0.25.3-0.20260421143802-60d9d8b415d6 // indirect
	github.com/caddyserver/zerossl v0.1.5 // indirect
	github.com/cloudflare/circl v1.6.5 // indirect
	github.com/database64128/netx-go v0.1.1 // indirect
	github.com/database64128/tfo-go/v2 v2.3.2 // indirect
	github.com/florianl/go-nfqueue/v2 v2.1.0 // indirect
	github.com/fsnotify/fsnotify v1.9.0 // indirect
	github.com/ghodss/yaml v1.0.1-0.20220118164431-d8423dcdf344 // indirect
	github.com/go-chi/chi/v5 v5.2.5 // indirect
	github.com/go-ole/go-ole v1.3.0 // indirect
	github.com/godbus/dbus/v5 v5.2.2 // indirect
	github.com/gofrs/uuid/v5 v5.5.1 // indirect
	github.com/google/btree v1.1.3 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/google/gopacket v1.1.19 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/gorilla/websocket v1.5.3 // indirect
	github.com/hashicorp/yamux v0.1.2 // indirect
	github.com/huin/goupnp v1.3.0 // indirect
	github.com/jackpal/go-nat-pmp v1.0.2 // indirect
	github.com/jsimonetti/rtnetlink v1.4.1 // indirect
	github.com/juju/ratelimit v1.0.2 // indirect
	github.com/klauspost/compress v1.19.1 // indirect
	github.com/klauspost/cpuid/v2 v2.4.0 // indirect
	github.com/koron/go-ssdp v0.0.4 // indirect
	github.com/libdns/acmedns v0.5.0 // indirect
	github.com/libdns/alidns v1.0.6 // indirect
	github.com/libdns/cloudflare v0.2.2 // indirect
	github.com/libdns/libdns v1.1.1 // indirect
	github.com/libp2p/go-nat v1.0.1-0.20250821073202-01afc089f138 // indirect
	github.com/libp2p/go-netroute v0.2.1 // indirect
	github.com/logrusorgru/aurora v2.0.3+incompatible // indirect
	github.com/mdlayher/netlink v1.11.2 // indirect
	github.com/mdlayher/socket v0.6.0 // indirect
	github.com/metacubex/utls v1.8.7 // indirect
	github.com/mholt/acmez/v3 v3.1.6 // indirect
	github.com/miekg/dns v1.1.73 // indirect
	github.com/pelletier/go-toml v1.9.5 // indirect
	github.com/pion/dtls/v3 v3.1.5 // indirect
	github.com/pion/logging v0.2.4 // indirect
	github.com/pion/stun/v3 v3.1.7 // indirect
	github.com/pion/transport/v4 v4.1.0 // indirect
	github.com/pires/go-proxyproto v0.15.0 // indirect
	github.com/quic-go/qpack v0.6.0 // indirect
	github.com/refraction-networking/utls v1.8.3-0.20260301010127-aa6edf4b11af // indirect
	github.com/robfig/cron/v3 v3.0.1 // indirect
	github.com/rogpeppe/go-internal v1.14.1 // indirect
	github.com/sagernet/bbolt v0.0.0-20260823094646-e24805439c9c // indirect
	github.com/sagernet/fswatch v0.1.2 // indirect
	github.com/sagernet/gvisor v0.0.0-20260727.0-sing-box-mod.1 // indirect
	github.com/sagernet/netlink v0.0.0-20260814022025-64455d367bbf // indirect
	github.com/sagernet/nftables v0.3.0-mod.4 // indirect
	github.com/sagernet/quic-go v0.61.0-sing-box-mod.7 // indirect
	github.com/sagernet/sing v0.9.0-beta.4 // indirect
	github.com/sagernet/sing-mux v0.3.5 // indirect
	github.com/sagernet/sing-shadowsocks v0.2.9 // indirect
	github.com/sagernet/sing-tun v0.9.0-beta.4 // indirect
	github.com/sagernet/sing-usbip v0.0.0-20260817040617-28bd42667eca // indirect
	github.com/sagernet/sing-vmess v0.2.8-0.20250909125414-3aed155119a1 // indirect
	github.com/sagernet/smux v1.5.50-sing-box-mod.1 // indirect
	github.com/vishvananda/netlink v1.3.1 // indirect
	github.com/vishvananda/netns v0.0.5 // indirect
	github.com/wlynxg/anet v0.0.5 // indirect
	github.com/xtls/reality v0.0.0-20260322125925-9234c772ba8f // indirect
	github.com/zeebo/blake3 v0.2.4 // indirect
	go.uber.org/mock v0.6.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.27.1 // indirect
	go.uber.org/zap/exp v0.3.0 // indirect
	go4.org/netipx v0.0.0-20231129151722-fdeea329fbba // indirect
	golang.org/x/crypto v0.55.0 // indirect
	golang.org/x/exp v0.0.0-20260410095643-746e56fc9e2f // indirect
	golang.org/x/mod v0.38.0 // indirect
	golang.org/x/net v0.58.0 // indirect
	golang.org/x/sync v0.22.0 // indirect
	golang.org/x/sys v0.47.0 // indirect
	golang.org/x/text v0.41.0 // indirect
	golang.org/x/time v0.15.0 // indirect
	golang.zx2c4.com/wintun v0.0.0-20230126152724-0fa3db229ce2 // indirect
	golang.zx2c4.com/wireguard v0.0.0-20250521234502-f333402bd9cb // indirect
	golang.zx2c4.com/wireguard/windows v1.0.1 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260526163538-3dc84a4a5aaa // indirect
	google.golang.org/grpc v1.83.1 // indirect
	google.golang.org/protobuf v1.36.12 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gvisor.dev/gvisor v0.0.0-20260122175437-89a5d21be8f0 // indirect
	lukechampine.com/blake3 v1.4.1 // indirect
)

replace github.com/Psiphon-Labs/psiphon-tunnel-core => ../../reference/psiphon-tunnel-core

replace github.com/pion/dtls/v2 => ../../reference/psiphon-tunnel-core/replace/dtls

replace github.com/quic-go/qpack => ../qpack-patched

// DNS-tunnel local sources + the kcp fork they require.
replace noizdns => ../../reference/slipnet-src/noizdns

replace vaydns-mobile => ../../reference/slipnet-src/vaydns-mobile

// MasterDnsVPN DNS tunnel (masterking32/MasterDnsVPN), bound via its mobile/ wrapper.
replace masterdnsvpn-go => ../../reference/MasterDnsVPN

replace www.bamsoftware.com/git/dnstt.git => ../../reference/slipnet-src/dnstt

replace github.com/net2share/vaydns => ../../reference/slipnet-src/vaydns

replace github.com/xtaci/kcp-go/v5 => github.com/net2share/kcp-go/v5 v5.0.0-20260325165956-416ba9d3856d

// AmneziaWG. amneziawg-go is a fork of wireguard-go whose conn.Bind and tun.Device interfaces and
// device.NewDevice signature are byte-identical to upstream, so proxy/wireguard was repointed at it
// and the WireGuard outbound gained the AmneziaWG obfuscation parameters. Pinned to our local clone.
// Note this must be repeated in BOTH modules: a replace in a non-main module is ignored, and
// vendor/AndroidLibXrayLite is the main module for the gomobile build.
replace github.com/amnezia-vpn/amneziawg-go/v3 => ../../reference/amneziawg-go
