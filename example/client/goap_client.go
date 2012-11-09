package main

import (
	"log"

	"github.com/dustin/go-coap"
)

func main() {

	req := coap.Message{
		Type:      coap.Confirmable,
		Code:      coap.GET,
		MessageID: 12345,
		Options: coap.Options{
			{coap.ETag, []byte("weetag")},
			{coap.MaxAge, []byte{0, 0, 0, 3}},
		},
		Payload: []byte("hello, world!"),
	}

	req.SetPathString("/some/path")

	c, err := coap.Dial("udp", "localhost:5683")
	if err != nil {
		log.Fatalf("Error dialing: %v", err)
	}

	rv, err := c.Send(req)
	if err != nil {
		log.Fatalf("Error sending request: %v", err)
	}

	if rv != nil {
		log.Printf("Response payload: %s", rv.Payload)
	}

}
