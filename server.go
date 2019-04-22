// Package coap provides a CoAP client and server.
package coap

import (
	"log"
	"math"
	"net"
	"time"
)

const maxPktLen = 1500

// Handler is a type that handles CoAP messages.
type Handler interface {
	// Handle the message and optionally return a response message.
	ServeCOAP(l *net.UDPConn, a *net.UDPAddr, m *Message) *Message
}

type funcHandler func(l *net.UDPConn, a *net.UDPAddr, m *Message) *Message

func (f funcHandler) ServeCOAP(l *net.UDPConn, a *net.UDPAddr, m *Message) *Message {
	return f(l, a, m)
}

// FuncHandler builds a handler from a function.
func FuncHandler(f func(l *net.UDPConn, a *net.UDPAddr, m *Message) *Message) Handler {
	return funcHandler(f)
}

func handlePacket(l *net.UDPConn, data []byte, u *net.UDPAddr,
	rh Handler) {

	msg, err := ParseMessage(data)
	if err != nil {
		log.Printf("Error parsing %v", err)
		return
	}

	if msg.Block2 == nil {
		msg.Block2 = &Block{
			Num:  0,
			More: false,
			Size: 1024,
		}
	}

	rv := rh.ServeCOAP(l, u, &msg)

	if rv != nil {
		header, _ := rv.MarshalBinary()

		size2 := rv.Option(Size2)

		if size2 != nil && len(header)+int(size2.(uint32))+1 > maxPktLen {
			count := math.Floor(float64(size2.(uint32)) / float64(msg.Block2.Size))

			var more bool
			if float64(msg.Block2.Num) < count {
				more = true
			} else if float64(msg.Block2.Num) == count {
				more = false
			}

			rv.Block2 = &Block{
				Num:  msg.Block2.Num,
				More: more,
				Size: msg.Block2.Size,
			}

			rv.AddOption(Block2, rv.Block2.MarshalBinary())
		}

		Transmit(l, u, *rv)
	}
}

// Transmit a message.
func Transmit(l *net.UDPConn, a *net.UDPAddr, m Message) error {
	d, err := m.MarshalBinary()
	if err != nil {
		return err
	}

	if a == nil {
		_, err = l.Write(d)
	} else {
		_, err = l.WriteTo(d, a)
	}
	return err
}

// Receive a message.
func Receive(l *net.UDPConn, buf []byte) (Message, error) {
	l.SetReadDeadline(time.Now().Add(ResponseTimeout))

	nr, _, err := l.ReadFromUDP(buf)
	if err != nil {
		return Message{}, err
	}
	return ParseMessage(buf[:nr])
}

// ListenAndServe binds to the given address and serve requests forever.
func ListenAndServe(n, addr string, rh Handler) error {
	uaddr, err := net.ResolveUDPAddr(n, addr)
	if err != nil {
		return err
	}

	l, err := net.ListenUDP(n, uaddr)
	if err != nil {
		return err
	}

	return Serve(l, rh)
}

// Serve processes incoming UDP packets on the given listener, and processes
// these requests forever (or until the listener is closed).
func Serve(listener *net.UDPConn, rh Handler) error {
	buf := make([]byte, maxPktLen)
	for {
		nr, addr, err := listener.ReadFromUDP(buf)
		if err != nil {
			if neterr, ok := err.(net.Error); ok && (neterr.Temporary() || neterr.Timeout()) {
				time.Sleep(5 * time.Millisecond)
				continue
			}
			return err
		}
		tmp := make([]byte, nr)
		copy(tmp, buf)
		go handlePacket(listener, tmp, addr, rh)
	}
}
