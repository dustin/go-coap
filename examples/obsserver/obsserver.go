package main

import (
	"log"
	"time"

	"github.com/dustin/go-coap"
)

func observeTime() {
	for {
		// MessageID is managed by service hub
		msg := &coap.Message{
			Type:    coap.Acknowledgement,
			Code:    coap.Content,
			Payload: []byte(time.Now().Format(time.Stamp)),
		}

		msg.SetOption(coap.ContentFormat, coap.TextPlain)
		msg.SetOption(coap.MaxAge, 1)

		log.Printf("%s ** broadcasting %v", "", msg)
		coap.Notify("/time", msg)
		time.Sleep(time.Second * 2)
	}
}

func handleTime(r *coap.RemoteAddr, m *coap.Message) *coap.Message {
	log.Printf("%s got message: %#v from %v", m.Path(), m, r)
	switch m.Code {
	case coap.GET:
		msg := &coap.Message{
			Type:      coap.Acknowledgement,
			Code:      coap.Content,
			MessageID: m.MessageID,
			Payload:   []byte(time.Now().Format(time.Stamp)),
		}
		msg.SetOption(coap.ContentFormat, coap.TextPlain)
		msg.SetOption(coap.MaxAge, 1)

		log.Printf("%s ** transmitting %v", m.Path(), msg)
		return msg
	}
	return nil
}

func main() {
	// Start observer resource (async)
	go observeTime()
	// Handle /time
	coap.HandleFunc("/time", handleTime)
	// Start coap listener
	log.Fatal(coap.ListenAndServe(":5683", nil))
}
