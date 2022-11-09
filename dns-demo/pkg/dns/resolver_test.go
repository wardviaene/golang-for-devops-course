package dns

import (
	"encoding/hex"
	"fmt"
	"math/rand"
	"net"
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

func TestServe(t *testing.T) {
	names := []string{"www.google.com.", "www.amazon.com."}
	for _, name := range names {
		message := dnsmessage.Message{
			Header: dnsmessage.Header{
				RCode:            dnsmessage.RCode(0),
				ID:               uint16(rand.Intn(int(^uint16(0)))),
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

		Serve(&MockPacketConn{}, &net.IPAddr{IP: net.ParseIP("127.0.0.1")}, buf)
	}
}

func TestUnpacking(t *testing.T) {
	reply := `cf868400000100060000000006676f6f676c6503636f6d0000010001c00c000100010000012c00048efa8a8bc00c000100010000012c00048efa8a71c00c000100010000012c00048efa8a65c00c000100010000012c00048efa8a64c00c000100010000012c00048efa8a66c00c000100010000012c00048efa8a8a`
	data, err := hex.DecodeString(reply)
	if err != nil {
		panic(err)
	}
	var p dnsmessage.Parser
	if _, err = p.Start(data); err != nil {
		t.Errorf("Error: %s", err)
	}

	p.SkipAllQuestions()

	h, err := p.AnswerHeader()
	fmt.Printf("AnswerHeader: %+v", h)
	if err != nil {
		t.Errorf("AnswerHeader error: %s", err)
	}
	_, err = p.Answer()
	if err != nil {
		t.Errorf("Answer error: %s", err)
	}
	fmt.Printf("% x", data)
}

