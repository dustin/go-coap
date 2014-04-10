package main

import (
	"log"

	"github.com/dustin/go-coap"
)

func main() {

	req := coap.Message{
		Type:      coap.NonConfirmable,
		Code:      coap.SUBSCRIBE,
		MessageID: 12345,
	}

	req.SetPathString("/uptime")

	c, err := coap.Dial("udp", "localhost:5683")
	if err != nil {
		log.Fatalf("Error dialing: %v", err)
	}

	for {
		rv, err := c.Send(req)
		if err != nil {
			log.Fatalf("Error sending request: %v", err)
		}
		for {

			if rv != nil {
				if err != nil {
					log.Fatalf("Error receiving: %v", err)
				}
				log.Printf("Got %s", rv.Payload)
			}
			rv, err = c.Receive()

		}
	}
	log.Printf("Done...\n")

}
