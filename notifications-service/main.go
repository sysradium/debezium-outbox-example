package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	nc "github.com/nats-io/nats.go"
	"github.com/spewerspew/spew"
)

type Root struct {
	Schema  Schema  `json:"schema"`
	Payload Payload `json:"payload"`
}

type Schema struct {
	Type     string  `json:"type"`
	Fields   []Field `json:"fields"`
	Optional bool    `json:"optional"`
	Name     string  `json:"name"`
}

type Field struct {
	Type     string  `json:"type"`
	Optional bool    `json:"optional"`
	Field    string  `json:"field"`
	Name     string  `json:"name,omitempty"`
	Version  int     `json:"version,omitempty"`
	Fields   []Field `json:"fields,omitempty"`
}

type Payload struct {
	Payload json.RawMessage `json:"payload"`
	Id      string          `json:"id"`
}

func main() {
	n, err := nc.Connect("nats://nats:4222")
	if err != nil {
		log.Fatal(err)
	}
	defer n.Close()

	// Create JetStream Context
	js, err := n.JetStream()
	if err != nil {
		log.Fatal(err)
	}

	sub, err := js.Subscribe("outbox.event.TypeA", func(msg *nc.Msg) {
		var m Root
		if err := json.Unmarshal(msg.Data, &m); err != nil {
			return
		}

		fmt.Printf("received message: %v\n", m.Payload.Id)
		spew.Dump(string(m.Payload.Payload))
		msg.Ack()
	})
	if err != nil {
		log.Fatal(err)
	}

	defer sub.Unsubscribe()

	time.Sleep(60 * time.Second)
}
