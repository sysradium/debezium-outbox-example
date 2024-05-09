package repository

import (
	"context"

	"github.com/sysradium/debezium-outbox-example/users-service/internal/domain"
	"github.com/sysradium/debezium-outbox-example/users-service/internal/outbox"
)

type Repository interface {
	Create(context.Context, domain.User) (domain.User, error)
	Delete(context.Context, uint) error
	Atomic(context.Context, TxFn) (domain.User, error)
	Outbox() outbox.Storer
}

type TxFn func(context.Context, Repository) (domain.User, error)
