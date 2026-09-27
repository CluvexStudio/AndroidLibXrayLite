package zeddns

import "github.com/net2share/vaydns/dns"

func client_NewTunnelServerForTest(domain string) (dns.Name, error) {
	return dns.ParseName(domain)
}
