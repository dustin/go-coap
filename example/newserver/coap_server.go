package main

import (
	"log"

	"github.com/dustin/go-coap"
)

func main() {

	log.Fatal(coap.UdpListenAndServe("udp", ":5683",
		coap.FuncRqHandler(func(rq coap.Request) error {
			log.Printf("Got message path=%q: %#v from %v", rq.Message().Path(), rq.Message(), rq.Addr())
			rq.Respond(coap.Content, []byte("Yeah!"), nil)
			return nil
		})))
}
