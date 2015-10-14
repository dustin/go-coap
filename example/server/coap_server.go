package main

import (
	"log"
	"net"

	"github.com/dustin/go-coap"
)

func main() {
	log.Fatal(coap.ListenAndServe("udp", ":5683",
		coap.FuncHandler(func(l *net.UDPConn, a *net.UDPAddr, m *coap.Message) *coap.Message {
			log.Printf("Got message path=%q: %#v from %v", m.Path(), m, a)
			if m.IsConfirmable() {
				res := &coap.Message{
					Type:      coap.Acknowledgement,
					Code:      coap.Content,
					MessageID: m.MessageID,
					Token:     m.Token,
					Payload:   []byte("hello to you!"),
				}
				res.SetOption(coap.ContentFormat, coap.TextPlain)

				log.Printf("Transmitting %#v", res)
				return res
			}
			return nil
		})))
}
