package dns

import (
	"crypto/rand"
	"math/big"
	"net"
	"strings"
	"testing"
	"time"

	"golang.org/x/net/dns/dnsmessage"
)

type MockPacketConn struct{}

func (m *MockPacketConn) WriteTo(p []byte, addr net.Addr) (n int, err error) {
	return 0, nil
}

func (m *MockPacketConn) Close() error {
	return nil
}

func (m *MockPacketConn) ReadFrom(p []byte) (n int, addr net.Addr, err error) {
	return 0, nil, nil
}
func (m *MockPacketConn) LocalAddr() net.Addr {
	return nil
}
func (m *MockPacketConn) SetDeadline(t time.Time) error {
	return nil
}
func (m *MockPacketConn) SetReadDeadline(t time.Time) error {
	return nil
}
func (m *MockPacketConn) SetWriteDeadline(t time.Time) error {
	return nil
}

func TestHandlePacket(t *testing.T) {
	names := []string{"www.google.com.", "www.amazon.com."}
	for _, name := range names {
		max := ^uint16(0)
		randomNumber, err := rand.Int(rand.Reader, big.NewInt(int64(max)))
		if err != nil {
			t.Fatalf("rand error: %s", err)
		}
		message := dnsmessage.Message{
			Header: dnsmessage.Header{
				RCode:            dnsmessage.RCode(0),
				ID:               uint16(randomNumber.Int64()),
				OpCode:           dnsmessage.OpCode(0),
				Response:         false,
				AuthenticData:    false,
				RecursionDesired: false,
			},
			Questions: []dnsmessage.Question{
				{
					Name:  dnsmessage.MustNewName(name),
					Type:  dnsmessage.TypeA,
					Class: dnsmessage.ClassINET,
				},
			},
		}
		buf, err := message.Pack()
		if err != nil {
			t.Fatalf("Pack error: %s", err)
		}

		err = handlePacket(&MockPacketConn{}, &net.IPAddr{IP: net.ParseIP("127.0.0.1")}, buf)
		if err != nil {
			t.Fatalf("serve error: %s", err)
		}
	}
}

func TestOutgoingDnsQuery(t *testing.T) {
	question := dnsmessage.Question{
		Name:  dnsmessage.MustNewName("com."),
		Type:  dnsmessage.TypeNS,
		Class: dnsmessage.ClassINET,
	}

	if len(ROOT_SERVERS) == 0 {
		t.Fatalf("No root servers found")
	}

	rootServers := strings.Split(ROOT_SERVERS, ",")

	servers := []net.IP{net.ParseIP(rootServers[0])}
	dnsAnswer, header, err := outgoingDnsQuery(servers, question)
	if err != nil {
		t.Fatalf("outgoingDnsQuery error: %s", err)
	}
	if header == nil {
		t.Fatalf("No header found")
	}
	if dnsAnswer == nil {
		t.Fatalf("no answer found")
	}
	if header.RCode != dnsmessage.RCodeSuccess {
		t.Fatalf("response was not succesful (maybe the DNS server has changed?)")
	}
	err = dnsAnswer.SkipAllAnswers()
	if err != nil {
		t.Fatalf("SkipAllAnswers error: %s", err)
	}
	parsedAuthorities, err := dnsAnswer.AllAuthorities()
	if err != nil {
		t.Fatalf("Error getting answers")
	}
	if len(parsedAuthorities) == 0 {
		t.Fatalf("No answers received")
	}
}
