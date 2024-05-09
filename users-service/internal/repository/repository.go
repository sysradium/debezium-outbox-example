package repository

import (
	"context"

	"github.com/sysradium/debezium-outbox-example/users-service/internal/outbox"
)

type Repository[T any] interface {
	Create(context.Context, T) (T, error)
	Delete(context.Context, uint) error
	Atomic(context.Context, TxFn[T]) (T, error)
	Outbox() outbox.Storer
}

type TxFn[T any] func(context.Context, Repository[T]) (T, error)
