package main

import (
	"log"

	"github.com/dustin/go-coap"
)

func main() {
	coap.HandleFunc("/hello", func(r *coap.RemoteAddr, m *coap.Message) *coap.Message {
		log.Printf("Got message path=%q: %#v from %v", m.Path(), m, r)
		if m.IsConfirmable() {
			res := &coap.Message{
				Type:      coap.Acknowledgement,
				Code:      coap.Content,
				MessageID: m.MessageID,
				Payload:   []byte("hello to you!"),
			}
			log.Printf("%v", res)
			return res
		}
		return nil
	})

	log.Fatal(coap.ListenAndServe(":5683", nil))
}
