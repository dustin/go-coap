package coap

import (
	"errors"
	"log"
	"math/rand"
	"net"
)

var nextMID uint16 = uint16(rand.Int())

// An incomming request waiting to be responded
type Request interface {
	Message() *Message
	Addr() net.Addr

	// Ack sends a separate acknowledgement from the response. To be used if the request will take some times for avoiding unattended retransmission from the sender.
	Ack() error

	// Respond with a content
	Respond(code COAPCode, content []byte, options map[OptionID]interface{}) error
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

// UdpListenAndServer listen and server CoAP resources using the given RequestHandler
func UdpListenAndServe(n, addr string, rh RequestHandler) error {

	uaddr, err := net.ResolveUDPAddr(n, addr)
	if err != nil {
		return err
	}

	l, err := net.ListenUDP(n, uaddr)
	if err != nil {
		return err
	}

	defer l.Close()

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
	// if it's not an confirmable message or it was already acked
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

func (rq *UDPRequest) Respond(code COAPCode, content []byte, options map[OptionID]interface{}) error {
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

	for k, v := range options {
		msg.AddOption(k, v)
	}
	return rq.retrans.Record(msg, rq.addr)
}

func TcpListenAndServe(n, addr string, rh RequestHandler) error {

	taddr, err := net.ResolveTCPAddr(n, addr)
	if err != nil {
		return err
	}

	l, err := net.ListenTCP(n, taddr)
	if err != nil {
		return err
	}

	defer l.Close()

	for {
		cnx, err := l.Accept()
		if err != nil {
			return err
		}

		go func() {
			buf := make([]byte, maxPktLen)

			for {
				nr, err := cnx.Read(buf)
				if err != nil {
					log.Print(err)
				}
				var msg TCPMessage
				err = msg.UnmarshalBinary(buf[:nr])
				if err == nil {
					log.Println(err)
					break
				}
				handleTCPRequest(&msg, cnx, rh)
			}
			cnx.Close()

		}()
	}

	return nil
}

func handleTCPRequest(msg *TCPMessage, s net.Conn, rh RequestHandler) {
	rq := TCPRequest{
		msg: msg,
		s:   s,
	}

	rh.Handle(&rq)
}

type TCPRequest struct {
	msg *TCPMessage
	s   net.Conn
}

func (rq *TCPRequest) Message() *Message {
	return &rq.msg.Message
}

func (rq *TCPRequest) Addr() net.Addr {
	return rq.s.RemoteAddr()
}

// Ack is not used for TCP CoAP because TCP is reliable
func (rq *TCPRequest) Ack() error {
	return nil
}

func (rq *TCPRequest) Respond(code COAPCode, content []byte, options map[OptionID]interface{}) error {
	var msg TCPMessage
	// piggybacked answer
	msg = TCPMessage{
		Message: Message{
			Type:      Acknowledgement,
			Code:      code,
			MessageID: 0,
			Payload:   content,
			Token:     rq.msg.Token,
		},
	}

	for k, v := range options {
		msg.AddOption(k, v)
	}

	// encode and send
	bin, err := msg.MarshalBinary()

	if err != nil {
		return err
	}

	nb, err := rq.s.Write(bin)

	if err != nil {
		return err
	}

	if nb != len(bin) {
		return errors.New("didn't totaly write the TCP message")
	}

	return nil
}
