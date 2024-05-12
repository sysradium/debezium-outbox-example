package handlers

import (
	"context"
	"log/slog"

	"github.com/ThreeDotsLabs/watermill/message"
	"google.golang.org/protobuf/proto"
)

type unmarshaler interface {
	Unmarshal(msg *message.Message, v proto.Message) error
}

type handler func(*message.Message) error

type base struct {
	ch                  <-chan *message.Message
	logger              *slog.Logger
	defaultUnmarshaler  unmarshaler
	debeziumUnmarshaler unmarshaler
}

func newBase(ch <-chan *message.Message) base {
	return base{
		ch:                  ch,
		defaultUnmarshaler:  DefaultUnmarshaler{},
		debeziumUnmarshaler: DebeziumUnmarshaler{},
		logger:              slog.Default(),
	}

}

func (b *base) Start(ctx context.Context, h handler) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case msg, ok := <-b.ch:
			if !ok {
				return nil
			}

			if err := h(msg); err != nil {
				b.logger.Info("unable to process message: %v", err)
				msg.Nack()
				continue
			}

			msg.Ack()
		}
	}
}

func (b *base) unarmshal(msg *message.Message, out proto.Message) error {
	if err := b.debeziumUnmarshaler.Unmarshal(msg, out); err != nil {
		if err := b.defaultUnmarshaler.Unmarshal(msg, out); err != nil {
			return err
		}
	}

	return nil
}
