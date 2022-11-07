package dns

import (
	"context"
	"fmt"
	"net"
	"time"

	"golang.org/x/net/dns/dnsmessage"
)

const A_ROOT_SERVER = "198.41.0.4"

func Serve(pc net.PacketConn, addr net.Addr, buf []byte) {

	fmt.Printf("Len: %d\n", len(buf))

	fmt.Printf("hex: %x\n\n", buf)

	var p dnsmessage.Parser
	if _, err := p.Start(buf); err != nil {
		panic(err)
	}

	for {
		q, err := p.Question()
		if err == dnsmessage.ErrSectionDone {
			break
		}
		if err != nil {
			panic(err)
		}

		fmt.Printf("Question: %+v\n", q)

		fmt.Printf("Found question for name '%s'\n", q.Name.String())

		if q.Name.String() == "." && q.Type.String() == "TypeNS" && q.Class.String() == "ClassINET" {
			nsRecords, err := resolver(A_ROOT_SERVER, q.Name.String())
			if err != nil {
				panic(err)
			}
			for _, v := range nsRecords {
				fmt.Printf("nsrecords: %+v", v)
			}
		}
		if err := p.SkipAllQuestions(); err != nil {
			panic(err)
		}
		break
	}
	fmt.Printf("Done!\n")
}
func resolver(addr, toResolve string) (addrs []*net.NS, err error) {
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
