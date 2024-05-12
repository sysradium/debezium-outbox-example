package handlers

import (
	"context"
	"log/slog"

	"github.com/ThreeDotsLabs/watermill/message"
)

type handler func(*message.Message) error

type base struct {
	ch     <-chan *message.Message
	logger *slog.Logger
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
