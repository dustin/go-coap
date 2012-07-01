// CoAP Client and Server in Go
package coap

import (
	"log"
	"net"
)

const maxPktLen = 1500

type RequestHandler interface {
	Handle(a *net.UDPAddr, m Message) *Message
}

type funcHandler func(a *net.UDPAddr, m Message) *Message

func (f funcHandler) Handle(a *net.UDPAddr, m Message) *Message {
	return f(a, m)
}

func FuncHandler(f func(a *net.UDPAddr, m Message) *Message) RequestHandler {
	return funcHandler(f)
}

func handlePacket(l *net.UDPConn, data []byte, u *net.UDPAddr,
	rh RequestHandler) {

	msg, err := parseMessage(data)
	if err != nil {
		log.Printf("Error parsing %v", err)
		return
	}

	rv := rh.Handle(u, msg)
	if rv != nil {
		b, err := encodeMessage(*rv)
		if err != nil {
			log.Printf("Error encoding %#v", msg)
			return
		}
		l.WriteTo(b, u)
	}
}

func ListenAndServe(n, addr string, rh RequestHandler) error {
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
		if err == nil {
			go handlePacket(l, buf[:nr], addr, rh)
		}
	}

	panic("Unreachable")
}
