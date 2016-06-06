// This file contains a core state handler for MySensors.
package mysensors

import (
	"bufio"
	"io"
	"log"
	"strconv"
	"time"
)

func NewHandler(r io.Reader, w io.Writer, c chan *Message, n *Network) *Handler {
	return &Handler{r: r, w: w, c: c, network: n}
}

type Handler struct {
	r     io.Reader
	w     io.Writer
	c     chan *Message
	ready bool
	network *Network
}

func (h *Handler) Start() {
	rCh := make(chan *Message)
	wCh := make(chan *Message)
	go h.messageWriter(wCh)
	go h.messageReader(rCh)

	for m := range rCh {
		var r *Message
		switch m.Type {
		case MsgInternal:
			r = h.processInternal(m)
		case MsgSet:
			r = h.processSet(m)
		case MsgPresentation:
			r = h.processPresentation(m)
		}
		if h.ready && r != nil {
			wCh <- r
		}
	}
	log.Printf("Read channel closed.")
	close(h.c)
}

func (h *Handler) processPresentation(m *Message) *Message {
	h.c <- m
	return nil
}

func (h *Handler) processSet(m *Message) *Message {
	h.c <- m
	return nil
}

func (h *Handler) processInternal(m *Message) *Message {
	var r *Message
	subType := m.SubType.(SubTypeInternal)
	switch subType {
	case I_ID_REQUEST:
		r = m.Copy()
		r.SubType = I_ID_RESPONSE
		sensorID := h.network.NextNodeID()
		r.Payload = []byte(strconv.Itoa(int(sensorID)))
	case I_CONFIG:
		r = m.Copy()
		r.SubType = I_CONFIG
		r.Payload = []byte("M")
	case I_GATEWAY_READY:
		h.ready = true
		h.c <- m
		log.Printf("Gateway ready!\n")
	case I_TIME:
		r = m.Copy()
		r.Payload = []byte(strconv.FormatInt(time.Now().Unix(), 10))
	default:
		log.Printf("UNSUPPORTED MSG: %s\n", m)
		h.c <- m
	}
	return r
}

func (h *Handler) messageReader(c chan *Message) {
	r := bufio.NewReader(h.r)
	for {
		d, err := r.ReadBytes('\x0a')
		if err != nil {
			log.Printf("Read error: %v\n", err)
			break
		}
		m := &Message{}
		if err = m.Unmarshal(d); err != nil {
			log.Printf("Error parsing [%s]: %v\n", string(d), err)
			continue
		}
		log.Printf("RX: %s\n", m)
		c <- m
	}
}

func (h *Handler) messageWriter(c chan *Message) {
	for m := range c {
		reply := m.Marshal()
		log.Printf("TX: %s\n", reply)
		if n, err := h.w.Write(reply); err != nil || n != len(reply) {
			log.Fatalf("Write error: %v\n", err)
		}
	}
}
