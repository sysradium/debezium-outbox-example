package debezium

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"
)

// OutboxPublisher is responsible for publishing events to the outbox
type OutboxPublisher struct {
	tx        *gorm.DB
	logger    *slog.Logger
	marshaler Marshaler
}

// Outbox represents the structure of the outbox table
type Outbox struct {
	ID            string `gorm:"column:id;type:uuid;primary_key;default:uuid_generate_v4()"`
	AggregateType string `gorm:"column:aggregatetype;type:varchar(255);not null"`
	AggregateID   string `gorm:"column:aggregateid;type:varchar(255);not null"`
	Type          string `gorm:"column:type;type:varchar(255);not null"`
	Payload       []byte `gorm:"column:payload;type:jsonb;not null"`
}

type Marshaler interface {
	Marshal(proto.Message) ([]byte, error)
}

func NewOutboxPublisher(tx *gorm.DB, opts ...option) *OutboxPublisher {
	pub := &OutboxPublisher{
		tx:     tx,
		logger: slog.Default(),
		marshaler: protojson.MarshalOptions{
			UseProtoNames: true,
			Multiline:     false,
		},
	}

	for _, o := range opts {
		o(pub)
	}

	return pub
}

func (p *OutboxPublisher) Store(ctx context.Context, id string, event proto.Message) error {
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

	// We do not need the entry itself to be preserved.
	// We only care about the event of record creation, which triggers debezium to generate an event
	tx.Delete(&outboxEntry)

	p.logger.Info("Event published successfully", "id", id, "type", outboxEntry.AggregateType)

	return nil
}

func newOutboxMessageFromEvent(event proto.Message) (Outbox, error) {
	marshaler := protojson.MarshalOptions{
		UseProtoNames: true,
		Multiline:     false,
	}
	jsonBytes, err := marshaler.Marshal(event)
	if err != nil {
		return Outbox{}, err
	}

	return Outbox{
		AggregateType: string(proto.MessageName(event).Name()),
		AggregateID:   uuid.New().String(),
		Payload:       jsonBytes,
	}, nil

}
