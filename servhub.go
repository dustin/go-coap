package coap

import (
	"math/rand"
	"time"
)

// Observer of some resource
type observer struct {
	tid  uint16
	rsc  string
	addr *RemoteAddr
	send chan *Message
	ack  chan uint16
}

// Internal hub
type hub struct {
	ack        chan map[string]uint16
	register   chan *observer
	unregister chan *observer
	observers  map[string]map[*observer]bool
}

var h = hub{
	ack:        make(chan map[string]uint16),
	register:   make(chan *observer),
	unregister: make(chan *observer),
	observers:  make(map[string]map[*observer]bool),
}

// Observer transmitter.
func (o *observer) transmitter() {
	/*
			   Implementation Note:  Several implementation strategies can be
			        employed for generating Message IDs.  In the simplest case a CoAP
			        endpoint generates Message IDs by keeping a single Message ID
			        variable, which is changed each time a new Confirmable or Non-
			        confirmable message is sent regardless of the destination address
			        or port.  Endpoints dealing with large numbers of transactions
			        could keep multiple Message ID variables, for example per prefix
			        or destination address (note that some receiving endpoints may not
			        be able to distinguish unicast and multicast packets addressed to
			        it, so endpoints generating Message IDs need to make sure these do
			        not overlap).  It is strongly recommended that the initial value
			        of the variable (e.g., on startup) be randomized, in order to make
			        successful off-path attacks on the protocol less likely.
		     http://tools.ietf.org/html/draft-ietf-core-coap-18#section-4.4
	*/
	o.tid = uint16(rand.Intn(65536))
	for msg := range o.send {
		o.tid++
		go func(m Message, tid uint16, mr int, rt time.Duration) {
			m.MessageID = tid
			ticker := time.NewTicker(rt)
			for i := 0; i <= mr; i++ {
				if i == 0 {
					debugMsg("** transmission of message %v of [%s] resource to %s", m.MessageID, o.rsc, o.addr)
					Transmit(o.addr, m)
				}
				select {
				case a := <-o.ack:
					if a == m.MessageID {
						return
					}
				case <-ticker.C:
					if i == mr {
						debugMsg("** transmission of message %v of [%s] resource timeout", m.MessageID, o.rsc)
						h.unregister <- o
						return
					}
					if i <= mr {
						debugMsg("** retransmission #%d of message %v of [%s] resource to %s", i+1, m.MessageID, o.rsc, o.addr)
						Transmit(o.addr, m)
					}
				}
			}
		}(*msg, o.tid, MaxRetransmit, ResponseTimeout)
	}
}

// Observe hub runtime.
func (h *hub) run() {
	for {
		select {
		case o := <-h.register:
			obs, ok := h.observers[o.rsc]
			if !ok {
				obs = make(map[*observer]bool)
				h.observers[o.rsc] = obs
			}
			h.observers[o.rsc][o] = true
			debugMsg("** observer %s added in [%s]", o.addr, o.rsc)
		case o := <-h.unregister:
			delete(h.observers[o.rsc], o)
			close(o.ack)
			close(o.send)
			debugMsg("** observer %s of [%s] removed", o.addr, o.rsc)
		case a := <-h.ack:
			for _, o := range h.observers {
				for obs, _ := range o {
					// send ack to right observer
					if tid, ok := a[obs.addr.String()]; ok && tid != 0 {
						obs.ack <- tid
					}
				}
			}
		}
	}
}
