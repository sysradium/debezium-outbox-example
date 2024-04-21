package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	nc "github.com/nats-io/nats.go"
	"github.com/spewerspew/spew"
	pb "github.com/sysradium/debezium-outbox-example/users-service/events"
	"google.golang.org/protobuf/encoding/protojson"
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

	sub, err := js.Subscribe("outbox.event.UserRegistered", func(msg *nc.Msg) {
		var m Root
		if err := json.Unmarshal(msg.Data, &m); err != nil {
			return
		}
		fmt.Printf("received message: %v\n", m.Payload.Id)

		unmarshaler := protojson.UnmarshalOptions{
			DiscardUnknown: true,
		}

		uMsg := pb.UserRegistered{}
		if err := unmarshaler.Unmarshal(m.Payload.Payload, &uMsg); err != nil {
			log.Println("unable to unmarshal message")
			return
		}

		spew.Dump(uMsg)
		msg.Ack()
	})
	if err != nil {
		log.Fatal(err)
	}

	defer sub.Unsubscribe()

	time.Sleep(60 * time.Second)
}
