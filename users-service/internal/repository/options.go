package repository

import (
	"database/sql"
	"log/slog"

	"github.com/sysradium/debezium-outbox-example/users-service/internal/outbox"
)

type option func(d *UserRepository)

func WithLogger(l *slog.Logger) option {
	return func(u *UserRepository) {
		u.logger = l
	}
}

type outboxFn func(db *sql.DB) outbox.Storer

func WithOutbox(o outboxFn) option {
	return func(u *UserRepository) {
		u.outboxFactory = o
	}
}
