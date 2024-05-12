package handlers

import (
	"context"
	"log/slog"

	"github.com/ThreeDotsLabs/watermill/message"
	pb "github.com/sysradium/debezium-outbox-example/users-service/events"
)

type UserRegisteredHandler struct {
	base
	logger *slog.Logger
}

func NewUserReigsterdHandler(ch <-chan *message.Message) *UserRegisteredHandler {
	return &UserRegisteredHandler{
		base:   newBase(ch),
		logger: slog.Default(),
	}
}

func (a *UserRegisteredHandler) Start(ctx context.Context) error {
	return a.base.Start(ctx, a.Handle)
}

func (a *UserRegisteredHandler) Handle(msg *message.Message) error {
	uMsg := pb.UserRegistered{}

	if err := a.base.unarmshal(msg, &uMsg); err != nil {
		return err
	}

	a.logger.Info("User registered", "user", uMsg.String())
	return nil
}
