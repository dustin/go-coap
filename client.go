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
	d, err := encodeMessage(req)
	if err != nil {
		return nil, err
	}

	_, err = c.conn.Write(d)
	if err != nil {
		return nil, err
	}

	c.conn.SetReadDeadline(time.Now().Add(RESPONSE_TIMEOUT))

	data := make([]byte, maxPktLen)
	nr, _, err := c.conn.ReadFromUDP(data)
	if err != nil {
		return nil, err
	}
	rv, err := parseMessage(data[:nr])
	if err != nil {
		return nil, err
	}

	return &rv, nil
}
