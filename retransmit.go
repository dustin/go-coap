package coap

import (
	"fmt"
	"math/rand"
	"net"
	"sync"
	"time"
)

const ACK_TIMEOUT = 2000

const ACK_MAX_TIMEOUT = 3000

const MAX_RETRANSMIT = 4

// generate a random timeout following the RFC
func randTimeout() time.Duration {
	return time.Duration((rand.Intn(ACK_MAX_TIMEOUT-ACK_TIMEOUT) + ACK_TIMEOUT)) * time.Millisecond
}

// a CoAP retransmitter
type Retransmitter struct {
	inflight map[string]*flight // currently non-acked CON messages
	lock     sync.RWMutex
	buf      []byte       // receive buffer
	c        *net.UDPConn // UDP socket

}

// a CoAP Confirmable message in-flight
type flight struct {
	msg     *Message      // the message to be retransmit if needed
	ack     chan struct{} // channel for signaling ack is received
	retrans int           // number of retransmission tentatives
}

func NewRetransmitter(c *net.UDPConn) *Retransmitter {
	s := new(Retransmitter)
	s.c = c
	s.inflight = make(map[string]*flight)
	s.buf = make([]byte, 1200)
	return s
}

func (retrans *Retransmitter) Record(req Message, a *net.UDPAddr) error {
	// record the message if needed
	if req.IsConfirmable() {
		f := flight{&req, make(chan struct{}), 0}

		key := fmt.Sprint("%s#%d", a.String(), req.MessageID)
		retrans.lock.Lock()
		retrans.inflight[key] = &f
		retrans.lock.Unlock()

		go func() {
			// wait for ack or timeout
			timeout := randTimeout()

			for f.retrans < MAX_RETRANSMIT {
				select {
				case <-f.ack:
					retrans.lock.Lock()
					delete(retrans.inflight, key)
					retrans.lock.Unlock()
					break
				case <-time.After(timeout):
					// try again
					f.retrans++
					// double timeout for the next tentative
					timeout = timeout * 2
					if err := retrans.send(req, a); err != nil {
						fmt.Print(err)
					}
				}
			}
		}()
	}

	return retrans.send(req, a)
}

func (retrans *Retransmitter) send(req Message, a *net.UDPAddr) error {
	// send to UDPConn
	d, err := req.MarshalBinary()
	if err != nil {
		return err
	}

	if a == nil {
		_, err = retrans.c.Write(d)
	} else {
		_, err = retrans.c.WriteTo(d, a)
	}
	return err

}

func (retrans *Retransmitter) Received(msg *Message, a *net.UDPAddr) error {
	key := fmt.Sprint("%s#%d", a.String(), msg.MessageID)

	retrans.lock.RLock()
	if f := retrans.inflight[key]; f != nil {
		f.ack <- struct{}{}
	}
	retrans.lock.RUnlock()
	return nil
}
