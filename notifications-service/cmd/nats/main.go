package main

import (
	"context"
	"log"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/google/uuid"
	nc "github.com/nats-io/nats.go"
	"github.com/sysradium/debezium-outbox-example/notifications-service/internal/handlers"
)

func main() {
	ctx := context.Background()

	n, err := nc.Connect("nats://nats:4222")
	if err != nil {
		log.Fatal(err)
	}
	defer n.Close()

	js, err := n.JetStream()
	if err != nil {
		log.Fatal(err)
	}

	messages := make(chan *message.Message)
	handler := handlers.NewUserReigsterdHandler(messages)

	sub, err := js.Subscribe("outbox.event.UserRegistered", func(msg *nc.Msg) {
		messages <- message.NewMessage(uuid.NewString(), msg.Data)
		msg.Ack()
	})
	if err != nil {
		log.Fatal(err)
	}

	defer sub.Unsubscribe()

	handler.Start(ctx)
}
