package dns

import (
	"fmt"
	"net"

	"golang.org/x/net/dns/dnsmessage"
)

const ROOT_SERVERS = "198.41.0.4,199.9.14.201,192.33.4.12,199.7.91.13,192.203.230.10,192.5.5.241,192.112.36.4,198.97.190.53"

func handlePacket(pc net.PacketConn, addr net.Addr, buf []byte) error {
	return fmt.Errorf("not implemented yet")
}

func outgoingDnsQuery(servers []net.IP, question dnsmessage.Question) (*dnsmessage.Parser, *dnsmessage.Header, error) {
	return nil, nil, nil
}
