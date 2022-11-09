package dns

import (
	"bufio"
	"fmt"
	"math/rand"
	"net"
	"strings"

	"golang.org/x/net/dns/dnsmessage"
)

const ROOT_SERVERS = "198.41.0.4,199.9.14.201,192.33.4.12,199.7.91.13,192.203.230.10,192.5.5.241,192.112.36.4,198.97.190.53"

func Serve(pc net.PacketConn, addr net.Addr, buf []byte) {
	var (
		p           dnsmessage.Parser
		queryHeader dnsmessage.Header
		err         error
	)
	if queryHeader, err = p.Start(buf); err != nil {
		panic(err)
	}
	resolverServers := getRootServers()

	q, err := p.Question()
	if err != nil {
		if err == dnsmessage.ErrSectionDone {
			return
		}
		panic(err)
	}

	fmt.Printf("Incoming DNS query for: '%s'\n", q.Name.String())
	response, err := dnsQuery(resolverServers, q)
	if err != nil {
		panic(err)
	}
	response.Header.ID = queryHeader.ID
	err = sendResponse(response, addr, pc)
	if err != nil {
		fmt.Printf("Warning: %s", err)
	}
}

func dnsQuery(resolverServers []net.IP, question dnsmessage.Question) (*dnsmessage.Message, error) {
	for i := 0; i <= 3; i++ {
		dnsAnswer, header, err := outgoingDnsQuery(resolverServers, question)
		if err != nil {
			panic(err)
		}
		parsedAnswers, err := dnsAnswer.AllAnswers()
		if err != nil {
			panic(err)
		}

		if header.Authoritative {
			return &dnsmessage.Message{
				Header:  dnsmessage.Header{Response: true},
				Answers: parsedAnswers,
			}, nil

		}

		// we didn't send answer, need to update server
		authorities, err := dnsAnswer.AllAuthorities()
		if err != nil {
			panic(err)
		}
		if len(authorities) == 0 && !header.Authoritative {
			break // no authorities found and server is not authoritative
		}
		nsServers := make([]string, len(authorities))
		for k, authority := range authorities {
			if authority.Header.Type == dnsmessage.TypeNS {
				nsServers[k] = authority.Body.(*dnsmessage.NSResource).NS.String()
			}
		}
		additionalRecords, err := dnsAnswer.AllAdditionals()
		if err != nil {
			panic(err)
		}

		newResolverServersFound := false
		resolverServers = []net.IP{}
		for _, additionalRecord := range additionalRecords {
			if additionalRecord.Header.Type == dnsmessage.TypeA {
				for _, nsServer := range nsServers {
					if additionalRecord.Header.Name.String() == nsServer {
						newResolverServersFound = true
						resolverServers = append(resolverServers, net.IP(additionalRecord.Body.(*dnsmessage.AResource).A[:]))
					}
				}
			}
		}

		if !newResolverServersFound {
			for _, nsServer := range nsServers {
				if !newResolverServersFound {
					response, err := dnsQuery(getRootServers(), dnsmessage.Question{Name: dnsmessage.MustNewName(nsServer), Type: dnsmessage.TypeA, Class: dnsmessage.ClassINET})
					if err != nil {
						fmt.Printf("Warning: failed to lookup nsServer: %s\n", err)
					} else {
						for _, answer := range response.Answers {
							if answer.Header.Type == dnsmessage.TypeA {
								resolverServers = append(resolverServers, net.IP(answer.Body.(*dnsmessage.AResource).A[:]))
								newResolverServersFound = true
							}
						}
					}
				}
			}
		}

		if !newResolverServersFound {
			return &dnsmessage.Message{
				Header: dnsmessage.Header{RCode: dnsmessage.RCodeServerFailure},
			}, nil
		}
	}
	return &dnsmessage.Message{
		Header: dnsmessage.Header{RCode: dnsmessage.RCodeServerFailure},
	}, nil
}

func sendResponse(response *dnsmessage.Message, addr net.Addr, pc net.PacketConn) error {
	buf, err := response.Pack()
	if err != nil {
		return err
	}
	_, err = pc.WriteTo(buf, addr)
	if err != nil {
		return err
	}
	return nil
}

func outgoingDnsQuery(servers []net.IP, question dnsmessage.Question) (*dnsmessage.Parser, *dnsmessage.Header, error) {
	fmt.Printf("New outgoing dns query for %s, servers: %+v\n", question.Name.String(), servers)
	msg := dnsmessage.Message{
		Header: dnsmessage.Header{
			RCode:            dnsmessage.RCode(0),
			ID:               uint16(rand.Intn(int(^uint16(0)))),
			OpCode:           dnsmessage.OpCode(0),
			Response:         false,
			AuthenticData:    false,
			RecursionDesired: false,
		},
		Questions: []dnsmessage.Question{question},
	}
	buf, err := msg.Pack()
	if err != nil {
		return nil, nil, err
	}
	var conn net.Conn
	for _, server := range servers {
		if server.String() != "<nil>" {
			conn, err = net.Dial("udp", server.String()+":53")
			if err == nil {
				break
			}
		}
	}
	if conn == nil {
		return nil, nil, fmt.Errorf("tried all servers. Error: %s", err)
	}

	_, err = conn.Write(buf)
	if err != nil {
		return nil, nil, err
	}
	answer := make([]byte, 512)
	n, err := bufio.NewReader(conn).Read(answer)
	if err != nil {
		return nil, nil, err
	}
	conn.Close()
	var p dnsmessage.Parser
	var header dnsmessage.Header
	if header, err = p.Start(answer[:n]); err != nil {
		return nil, nil, err
	}

	questions, err := p.AllQuestions()
	if err != nil {
		return nil, nil, err
	}
	if len(questions) != len(msg.Questions) {
		return nil, nil, fmt.Errorf("questions in request and response don't match")
	}
	if questions[0].Name != msg.Questions[0].Name {
		return nil, nil, fmt.Errorf("question Name in request and response don't match")
	}
	if questions[0].Type != msg.Questions[0].Type {
		return nil, nil, fmt.Errorf("question Name in request and response don't match")
	}
	if questions[0].Class != msg.Questions[0].Class {
		return nil, nil, fmt.Errorf("question Class in request and response don't match")
	}

	err = p.SkipAllQuestions()
	if err != nil {
		return nil, nil, err
	}

	return &p, &header, err
}

func getRootServers() []net.IP {
	rootServers := []net.IP{}
	for _, ip := range strings.Split(ROOT_SERVERS, ",") {
		rootServers = append(rootServers, net.ParseIP(ip))
	}
	return rootServers
}

