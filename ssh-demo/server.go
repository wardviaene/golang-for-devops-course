package ssh

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"strings"

	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
)

func StartServer(authorizedKeys []byte) error {
	authorizedKeysMap := map[string]bool{}
	for len(authorizedKeys) > 0 {
		pubKey, _, _, rest, err := ssh.ParseAuthorizedKey(authorizedKeys)
		if err != nil {
			return err
		}
		authorizedKeysMap[string(pubKey.Marshal())] = true
		authorizedKeys = rest
	}
	config := &ssh.ServerConfig{
		PublicKeyCallback: func(c ssh.ConnMetadata, pubKey ssh.PublicKey) (*ssh.Permissions, error) {
			if authorizedKeysMap[string(pubKey.Marshal())] {
				return &ssh.Permissions{
					// Record the public key used for authentication.
					Extensions: map[string]string{
						"pubkey-fp": ssh.FingerprintSHA256(pubKey),
					},
				}, nil
			}
			return nil, fmt.Errorf("unknown public key for %q", c.User())
		},
	}

	privateBytes, err := ioutil.ReadFile("server.pem")
	if err != nil {
		return err
	}

	private, err := ssh.ParsePrivateKey(privateBytes)
	if err != nil {
		return err
	}

	config.AddHostKey(private)

	listener, err := net.Listen("tcp", "0.0.0.0:2022")
	if err != nil {
		return fmt.Errorf("failed to listen for connection: %s", err)
	}

	for {
		fmt.Printf("Accepting new connections\n")
		nConn, err := listener.Accept()
		if err != nil {
			fmt.Printf("failed to accept incoming connection: %s\n", err)
			continue
		}

		conn, chans, reqs, err := ssh.NewServerConn(nConn, config)
		if err != nil {
			fmt.Printf("failed to handshake: %s", err)
			continue
		}
		log.Printf("logged in with key %s", conn.Permissions.Extensions["pubkey-fp"])

		go ssh.DiscardRequests(reqs)

		go handleConnection(conn, chans)
	}

	//return nil
}

func handleConnection(conn *ssh.ServerConn, chans <-chan ssh.NewChannel) {
	for newChannel := range chans {
		// Channels have a type, depending on the application level
		// protocol intended. In the case of a shell, the type is
		// "session" and ServerShell may be used to present a simple
		// terminal interface.
		if newChannel.ChannelType() != "session" {
			newChannel.Reject(ssh.UnknownChannelType, "unknown channel type")
			continue
		}
		channel, requests, err := newChannel.Accept()
		if err != nil {
			log.Fatalf("Could not accept channel: %v", err)
		}

		// Sessions have out-of-band requests such as "shell",
		// "pty-req" and "env".  Here we handle only the
		// "shell" request.
		go func(in <-chan *ssh.Request) {
			for req := range in {
				fmt.Printf("ssh client requested: %s\n", req.Type)
				switch req.Type {
				case "exec":
					out := []byte(execSomething(conn, string(bytes.TrimPrefix(req.Payload, []byte{0, 0, 0, 6}))))
					channel.Write(out)
					channel.SendRequest("exit-status", false, []byte{0, 0, 0, 0})
					req.Reply(true, nil)
					channel.Close()
				case "pty-req":
					startTerminal(conn, channel)
					req.Reply(true, nil)
				case "shell", "env":
					req.Reply(true, nil)
				default:
				}
			}
		}(requests)

	}
}
func startTerminal(conn *ssh.ServerConn, channel ssh.Channel) {
	termInstance := term.NewTerminal(channel, "> ")

	go func() {
		defer channel.Close()
		for {
			line, err := termInstance.ReadLine()
			if err != nil {
				break
			}
			fmt.Println(line)
			switch line {
			case "quit":
				termInstance.Write([]byte("Goodbye\n"))
				channel.Close()
			case "whoami":
				termInstance.Write(execSomething(conn, "whoami"))
				channel.SendRequest("exit-status", false, []byte{0, 0, 0, 0})
			case "":
			default:
				termInstance.Write([]byte("Command not recognized\n"))
			}
		}
	}()
}

func execSomething(conn *ssh.ServerConn, payload string) []byte {
	switch strings.TrimRight(strings.TrimSpace(string(payload)), "\n") {
	case "whoami":
		fmt.Printf("whoami executed")
		return []byte("you are ssh user: " + conn.Conn.User() + "\n")
	default:
		fmt.Printf("Unrecognized command: '%v'\n", []byte(payload))
	}
	return nil
}
