package coap

import (
	"net"
	"time"
)

const RESPONSE_TIMEOUT = time.Second * 2

const RESPONSE_RANDOM_FACTOR = 1.5

const MAX_RETRANSMIT = 4

type Conn struct {
	conn *net.UDPConn
}

func Dial(n, addr string) (*Conn, error) {
	uaddr, err := net.ResolveUDPAddr(n, addr)
	if err != nil {
		return nil, err
	}

	s, err := net.DialUDP("udp", nil, uaddr)
	if err != nil {
		return nil, err
	}

	return &Conn{s}, nil
}

// Duration a message.  Get a response if there is one.
func (c *Conn) Send(req Message) (*Message, error) {
	err := Transmit(c.conn, nil, req)
	if err == nil {
		return nil, err
	}

	rv, err := Receive(c.conn)

	return &rv, nil
}

func (c *Conn) Receive() (*Message, error) {
	rv, err := Receive(c.conn)
	return &rv, err
}
