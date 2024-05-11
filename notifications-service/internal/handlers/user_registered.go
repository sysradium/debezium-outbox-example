package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/spewerspew/spew"
	"github.com/sysradium/debezium-outbox-example/notifications-service/internal/debezium"
	pb "github.com/sysradium/debezium-outbox-example/users-service/events"
	"google.golang.org/protobuf/encoding/protojson"
)

type UserRegisteredHandler struct {
	ch <-chan *message.Message
}

func NewUserReigsterdHandler(ch <-chan *message.Message) *UserRegisteredHandler {
	return &UserRegisteredHandler{
		ch: ch,
	}
}

func (a *UserRegisteredHandler) Start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-a.ch:
			if !ok {
				return
			}

			if err := a.Handle(msg); err != nil {
				log.Printf("unable to process message: %v", err)
				msg.Nack()
				continue
			}

			msg.Ack()
		}
	}
}

func (a *UserRegisteredHandler) Handle(msg *message.Message) error {
	var eventPayload []byte

	var dMsg debezium.Root
	if err := json.Unmarshal(msg.Payload, &dMsg); err != nil || len(dMsg.Payload.Payload) == 0 {
		fmt.Println("unable to decode, maybe not a debezium event")
		eventPayload = msg.Payload
	} else {
		eventPayload = dMsg.Payload.Payload
	}

	unmarshaler := protojson.UnmarshalOptions{
		DiscardUnknown: true,
	}

	uMsg := pb.UserRegistered{}
	if err := unmarshaler.Unmarshal(eventPayload, &uMsg); err != nil {
		spew.Dump(string(eventPayload))
		return err
	}

	spew.Dump(uMsg)
	return nil
}
