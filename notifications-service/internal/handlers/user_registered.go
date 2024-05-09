package handlers

import (
	"context"
	"log"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/spewerspew/spew"
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
	// var dMsg debezium.Root
	// if err := json.Unmarshal(msg.Payload, &dMsg); err != nil {
	// 	return err
	// }
	// fmt.Printf("received message: %v\n", dMsg.Payload.Id)

	unmarshaler := protojson.UnmarshalOptions{
		DiscardUnknown: true,
	}

	spew.Dump(string(msg.Payload))
	uMsg := pb.UserRegistered{}
	if err := unmarshaler.Unmarshal(msg.Payload, &uMsg); err != nil {
		return err
	}
	// if err := unmarshaler.Unmarshal(dMsg.Payload.Payload, &uMsg); err != nil {
	// 	return err
	// }

	spew.Dump(uMsg)
	return nil
}
