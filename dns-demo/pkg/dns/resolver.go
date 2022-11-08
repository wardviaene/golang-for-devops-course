package dns

import (
	"bufio"
	"fmt"
	"math/rand"
	"net"

	"golang.org/x/net/dns/dnsmessage"
)

const A_ROOT_SERVER = "198.41.0.4"

func Serve(pc net.PacketConn, addr net.Addr, buf []byte) {

	fmt.Printf("Len: %d\n", len(buf))

	fmt.Printf("hex: %x\n\n", buf)

	var (
		p           dnsmessage.Parser
		queryHeader dnsmessage.Header
		err         error
	)
	if queryHeader, err = p.Start(buf); err != nil {
		panic(err)
	}
	//resolverServer := A_ROOT_SERVER
	resolverServer := "216.239.32.10"
	for {
		q, err := p.Question()
		if err != nil {
			if err == dnsmessage.ErrSectionDone {
				break
			}
			panic(err)
		}

		fmt.Printf("Found question for name '%s'\n", q.Name.String())
		fmt.Printf("Question header: %+v\n", queryHeader)
		if q.Type.String() == "TypeA" && q.Class.String() == "ClassINET" { // we have an A request
			dnsAnswer, err := dnsQuery(resolverServer, q.Name.String())
			if err != nil {
				panic(err)
			}
			parsedAnswer, err := dnsAnswer.Answer()
			if err != nil {
				panic(err)
			}

			fmt.Printf("Answer: %s\n", parsedAnswer.Body.GoString())
			response := dnsmessage.Message{
				Header: dnsmessage.Header{Response: true, ID: queryHeader.ID},
				Answers: []dnsmessage.Resource{
					parsedAnswer,
				},
			}
			buf, err := response.Pack()
			if err != nil {
				panic(err)
			}
			_, err = pc.WriteTo(buf, addr)
			if err != nil {
				panic(err)
			}
		}
		if err := p.SkipAllQuestions(); err != nil {
			panic(err)
		}
		break
	}
	fmt.Printf("Done!\n")
}

/*
func nsResolver(addr, toResolve string) (addrs []*net.NS, err error) {
	r := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{
				Timeout: time.Millisecond * time.Duration(10000),
			}
			return d.DialContext(ctx, network, addr+":53")
		},
	}
	return r.LookupNS(context.Background(), toResolve)
}

func aResolver(addr, toResolve string) ([]string, error) {
	r := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{
				Timeout: time.Millisecond * time.Duration(10000),
			}
			return d.DialContext(ctx, network, addr+":53")
		},
	}
	return r.LookupAddr(context.Background(), toResolve)
}
*/

func dnsQuery(server, query string) (*dnsmessage.Parser, error) {

	msg := dnsmessage.Message{
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
				Name:  dnsmessage.MustNewName(query),
				Type:  dnsmessage.TypeA,
				Class: dnsmessage.ClassINET,
			},
		},
	}
	fmt.Printf("Message to google: %+v\n", msg)
	buf, err := msg.Pack()
	if err != nil {
		return nil, err
	}
	conn, err := net.Dial("udp", server+":53")
	if err != nil {
		return nil, err
	}
	fmt.Printf("Message to google (hex): %x\n", buf)
	nn, err := conn.Write(buf)
	if err != nil {
		return nil, err
	}
	fmt.Printf("Wrote bytes: %d (length of buffer was: %d)\n", nn, len(buf))
	answer := make([]byte, 1024)
	n, err := bufio.NewReader(conn).Read(answer)
	if err != nil {
		return nil, err
	}
	conn.Close()
	fmt.Printf("Response from google: %x\n", answer[:n])
	var p dnsmessage.Parser
	if _, err = p.Start(answer[:n]); err != nil {
		return nil, err
	}

	answerTmp, err := p.Answer()
	fmt.Printf("answer: %+v\n", answerTmp)

	return &p, err
}
