package zeddns

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"time"
)

const (
	socksVersion   = 0x05
	socksAuthNone  = 0x00
	socksAuthUser  = 0x02
	socksAuthNoAcc = 0xFF

	socksUserAuthVersion = 0x01
	socksCmdConnect      = 0x01

	socksAddrIPv4   = 0x01
	socksAddrDomain = 0x03
	socksAddrIPv6   = 0x04
)

type socks5Dialer struct {
	proxy    string
	username string
	password string
}

func (d socks5Dialer) DialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	if network != "tcp" && network != "tcp4" && network != "tcp6" {
		return nil, fmt.Errorf("socks5 carries tcp only, not %s", network)
	}
	conn, err := (&net.Dialer{Timeout: 10 * time.Second}).DialContext(ctx, "tcp", d.proxy)
	if err != nil {
		return nil, err
	}
	if deadline, ok := ctx.Deadline(); ok {
		_ = conn.SetDeadline(deadline)
	} else {
		_ = conn.SetDeadline(time.Now().Add(15 * time.Second))
	}
	if err := d.handshake(conn, addr); err != nil {
		conn.Close()
		return nil, err
	}
	_ = conn.SetDeadline(time.Time{})
	return conn, nil
}

func (d socks5Dialer) handshake(conn net.Conn, addr string) error {
	methods := []byte{socksAuthNone}
	if d.username != "" {
		methods = []byte{socksAuthUser, socksAuthNone}
	}
	greeting := append([]byte{socksVersion, byte(len(methods))}, methods...)
	if _, err := conn.Write(greeting); err != nil {
		return err
	}

	var choice [2]byte
	if _, err := io.ReadFull(conn, choice[:]); err != nil {
		return err
	}
	if choice[0] != socksVersion {
		return errors.New("proxy is not SOCKS5")
	}
	switch choice[1] {
	case socksAuthNone:
	case socksAuthUser:
		if err := writeUserAuth(conn, d.username, d.password); err != nil {
			return err
		}
	default:
		return errors.New("proxy rejected every authentication method")
	}

	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		return err
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return fmt.Errorf("bad port in %q", addr)
	}

	req := []byte{socksVersion, socksCmdConnect, 0x00}
	switch ip := net.ParseIP(host); {
	case ip == nil:
		if len(host) > 255 {
			return errors.New("hostname too long for SOCKS5")
		}
		req = append(req, socksAddrDomain, byte(len(host)))
		req = append(req, host...)
	case ip.To4() != nil:
		req = append(req, socksAddrIPv4)
		req = append(req, ip.To4()...)
	default:
		req = append(req, socksAddrIPv6)
		req = append(req, ip.To16()...)
	}
	req = append(req, byte(port>>8), byte(port))
	if _, err := conn.Write(req); err != nil {
		return err
	}
	return readConnectReply(conn)
}

func writeUserAuth(conn net.Conn, user, pass string) error {
	if len(user) > 255 || len(pass) > 255 {
		return errors.New("SOCKS5 credentials too long")
	}
	msg := []byte{socksUserAuthVersion, byte(len(user))}
	msg = append(msg, user...)
	msg = append(msg, byte(len(pass)))
	msg = append(msg, pass...)
	if _, err := conn.Write(msg); err != nil {
		return err
	}
	var reply [2]byte
	if _, err := io.ReadFull(conn, reply[:]); err != nil {
		return err
	}
	if reply[1] != 0x00 {
		return errors.New("proxy rejected the credentials")
	}
	return nil
}

func readConnectReply(conn net.Conn) error {
	var head [4]byte
	if _, err := io.ReadFull(conn, head[:]); err != nil {
		return err
	}
	if head[1] != 0x00 {
		return fmt.Errorf("proxy refused the connection (status %d)", head[1])
	}
	return skipBoundAddress(conn, head[3])
}

func skipBoundAddress(conn net.Conn, addrType byte) error {
	var n int
	switch addrType {
	case socksAddrIPv4:
		n = 4
	case socksAddrIPv6:
		n = 16
	case socksAddrDomain:
		var l [1]byte
		if _, err := io.ReadFull(conn, l[:]); err != nil {
			return err
		}
		n = int(l[0])
	default:
		return errors.New("proxy returned an unknown address type")
	}
	_, err := io.CopyN(io.Discard, conn, int64(n)+2)
	return err
}

type credentialRelay struct {
	username string
	password string
}

func (r credentialRelay) negotiate(client net.Conn, server net.Conn) error {
	if err := r.acceptLocal(client); err != nil {
		return err
	}
	return r.authenticateUpstream(server)
}

func (r credentialRelay) acceptLocal(client net.Conn) error {
	var head [2]byte
	if _, err := io.ReadFull(client, head[:]); err != nil {
		return err
	}
	if head[0] != socksVersion {
		return errors.New("local client is not SOCKS5")
	}
	if _, err := io.CopyN(io.Discard, client, int64(head[1])); err != nil {
		return err
	}
	_, err := client.Write([]byte{socksVersion, socksAuthNone})
	return err
}

func (r credentialRelay) authenticateUpstream(server net.Conn) error {
	if _, err := server.Write([]byte{socksVersion, 0x01, socksAuthUser}); err != nil {
		return err
	}
	var choice [2]byte
	if _, err := io.ReadFull(server, choice[:]); err != nil {
		return err
	}
	if choice[0] != socksVersion {
		return errors.New("tunnel exit is not SOCKS5")
	}
	switch choice[1] {
	case socksAuthUser:
		return writeUserAuth(server, r.username, r.password)
	case socksAuthNone:
		return nil
	case socksAuthNoAcc:
		return errors.New("tunnel exit accepts no authentication method we offer")
	default:
		return fmt.Errorf("tunnel exit chose an unsupported authentication method (%d)", choice[1])
	}
}
