// CoAP Client and Server in Go
package coap

import (
	"log"
	"net"
	"time"
)

const maxPktLen = 1500

type RemoteAddr struct {
	*net.UDPAddr
}

// Handler is a type that handles CoAP messages.
type Handler interface {
	// Handle the message and optionally return a response message.
	ServeCOAP(r *RemoteAddr, m *Message) *Message
}

type Server struct {
	conn           *net.UDPConn
	Addr           string        // UDP address to listen on, ":coap" if empty
	Handler        Handler       // handler to invoke, coap.ServeMux if nil
	ReadTimeout    time.Duration // maximum duration before timing out read of the request
	WriteTimeout   time.Duration // maximum duration before timing out write of the response
	MaxHeaderBytes int           // maximum size of request headers, DefaultMaxHeaderBytes if 0
}

var DefaultServer = new(Server)

type HandlerFunc func(r *RemoteAddr, m *Message) *Message

func (f HandlerFunc) ServeCOAP(r *RemoteAddr, m *Message) *Message {
	return f(r, m)
}

func HandleFunc(pattern string, handler func(r *RemoteAddr, m *Message) *Message) {
	DefaultServeMux.HandleFunc(pattern, handler)
}

func Handle(pattern string, handler Handler) { DefaultServeMux.Handle(pattern, handler) }

func handlePacket(data []byte, l *net.UDPConn, a *net.UDPAddr) {
	msg, err := parseMessage(data)

	if err != nil {
		debugMsg("Error parsing %v", err)
		return
	}
	/*
		  if msg.Type == Reset {
				return
			}
	*/
	if msg.Type == Acknowledgement {
		h.ack <- map[string]uint16{a.String(): msg.MessageID}
		return
	}

	if msg.IsObserver() {
		o := &observer{addr: &RemoteAddr{a},
			send: make(chan *Message),
			ack:  make(chan uint16),
			rsc:  msg.PathString()}
		// adds new observer
		h.register <- o
		go o.transmitter()
	}

	rh, _ := DefaultServeMux.match(msg.PathString())
	if rh != nil {
		rv := rh.ServeCOAP(&RemoteAddr{a}, &msg)
		if rv != nil {
			transmit(l, a, *rv)
		}
	} else if msg.IsConfirmable() {
		// resource not found
		transmit(l, a, Message{
			Type: Acknowledgement,
			Code: NotFound,
		})
	}
}

func Transmit(r *RemoteAddr, m Message) error {
	return transmit(DefaultServer.conn, r.UDPAddr, m)
}

func transmit(c *net.UDPConn, a *net.UDPAddr, m Message) error {
	d, err := m.encode()
	if err != nil {
		return err
	}
	if a == nil {
		_, err = c.Write(d)
	} else {
		_, err = c.WriteTo(d, a)
	}
	return err
}

// Receive a message with timeout.
func ReceiveTimeout(l *net.UDPConn, rt time.Duration, buf []byte) (Message, error) {
	l.SetReadDeadline(time.Now().Add(rt))

	nr, _, err := l.ReadFromUDP(buf)
	if err != nil {
		return Message{}, err
	}
	m, err := parseMessage(buf[:nr])
	return m, err
}

// Receive a message.
func Receive(l *net.UDPConn, buf []byte) (Message, error) {
	nr, _, err := l.ReadFromUDP(buf)
	if err != nil {
		return Message{}, err
	}
	m, err := parseMessage(buf[:nr])
	return m, err
}

func Serve(l *net.UDPConn, handler Handler) error {
	DefaultServer = &Server{conn: l, Handler: handler}
	return DefaultServer.Serve(l)
}

func (srv *Server) ListenAndServe() error {
	addr := srv.Addr
	if addr == "" {
		addr = ":5683"
	}

	_addr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return err
	}

	l, e := net.ListenUDP("udp", _addr)
	if e != nil {
		return e
	}
	DefaultServer.conn = l
	go h.run()
	return srv.Serve(l)
}

// Bind to the given address and serve requests forever.
func ListenAndServe(addr string, handler Handler) error {
	DefaultServer = &Server{Addr: addr, Handler: handler}
	return DefaultServer.ListenAndServe()
}

func (srv *Server) Serve(l *net.UDPConn) error {
	defer l.Close()
	var tempDelay time.Duration // how long to sleep on accept failure
	buf := make([]byte, maxPktLen)
	for {
		nr, addr, err := l.ReadFromUDP(buf)
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}
				if max := 1 * time.Second; tempDelay > max {
					tempDelay = max
				}
				debugMsg("coap: Accept error: %v; retrying in %v", err, tempDelay)
				time.Sleep(tempDelay)
				continue
			}
			return err
		}
		tempDelay = 0
		tmp := make([]byte, nr)
		copy(tmp, buf)
		go handlePacket(tmp, l, addr)
	}
}

// Notify observers of resource.
func Notify(resource string, m *Message) {
	if m != nil {
		for o := range h.observers[resource[1:]] {
			select {
			case o.send <- m:
			}
		}
	}
}

// Print debug messages
func debugMsg(m string, v ...interface{}) {
	if Verbose {
		log.Printf(m, v...)
	}
}
