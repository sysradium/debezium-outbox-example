package basic

import (
	"context"
	"log/slog"

	"github.com/sysradium/debezium-outbox-example/users-service/internal/outbox"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"
)

type Outbox struct {
	ID            string `gorm:"column:id;type:uuid;primary_key;default:uuid_generate_v4()"`
	AggregateType string `gorm:"column:aggregatetype;type:varchar(255);not null"`
	AggregateID   string `gorm:"column:aggregateid;type:varchar(255);not null"`
	Payload       []byte `gorm:"column:payload;type:jsonb;not null"`
}

type Marshaler interface {
	Marshal(proto.Message) ([]byte, error)
}

type OutboxStorage struct {
	tx        *gorm.DB
	logger    *slog.Logger
	marshaler Marshaler
}

func NewOutbox() func(*gorm.DB) outbox.Storer {
	return func(tx *gorm.DB) outbox.Storer {
		pub := &OutboxStorage{
			tx:     tx.Table("my-outbox"),
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

	outboxEntry := Outbox{
		AggregateType: string(proto.MessageName(event).Name()),
		AggregateID:   id,
		Payload:       jsonBytes,
	}

	tx := p.tx.WithContext(ctx)
	if res := tx.Create(&outboxEntry); res.Error != nil {
		return err
	}

	return nil
}
