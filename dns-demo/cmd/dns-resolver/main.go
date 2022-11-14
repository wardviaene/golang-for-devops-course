package main

import (
	"fmt"
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
		buf := make([]byte, 512)
		n, addr, err := p.ReadFrom(buf)
		if err != nil {
			fmt.Printf("Connection error [%s]: %s\n", addr.String(), err)
			continue
		}
		go dns.HandlePacket(p, addr, buf[:n])
	}
}
