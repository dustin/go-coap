package coap

import (
	"log"
	"math/rand"
	"net"
)

var nextMID uint16 = uint16(rand.Int())

type Request interface {
	Message() *Message
	Addr() net.Addr
	Ack() error

	RespondCode(code COAPCode) error

	Respond(code COAPCode, content []byte, contentType *MediaType) error
}

type RequestHandler interface {
	// Handle the message and optionally return a response message.
	Handle(rq Request) error
}
type funcRqHandler func(rq Request) error

func (f funcRqHandler) Handle(rq Request) error {
	return f(rq)
}

// FuncRqHandler builds a handler from a function.
func FuncRqHandler(f func(rq Request) error) RequestHandler {
	return funcRqHandler(f)
}

func UdpListenAndServe(n, addr string, rh RequestHandler) error {

	uaddr, err := net.ResolveUDPAddr(n, addr)
	if err != nil {
		return err
	}

	l, err := net.ListenUDP(n, uaddr)
	if err != nil {
		return err
	}

	// init retransmission facilities
	retrans := NewRetransmitter(l)

	buf := make([]byte, maxPktLen)
	for {
		nr, addr, err := l.ReadFromUDP(buf)
		if err == nil {
			msg, err := parseMessage(buf[:nr])
			if err != nil {
				log.Printf("Error parsing %v", err)
			}
			retrans.Received(&msg, addr)
			go handleRequest(&msg, addr, l, rh, retrans)
		}
	}

	return nil
}

func handleRequest(msg *Message, addr *net.UDPAddr, s *net.UDPConn, rh RequestHandler, retrans *Retransmitter) {
	rq := UDPRequest{
		msg:     msg,
		addr:    addr,
		s:       s,
		acked:   false,
		retrans: retrans,
	}

	rh.Handle(&rq)
}

type UDPRequest struct {
	msg     *Message
	addr    *net.UDPAddr
	s       *net.UDPConn
	acked   bool
	retrans *Retransmitter
}

func (rq *UDPRequest) Message() *Message {
	return rq.msg
}

func (rq *UDPRequest) Addr() net.Addr {
	return rq.addr
}

func (rq *UDPRequest) Ack() error {
	// if it's not an ackable message or it was already acked
	// just do nothing silently
	if rq.msg.Type != Confirmable || rq.acked {
		return nil
	}

	ackMsg := Message{
		Type:      Acknowledgement,
		Code:      0,
		MessageID: rq.msg.MessageID,
		Payload:   nil,
	}

	if err := Transmit(rq.s, rq.addr, ackMsg); err != nil {
		return err
	}
	rq.acked = true
	return nil
}
func (rq *UDPRequest) RespondCode(code COAPCode) error {
	return rq.Respond(code, []byte{}, nil)

}

func (rq *UDPRequest) Respond(code COAPCode, content []byte, contentType *MediaType) error {
	var msg Message

	if rq.acked {
		// answer after ack
		// generate a new MID
		msg = Message{
			Type:      NonConfirmable,
			Code:      code,
			MessageID: nextMID,
			Payload:   content,
			Token:     rq.msg.Token,
		}
		nextMID = nextMID + 1
	} else {
		// piggybacked answer
		msg = Message{
			Type:      Acknowledgement,
			Code:      code,
			MessageID: rq.msg.MessageID,
			Payload:   content,
			Token:     rq.msg.Token,
		}
	}

	return rq.retrans.Record(msg, rq.addr)
}
