module github.com/2dust/AndroidLibXrayLite

go 1.27

replace github.com/xtls/xray-core => ../xray-core

// gomobile binds these as separate roots; they are NOT imported by this module's own Go code, so
// `go mod tidy` prunes them and the bind then fails with "no exported names". Keep them here
// explicitly and do NOT run `go mod tidy` on this module.
require (
	github.com/Psiphon-Labs/psiphon-tunnel-core v1.0.11-0.20250319154633-ceb78316d06e
	github.com/apernet/quic-go v0.61.1-0.20260806010916-184d081eef3e
	// sing-box: bound as its own root (experimental/libbox, the full engine); Xray also links it for
	// the "singbox" outbound.
	github.com/sagernet/sing-box v1.14.5-0.20261119022534-039d829d20c7
	github.com/xtls/xray-core v1.260327.1-0.20261109052709-676eaf66f8f0
	golang.org/x/mobile v0.0.0-20260709172247-6129f5bee9d5
	masterdnsvpn-go v0.0.0
	zeddns v0.0.0
)

require (
	filippo.io/age v1.3.1 // indirect
	filippo.io/edwards25519 v1.2.0 // indirect
	filippo.io/hpke v0.4.0 // indirect
	github.com/BurntSushi/toml v1.4.1-0.20240526193622-a339e1f7089c // indirect
	github.com/Jigsaw-Code/outline-sdk v0.0.16 // indirect
	github.com/Psiphon-Labs/bolt v0.0.0-20200624191537-23cedaef7ad7 // indirect
	github.com/Psiphon-Labs/consistent v0.0.0-20240322131436-20aaa4e05737 // indirect
	github.com/Psiphon-Labs/covert-dtls v0.0.0-20260423180104-2d3b15cbc8be // indirect
	github.com/Psiphon-Labs/pion-ice/v4 v4.0.0-20260423180031-37d6e92d021f // indirect
	github.com/Psiphon-Labs/pion-webrtc/v4 v4.0.0-20260423180049-7ff453f9681a // indirect
	github.com/Psiphon-Labs/psiphon-tls v0.0.0-20260429174135-190516348878 // indirect
	github.com/Psiphon-Labs/quic-go v0.0.0-20250527153145-79fe45fb83b1 // indirect
	github.com/Psiphon-Labs/utls v1.1.1-0.20260423193914-69eb96d02588 // indirect
	github.com/RyuaNerin/go-krypto v1.3.0 // indirect
	github.com/ajg/form v1.5.1 // indirect
	github.com/amnezia-vpn/amneziawg-go/v3 v3.0.0-00010101000000-000000000000 // indirect
	github.com/anchore/go-lzo v0.1.0 // indirect
	github.com/andybalholm/brotli v1.2.3 // indirect
	github.com/anmitsu/go-shlex v0.0.0-20200514113438-38f4b401e2be // indirect
	github.com/anytls/sing-anytls v0.0.11 // indirect
	github.com/axiomhq/hyperloglog v0.2.6 // indirect
	github.com/bifurcation/mint v0.0.0-20180306135233-198357931e61 // indirect
	github.com/bits-and-blooms/bitset v1.10.0 // indirect
	github.com/bits-and-blooms/bloom/v3 v3.6.0 // indirect
	github.com/caddyserver/certmagic v0.25.3-0.20260421143802-60d9d8b415d6 // indirect
	github.com/caddyserver/zerossl v0.1.5 // indirect
	github.com/cespare/xxhash v1.1.0 // indirect
	github.com/cheekybits/genny v0.0.0-20170328200008-9127e812e1e9 // indirect
	github.com/cloudflare/circl v1.6.5 // indirect
	github.com/coder/websocket v1.8.14 // indirect
	github.com/cognusion/go-cache-lru v0.0.0-20170419142635-f73e2280ecea // indirect
	github.com/coreos/go-iptables v0.7.1-0.20240112124308-65c67c9f46e6 // indirect
	github.com/creack/pty v1.1.24 // indirect
	github.com/cretz/bine v0.2.0 // indirect
	github.com/database64128/netx-go v0.1.1 // indirect
	github.com/database64128/tfo-go/v2 v2.3.2 // indirect
	github.com/dgryski/go-camellia v0.0.0-20191119043421-69a8a13fb23d // indirect
	github.com/dgryski/go-metro v0.0.0-20250106013310-edb8663e5e33 // indirect
	github.com/florianl/go-nfqueue/v2 v2.1.0 // indirect
	github.com/flynn/noise v1.1.0 // indirect
	github.com/fsnotify/fsnotify v1.9.0 // indirect
	github.com/fxamacker/cbor/v2 v2.9.0 // indirect
	github.com/gaissmai/bart v0.26.1 // indirect
	github.com/ghodss/yaml v1.0.1-0.20220118164431-d8423dcdf344 // indirect
	github.com/go-chi/chi/v5 v5.2.5 // indirect
	github.com/go-chi/render v1.0.3 // indirect
	github.com/go-ole/go-ole v1.3.0 // indirect
	github.com/go4org/hashtriemap v0.0.0-20251130024219-545ba229f689 // indirect
	github.com/gobwas/httphead v0.1.0 // indirect
	github.com/gobwas/pool v0.2.1 // indirect
	github.com/godbus/dbus/v5 v5.2.2 // indirect
	github.com/gofrs/uuid/v5 v5.5.1 // indirect
	github.com/golang/groupcache v0.0.0-20241129210726-2c02b8208cf8 // indirect
	github.com/google/btree v1.1.3 // indirect
	github.com/google/certificate-transparency-go v1.3.2 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/google/gopacket v1.1.19 // indirect
	github.com/google/nftables v0.2.1-0.20240414091927-5e242ec57806 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/gorilla/websocket v1.5.3 // indirect
	github.com/grafov/m3u8 v0.0.0-20171211212457-6ab8f28ed427 // indirect
	github.com/hashicorp/yamux v0.1.2 // indirect
	github.com/hdevalence/ed25519consensus v0.2.0 // indirect
	github.com/huin/goupnp v1.3.0 // indirect
	github.com/jackpal/go-nat-pmp v1.0.2 // indirect
	github.com/jsimonetti/rtnetlink v1.4.1 // indirect
	github.com/juju/ratelimit v1.0.2 // indirect
	github.com/kamstrup/intmap v0.5.2 // indirect
	github.com/klauspost/compress v1.20.1 // indirect
	github.com/klauspost/cpuid/v2 v2.4.2 // indirect
	github.com/klauspost/reedsolomon v1.13.0 // indirect
	github.com/koron/go-ssdp v0.0.4 // indirect
	github.com/kr/fs v0.1.0 // indirect
	github.com/libdns/acmedns v0.5.0 // indirect
	github.com/libdns/alidns v1.0.6 // indirect
	github.com/libdns/cloudflare v0.2.2 // indirect
	github.com/libdns/libdns v1.1.1 // indirect
	github.com/libp2p/go-nat v1.0.1-0.20250821073202-01afc089f138 // indirect
	github.com/libp2p/go-netroute v0.2.1 // indirect
	github.com/logrusorgru/aurora v2.0.3+incompatible // indirect
	github.com/marusama/semaphore v0.0.0-20171214154724-565ffd8e868a // indirect
	github.com/mdlayher/netlink v1.11.2 // indirect
	github.com/mdlayher/socket v0.6.0 // indirect
	github.com/metacubex/utls v1.9.0-alpha.1 // indirect
	github.com/mholt/acmez/v3 v3.1.6 // indirect
	github.com/miekg/dns v1.1.73 // indirect
	github.com/mitchellh/go-ps v1.0.0 // indirect
	github.com/molecule-man/go-brrr v1.2.0 // indirect
	github.com/net2share/vaydns v0.0.0-00010101000000-000000000000 // indirect
	github.com/oschwald/maxminddb-golang v1.13.1 // indirect
	github.com/pelletier/go-toml v1.9.5 // indirect
	github.com/pierrec/lz4/v4 v4.1.26 // indirect
	github.com/pion/datachannel v1.6.0 // indirect
	github.com/pion/dtls/v3 v3.1.5 // indirect
	github.com/pion/interceptor v0.1.44 // indirect
	github.com/pion/logging v0.2.4 // indirect
	github.com/pion/mdns/v2 v2.1.0 // indirect
	github.com/pion/randutil v0.1.0 // indirect
	github.com/pion/rtcp v1.2.16 // indirect
	github.com/pion/rtp v1.10.1 // indirect
	github.com/pion/sctp v1.9.2 // indirect
	github.com/pion/sdp/v3 v3.0.18 // indirect
	github.com/pion/srtp/v3 v3.0.10 // indirect
	github.com/pion/stun/v3 v3.1.7 // indirect
	github.com/pion/transport/v4 v4.1.0 // indirect
	github.com/pion/turn/v4 v4.1.4 // indirect
	github.com/pires/go-proxyproto v0.16.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pkg/sftp v1.13.10 // indirect
	github.com/quic-go/qpack v0.6.0 // indirect
	github.com/refraction-networking/utls v1.8.3-0.20261108164504-a13c59ffe98b // indirect
	github.com/robfig/cron/v3 v3.0.1 // indirect
	github.com/rogpeppe/go-internal v1.14.1 // indirect
	github.com/safchain/ethtool v0.3.0 // indirect
	github.com/sagernet/bbolt v0.0.0-20260915102804-500ee1e84832 // indirect
	github.com/sagernet/cors v1.2.1 // indirect
	github.com/sagernet/cronet-go v0.0.0-20261118105239-7a29333436ca // indirect
	github.com/sagernet/cronet-go/all v0.0.0-20261118105239-7a29333436ca // indirect
	github.com/sagernet/cronet-go/lib/android_386 v0.0.0-20261118104455-4ecba4761b20 // indirect
	github.com/sagernet/cronet-go/lib/android_amd64 v0.0.0-20261118104455-4ecba4761b20 // indirect
	github.com/sagernet/cronet-go/lib/android_arm v0.0.0-20261118104455-4ecba4761b20 // indirect
	github.com/sagernet/cronet-go/lib/android_arm64 v0.0.0-20261118104455-4ecba4761b20 // indirect
	github.com/sagernet/cronet-go/lib/linux_amd64 v0.0.0-20261118104455-4ecba4761b20 // indirect
	github.com/sagernet/fswatch v0.1.2 // indirect
	github.com/sagernet/gliderssh v0.3.4-0.20261023092853-e0cad6d28647 // indirect
	github.com/sagernet/gvisor v0.0.0-20260727.0-sing-box-mod.1 // indirect
	github.com/sagernet/netlink v0.0.0-20261116064630-4ea47d5d44b1 // indirect
	github.com/sagernet/nftables v0.3.0-mod.4.0.20261114074246-0b310c3e010b // indirect
	github.com/sagernet/quic-go v0.61.0-sing-box-mod.14 // indirect
	github.com/sagernet/sing v0.9.7-0.20261110095940-27b73774b240 // indirect
	github.com/sagernet/sing-anytls v0.0.0-20260928104022-580984e4d8cb // indirect
	github.com/sagernet/sing-mux v0.3.10-0.20261110064256-375087483ea2 // indirect
	github.com/sagernet/sing-openconnect v0.1.1-0.20261119021124-31db24231feb // indirect
	github.com/sagernet/sing-openvpn v0.1.1-0.20261110111939-ad7a6ce7d11b // indirect
	github.com/sagernet/sing-quic v0.7.2-0.20261114063120-0979e001618d // indirect
	github.com/sagernet/sing-shadowsocks v0.2.9 // indirect
	github.com/sagernet/sing-shadowsocks2 v0.2.2-0.20261110092058-52ade8ab7e95 // indirect
	github.com/sagernet/sing-shadowtls v0.2.2-0.20261114073233-97863830a3b3 // indirect
	github.com/sagernet/sing-snell v0.0.0-20261110094602-04667b0a321c // indirect
	github.com/sagernet/sing-tun v0.9.7-0.20261119022030-4486c14ba472 // indirect
	github.com/sagernet/sing-usbip v0.0.0-20260817040617-28bd42667eca // indirect
	github.com/sagernet/sing-vmess v0.2.9-0.20261116164733-ca746ebade33 // indirect
	github.com/sagernet/smux v1.5.50-sing-box-mod.1 // indirect
	github.com/sagernet/tailscale v1.102.1-sing-box-1.14-mod.5.0.20261028024127-670967df7dbf // indirect
	github.com/sagernet/wireguard-go v0.0.8-0.20261022215849-fdf0ad046586 // indirect
	github.com/sagernet/ws v0.0.0-20231204124109-acfe8907c854 // indirect
	github.com/shadowsocks/go-shadowsocks2 v0.1.5 // indirect
	github.com/sirupsen/logrus v1.9.4 // indirect
	github.com/smallstep/pkcs7 v0.1.1 // indirect
	github.com/syndtr/gocapability v0.0.0-20170704070218-db04d3cc01c8 // indirect
	github.com/tailscale/goupnp v1.0.1-0.20210804011211-c64d0f06ea05 // indirect
	github.com/tailscale/hujson v0.0.0-20260302212456-ecc657c15afd // indirect
	github.com/tailscale/netlink v1.1.1-0.20240822203006-4d49adab4de7 // indirect
	github.com/tailscale/peercred v0.0.0-20250107143737-35a0c7bd7edc // indirect
	github.com/tjfoc/gmsm v1.4.1 // indirect
	github.com/vishvananda/netlink v1.3.1 // indirect
	github.com/vishvananda/netns v0.0.5 // indirect
	github.com/wader/filtertransport v0.0.0-20200316221534-bdd9e61eee78 // indirect
	github.com/wlynxg/anet v0.0.5 // indirect
	github.com/x448/float16 v0.8.4 // indirect
	github.com/xtaci/kcp-go/v5 v5.6.61 // indirect
	github.com/xtaci/smux v1.5.50 // indirect
	github.com/xtls/reality v0.0.0-20261108101329-4bb46b2afd8b // indirect
	github.com/youmark/pkcs8 v0.0.0-20240726163527-a2c0da244d78 // indirect
	github.com/yuin/gopher-lua v1.1.2 // indirect
	github.com/zeebo/blake3 v0.2.4 // indirect
	go.uber.org/mock v0.6.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.27.1 // indirect
	go.uber.org/zap/exp v0.3.0 // indirect
	go4.org/mem v0.0.0-20240501181205-ae6ca9944745 // indirect
	go4.org/netipx v0.0.0-20231129151722-fdeea329fbba // indirect
	golang.org/x/crypto v0.58.0 // indirect
	golang.org/x/crypto/x509roots/fallback v0.0.0-20261020153306-9f2b28d4331a // indirect
	golang.org/x/exp v0.0.0-20260410095643-746e56fc9e2f // indirect
	golang.org/x/mod v0.41.0 // indirect
	golang.org/x/net v0.62.0 // indirect
	golang.org/x/oauth2 v0.36.0 // indirect
	golang.org/x/sync v0.24.0 // indirect
	golang.org/x/sys v0.49.0 // indirect
	golang.org/x/term v0.47.0 // indirect
	golang.org/x/text v0.43.0 // indirect
	golang.org/x/time v0.15.0 // indirect
	golang.zx2c4.com/wintun v0.0.0-20230126152724-0fa3db229ce2 // indirect
	golang.zx2c4.com/wireguard v0.0.0-20250521234502-f333402bd9cb // indirect
	golang.zx2c4.com/wireguard/windows v1.1.1 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260817212433-ac3dfec99bb1 // indirect
	google.golang.org/grpc v1.85.0 // indirect
	google.golang.org/protobuf v1.36.12 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gvisor.dev/gvisor v0.0.0-20260122175437-89a5d21be8f0 // indirect
	layeh.com/gopher-luar v1.0.11 // indirect
	lukechampine.com/blake3 v1.4.1 // indirect
	tailscale.com v1.58.2 // indirect
)

replace github.com/Psiphon-Labs/psiphon-tunnel-core => ../../reference/psiphon-tunnel-core

replace github.com/pion/dtls/v2 => ../../reference/psiphon-tunnel-core/replace/dtls

replace github.com/quic-go/qpack => ../qpack-patched

replace zeddns => ./zeddns

// MasterDnsVPN DNS tunnel (masterking32/MasterDnsVPN), bound via its mobile/ wrapper.
replace masterdnsvpn-go => ../../reference/MasterDnsVPN

replace github.com/net2share/vaydns => ../../reference/vaydns

// AmneziaWG. amneziawg-go is a fork of wireguard-go whose conn.Bind and tun.Device interfaces and
// device.NewDevice signature are byte-identical to upstream, so proxy/wireguard was repointed at it
// and the WireGuard outbound gained the AmneziaWG obfuscation parameters. Pinned to our local clone.
// Note this must be repeated in BOTH modules: a replace in a non-main module is ignored, and
// vendor/AndroidLibXrayLite is the main module for the gomobile build.
replace github.com/amnezia-vpn/amneziawg-go/v3 => ../../reference/amneziawg-go

replace github.com/sagernet/sing-openvpn => ../../reference/sing-openvpn

