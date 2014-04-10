package main

import (
	"log"

	"github.com/dustin/go-coap"
)

func main() {

	req := coap.Message{
		Type:      coap.Confirmable,
		Code:      coap.GET,
		MessageID: 0,
	}

	req.SetOption(coap.Observe, 0)
	req.SetPathString("/time")

	c, err := coap.Dial("udp", "localhost:5683")
	if err != nil {
		log.Fatalf("Error dialing: %v", err)
	}

	rv, err := c.Send(req)
	if err != nil {
		log.Fatalf("Error sending request: %v", err)
	}

	for err == nil {
		if rv != nil {
			if err != nil {
				log.Fatalf("Error receiving: %v", err)
			}
			if string(rv.Payload) != "" {
				log.Printf("Got %s", rv.Payload)
			}

			req = coap.Message{
				Type:      coap.Acknowledgement,
				Code:      0,
				MessageID: rv.MessageID,
			}
			c.Send(req)
		}
		rv, err = c.Receive()
	}
	log.Printf("Done...\n")

}
