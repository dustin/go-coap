package coap

import (
	"net"
	"time"
)

const RESPONSE_TIMEOUT = time.Second * 2

const RESPONSE_RANDOM_FACTOR = 1.5

const MAX_RETRANSMIT = 4

// A CoAP client connection.
type Conn struct {
	conn *net.UDPConn
	buf  []byte
}

// Get a CoAP client.
func Dial(n, addr string) (*Conn, error) {
	uaddr, err := net.ResolveUDPAddr(n, addr)
	if err != nil {
		return nil, err
	}

	s, err := net.DialUDP("udp", nil, uaddr)
	if err != nil {
		return nil, err
	}

	return &Conn{s, make([]byte, maxPktLen)}, nil
}

// Duration a message.  Get a response if there is one.
func (c *Conn) Send(req Message) (*Message, error) {
	err := Transmit(c.conn, nil, req)
	if err != nil {
		return nil, err
	}

	if !req.IsConfirmable() {
		return nil, nil
	}

	rv, err := Receive(c.conn, c.buf)

	return &rv, nil
}

// Receive a message.
func (c *Conn) Receive() (*Message, error) {
	rv, err := Receive(c.conn, c.buf)
	return &rv, err
}
