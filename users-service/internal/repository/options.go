package repository

import (
	"log/slog"

	"github.com/sysradium/debezium-outbox-example/users-service/internal/outbox"
	"gorm.io/gorm"
)

type option func(d *UserRepository)

func WithLogger(l *slog.Logger) option {
	return func(u *UserRepository) {
		u.logger = l
	}
}

type outboxFn func(db *gorm.DB) outbox.Storer

func WithOutbox(o outboxFn) option {
	return func(u *UserRepository) {
		u.outboxFactory = o
	}
}
