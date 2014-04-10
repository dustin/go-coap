package main

import (
	"fmt"
	"log"
	"time"

	"github.com/dustin/go-coap"
)

func periodicTransmitter(r *coap.RemoteAddr, m *coap.Message) {
	subded := time.Now()

	for {
		m.Type = coap.Acknowledgement
		m.Code = coap.Content
		m.Payload = []byte(fmt.Sprintf("Been running for %v", time.Since(subded)))

		log.Printf("Transmitting %v", r)
		err := coap.Transmit(r, *m)
		if err != nil {
			log.Printf("Error on transmitter, stopping: %v", err)
		}

		time.Sleep(time.Second)
	}
}

func main() {
	coap.HandleFunc("/uptime", func(r *coap.RemoteAddr, m *coap.Message) *coap.Message {
		log.Printf("Got message path=%q: %#v from %v", m.Path(), m, r)
		if m.Code == coap.SUBSCRIBE {
			go periodicTransmitter(r, m)
		}
		return nil
	})
	log.Fatal(coap.ListenAndServe(":5683", nil))
}
