// CoAP Client and Server in Go
package coap

import (
	"log"
	"net"
)

const maxPktLen = 1500

func handlePacket(l *net.UDPConn, data []byte, u *net.UDPAddr) {
	log.Printf("Got %v", data)
}

func ListenAndServe(n, addr string) error {
	uaddr, err := net.ResolveUDPAddr(n, addr)
	if err != nil {
		return err
	}

	l, err := net.ListenUDP(n, uaddr)
	if err != nil {
		return err
	}

	for {
		buf := make([]byte, maxPktLen)

		nr, addr, err := l.ReadFromUDP(buf)
		if err != nil {
			go handlePacket(l, buf[:nr], addr)
		}
	}

	panic("Unreachable")
}
