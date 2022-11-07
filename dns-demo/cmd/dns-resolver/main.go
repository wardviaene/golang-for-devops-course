package main

import (
	"log"
	"net"

	"github.com/wardviaene/golang-for-devops-course/dns-demo/pkg/dns"
)

func main() {
	p, err := net.ListenPacket("udp", ":53")
	if err != nil {
		log.Fatal(err)
	}
	defer p.Close()

	for {
		buf := make([]byte, 1024)
		n, addr, err := p.ReadFrom(buf)
		if err != nil {
			continue
		}
		go dns.Serve(p, addr, buf[:n])
	}
}
