package handlers

import (
	"encoding/json"
	"errors"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/sysradium/debezium-outbox-example/notifications-service/internal/debezium"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type DefaultUnmarshaler struct{}

func (d DefaultUnmarshaler) Unmarshal(msg *message.Message, v proto.Message) error {
	unmarshaler := protojson.UnmarshalOptions{
		DiscardUnknown: true,
	}

	if err := unmarshaler.Unmarshal(msg.Payload, v); err != nil {
		return err
	}
	return nil
}

type DebeziumUnmarshaler struct{}

func (d DebeziumUnmarshaler) Unmarshal(msg *message.Message, v proto.Message) error {
	var dMsg debezium.Root
	if err := json.Unmarshal(msg.Payload, &dMsg); err != nil {
		return err
	}

	unmarshaler := protojson.UnmarshalOptions{
		DiscardUnknown: true,
	}

	if len(dMsg.Payload.Payload) == 0 {
		return errors.New("invalid message")
	}

	if err := unmarshaler.Unmarshal(dMsg.Payload.Payload, v); err != nil {
		return err
	}

	return nil
}
