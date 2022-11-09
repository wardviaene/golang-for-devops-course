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
	resolverServers := []net.IP{}
	for _, ip := range strings.Split(ROOT_SERVERS, ",") {
		resolverServers = append(resolverServers, net.ParseIP(ip))
	}

	q, err := p.Question()
	if err != nil {
		if err == dnsmessage.ErrSectionDone {
			return
		}
		panic(err)
	}

	fmt.Printf("Found question for name '%s'\n", q.Name.String())
	fmt.Printf("Question header: %+v\n", queryHeader)
	response, err := dnsQuery(resolverServers, q.Name.String())
	if err != nil {
		panic(err)
	}
	response.Header.ID = queryHeader.ID
	err = sendResponse(response, addr, pc)
	if err != nil {
		fmt.Printf("Warning: %s", err)
	}
	fmt.Printf("Done!\n")
}

func dnsQuery(resolverServers []net.IP, hostnameToQuery string) (*dnsmessage.Message, error) {
	for i := 0; true; i++ {
		dnsAnswer, header, err := outgoingDnsQuery(resolverServers, hostnameToQuery)
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
		referralFound := false
		authorities, err := dnsAnswer.AllAuthorities()
		if err != nil {
			panic(err)
		}
		if len(authorities) > 0 {
			referralFound = true
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
		fmt.Printf("Additional records: %+v\n", additionalRecords)
		resolverServers = []net.IP{}
		for _, additionalRecord := range additionalRecords {
			if additionalRecord.Header.Type == dnsmessage.TypeA {
				for _, nsServer := range nsServers {
					if additionalRecord.Header.Name.String() == nsServer {
						resolverServers = append(resolverServers, net.IP(additionalRecord.Body.(*dnsmessage.AResource).A[:]))
					}
				}
			}
		}
		fmt.Printf("Resolver new data: %+v\n\n", resolverServers)
		if i == 3 || !referralFound { // we're not doing more iterations than 3
			return &dnsmessage.Message{
				Header: dnsmessage.Header{RCode: dnsmessage.RCodeNameError},
			}, nil
		}
	}
	return &dnsmessage.Message{
		Header: dnsmessage.Header{RCode: dnsmessage.RCodeNameError},
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

func outgoingDnsQuery(servers []net.IP, query string) (*dnsmessage.Parser, *dnsmessage.Header, error) {
	fmt.Printf("New dns query. DNS servers: %+v\n", servers)
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
	answer := make([]byte, 1024)
	n, err := bufio.NewReader(conn).Read(answer)
	if err != nil {
		return nil, nil, err
	}
	conn.Close()
	fmt.Printf("Response from server: %x\n", answer[:n])
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
