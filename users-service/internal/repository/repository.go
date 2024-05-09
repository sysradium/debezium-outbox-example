package repository

import (
	"github.com/sysradium/debezium-outbox-example/users-service/internal/domain"
	"github.com/sysradium/debezium-outbox-example/users-service/internal/outbox"
)

type Repository interface {
	Create(user domain.User) (domain.User, error)
	Delete(id uint) error
	Atomic(fn TxFn) (domain.User, error)
	Outbox() outbox.Storer
}
