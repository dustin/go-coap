package main

import (
	"log"
	"net"

	"github.com/dustin/go-coap"
)

func main() {
	log.Fatal(coap.ListenAndServe("udp", ":5683",
		coap.FuncHandler(func(l *net.UDPConn, a *net.UDPAddr, m coap.Message) *coap.Message {
			log.Printf("Got message path=%q: %#v from %v", m.Path(), m, a)
			if m.IsConfirmable() {
				res := &coap.Message{
					Type:      coap.Acknowledgement,
					Code:      coap.Content,
					MessageID: m.MessageID,
					Payload:   []byte("hello to you!"),
				}
				res.SetOption(coap.ContentType, coap.TextPlain)
				res.SetOption(coap.LocationPath, m.Path())

				return res
			}
			return nil
		})))
}
