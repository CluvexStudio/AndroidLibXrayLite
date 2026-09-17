# ZedSecure patches

This branch is [2dust/AndroidLibXrayLite](https://github.com/2dust/AndroidLibXrayLite) at commit
`544a734` with the changes ZedSecure needs. Compare it against `master` to see them.

## What was changed

* `libv2ray_autoselect.go` — the binding for the `autoselect` outbound added to our Xray fork.
* `go.mod` / `go.sum` — the module is the gomobile build's main module, so every engine bound into
  the AAR is required from here: sing-box, Psiphon, the DNS tunnels (dnstt, VayDNS, noizdns,
  MasterDnsVPN) and the AmneziaWG swap. Several are resolved by `replace` to paths outside this
  repository; <https://github.com/CluvexStudio/zedcore> is the tree where those paths resolve, and
  is where a build should start.

  gomobile binds those packages as separate roots — this module's own Go code does not import them —
  so `go mod tidy` prunes them and the bind then fails with `no exported names`. **Do not run
  `go mod tidy` here.**

The geo data under `data/` is not committed; `./gen_assets.sh` downloads it.

## Licensing

AndroidLibXrayLite is **LGPL-3.0**; `LICENSE` is unchanged and still applies.

The AAR built from it links sing-box, sing, sing-shadowsocks and psiphon-tunnel-core
(GPL-3.0-or-later) and noizdns (AGPL-3.0), so the built library as a whole is distributed under the
GPL-3.0-or-later, with the AGPL-3.0 terms continuing to apply to the noizdns portion. LGPL-3.0 and
MPL-2.0 both permit this. Corresponding source for the whole of it: the zedcore repository above.

Independent fork. Not affiliated with, endorsed by, or associated with the sing-box project or its
authors.
