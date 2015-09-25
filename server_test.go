package coap

import (
	"bytes"
	"net"
	"testing"
)

// compareMessages compares the e(xptected) message to the a(ctual) message and
// reports diffs with t.Errorf.
func assertEqualMessages(t *testing.T, e, a Message) {
	if e.Type != a.Type {
		t.Errorf("Expected type %v, got %v", e.Type, a.Type)
	}
	if e.Code != a.Code {
		t.Errorf("Expected code %v, got %v", e.Code, a.Code)
	}
	if e.MessageID != a.MessageID {
		t.Errorf("Expected MessageID %v, got %v", e.MessageID, a.MessageID)
	}
	if !bytes.Equal(e.Token, a.Token) {
		t.Errorf("Expected token %#v, got %#v", e.Token, a.Token)
	}
	if !bytes.Equal(e.Payload, a.Payload) {
		t.Errorf("Expected payload %#v, got %#v", e.Payload, a.Payload)
	}

	optionsEqual := true
	if len(e.opts) != len(a.opts) {
		optionsEqual = false
	} else {
		for i, _ := range e.opts {
			if e.opts[i] != a.opts[i] {
				optionsEqual = false
			}
		}
	}
	if !optionsEqual {
		t.Errorf("Expected options %#v, got %#v", e.opts, a.opts)
	}
}

func startUDPLisenter(t *testing.T) (*net.UDPConn, string) {
	udpAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal("Can't resolve UDP addr")
	}
	udpListener, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		t.Fatal("Can't listen on UDP")
	}
	coapServerAddr := udpListener.LocalAddr().String()
	return udpListener, coapServerAddr
}

func dialAndSend(t *testing.T, addr string, req Message) *Message {
	c, err := Dial("udp", addr)
	if err != nil {
		t.Fatalf("Error dialing: %v", err)
	}
	m, err := c.Send(req)
	if err != nil {
		t.Fatalf("Error sending request: %v", err)
	}
	return m
}

func TestServeWithAckResponse(t *testing.T) {
	req := Message{
		Type:      Confirmable,
		Code:      POST,
		MessageID: 9876,
		Payload:   []byte("Content sent by client"),
	}
	req.SetOption(ContentFormat, TextPlain)
	req.SetPathString("/req/path")

	res := Message{
		Type:      Acknowledgement,
		Code:      Content,
		MessageID: req.MessageID,
		Payload:   []byte("Reply from CoAP server"),
	}
	res.SetOption(ContentFormat, TextPlain)
	res.SetPath(req.Path())

	handler := FuncHandler(func(l *net.UDPConn, a *net.UDPAddr, m *Message) *Message {
		assertEqualMessages(t, req, *m)
		return &res
	})

	udpListener, coapServerAddr := startUDPLisenter(t)
	defer udpListener.Close()
	go Serve(udpListener, handler)

	m := dialAndSend(t, coapServerAddr, req)
	if m == nil {
		t.Fatalf("Didn't receive CoAP response")
	}
	assertEqualMessages(t, res, *m)
}

func TestServeWithoutAckResponse(t *testing.T) {
	req := Message{
		Type:      NonConfirmable,
		Code:      POST,
		MessageID: 54321,
		Payload:   []byte("Content sent by client"),
	}
	req.SetOption(ContentFormat, AppOctets)

	handler := FuncHandler(func(l *net.UDPConn, a *net.UDPAddr, m *Message) *Message {
		assertEqualMessages(t, req, *m)
		return nil
	})

	udpListener, coapServerAddr := startUDPLisenter(t)
	defer udpListener.Close()
	go Serve(udpListener, handler)

	m := dialAndSend(t, coapServerAddr, req)
	if m != nil {
		t.Fatalf("Received response packet, but expected none")
	}
}
