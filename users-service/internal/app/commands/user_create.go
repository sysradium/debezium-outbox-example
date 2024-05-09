package commands

import (
	"context"

	"github.com/google/uuid"
	"github.com/sysradium/debezium-outbox-example/users-service/events"
	"github.com/sysradium/debezium-outbox-example/users-service/internal/domain"
	"github.com/sysradium/debezium-outbox-example/users-service/internal/repository"
)

type CreateUser struct {
	Username  string
	FirstName string
	LastName  string
}
type CreateUserHandler CommandHandler[CreateUser, domain.User]

type userCreateHandler struct {
	repo repository.Repository[domain.User]
}

func (c userCreateHandler) Handle(ctx context.Context, cmd CreateUser) (domain.User, error) {
	u, err := c.repo.Atomic(
		ctx,
		func(ctx context.Context, r repository.Repository[domain.User]) (domain.User, error) {
			u, err := r.Create(
				ctx,
				domain.User{
					Username:  cmd.Username,
					FirstName: cmd.FirstName,
					LastName:  cmd.LastName,
				},
			)
			if err != nil {
				return u, err
			}

			event := &events.UserRegistered{
				Id:        u.ID.String(),
				Username:  u.Username,
				FirstName: u.FirstName,
				LastName:  u.LastName,
			}

			if err := r.Outbox().Store(ctx, uuid.NewString(), event); err != nil {
				return u, err
			}

			return u, nil
		},
	)

	if err != nil {
		return u, err
	}

	return u, nil

}

func NewUserCreateHandler(repo repository.Repository[domain.User]) CreateUserHandler {
	return &userCreateHandler{
		repo: repo,
	}
}
