package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"syscall"
)

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
		// copy what client is sending to server to STDOUT
		w := io.MultiWriter(r, os.Stdout)
		io.Copy(w, c)
	}()

	io.Copy(c, r)
}
