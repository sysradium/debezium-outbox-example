package handlers

import (
	"context"
	"fmt"

	"github.com/ThreeDotsLabs/watermill/message"
	pb "github.com/sysradium/debezium-outbox-example/users-service/events"
)

type UserRegisteredHandler struct {
	base
}

func NewUserReigsterdHandler(ch <-chan *message.Message) *UserRegisteredHandler {
	return &UserRegisteredHandler{
		base: newBase(ch),
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

	fmt.Println(uMsg.String())
	return nil
}
