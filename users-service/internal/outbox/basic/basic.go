package basic

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/sysradium/debezium-outbox-example/users-service/internal/outbox"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

const createSQLStatement = `
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS "my-outbox" (
    id uuid NOT NULL DEFAULT uuid_generate_v4(),
    aggregatetype character varying(255) NOT NULL,
    aggregateid character varying(255) NOT NULL,
    payload jsonb NOT NULL,
    attempts integer NOT NULL,
    PRIMARY KEY (id)
);`

type Outbox struct {
	ID            string
	AggregateType string
	AggregateID   string
	Attempts      int
	Payload       []byte
}

type Marshaler interface {
	Marshal(proto.Message) ([]byte, error)
}

type OutboxStorage struct {
	tx        *sql.DB
	logger    *slog.Logger
	marshaler Marshaler
}

func NewOutbox() func(*sql.DB) outbox.Storer {
	return func(tx *sql.DB) outbox.Storer {
		_, err := tx.Exec(createSQLStatement)
		if err != nil {
			panic(err)
		}
		pub := &OutboxStorage{
			tx:     tx,
			logger: slog.Default(),
			marshaler: protojson.MarshalOptions{
				UseProtoNames: true,
				Multiline:     false,
			},
		}

		return pub
	}
}

func (p *OutboxStorage) Store(ctx context.Context, id string, event proto.Message) error {
	jsonBytes, err := p.marshaler.Marshal(event)
	if err != nil {
		return err
	}

	if _, err := p.tx.ExecContext(
		ctx,
		"INSERT INTO my-outbox (id, aggregateType, aggregateId, payload) VALUES(uuid_generate_v4(), ? ? ?)",
		string(proto.MessageName(event).Name()),
		id,
		jsonBytes,
	); err != nil {
		return err
	}

	return nil
}
