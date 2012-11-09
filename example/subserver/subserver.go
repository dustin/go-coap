package main

import (
	"fmt"
	"log"
	"net"
	"time"

	"github.com/dustin/go-coap"
)

func periodicTransmitter(l *net.UDPConn, a *net.UDPAddr, m coap.Message) {
	subded := time.Now()

	for {
		msg := coap.Message{
			Type:      coap.Acknowledgement,
			Code:      coap.Content,
			MessageID: m.MessageID,
			Options: coap.Options{
				{coap.ContentType, []byte{byte(coap.TextPlain)}},
				{coap.LocationPath, m.Path()},
			},
			Payload: []byte(fmt.Sprintf("Been running for %v", time.Since(subded))),
		}

		log.Printf("Transmitting %v", msg)
		err := coap.Transmit(l, a, msg)
		if err != nil {
			log.Printf("Error on transmitter, stopping: %v", err)
			return
		}

		time.Sleep(time.Second)
	}
}

func main() {
	log.Fatal(coap.ListenAndServe("udp", ":5683",
		coap.FuncHandler(func(l *net.UDPConn, a *net.UDPAddr, m coap.Message) *coap.Message {
			log.Printf("Got message path=%q: %#v from %v", m.Path(), m, a)
			if m.Code == coap.SUBSCRIBE {
				go periodicTransmitter(l, a, m)
			}
			return nil
		})))
}
