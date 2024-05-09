package debezium

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"
)

// OutboxPublisher is responsible for publishing events to the outbox
type OutboxPublisher struct {
	db *gorm.DB
}

// Outbox represents the structure of the outbox table
type Outbox struct {
	ID            string `gorm:"column:id;type:uuid;primary_key;default:uuid_generate_v4()"`
	AggregateType string `gorm:"column:aggregatetype;type:varchar(255);not null"`
	AggregateID   string `gorm:"column:aggregateid;type:varchar(255);not null"`
	Type          string `gorm:"column:type;type:varchar(255);not null"`
	Payload       []byte `gorm:"column:payload;type:jsonb;not null"`
}

func NewOutboxPublisher(db *gorm.DB) *OutboxPublisher {
	return &OutboxPublisher{db: db}
}

func (p *OutboxPublisher) Store(_ context.Context, id string, event proto.Message) error {
	marshaler := protojson.MarshalOptions{
		UseProtoNames: true,
		Multiline:     false,
	}
	jsonBytes, err := marshaler.Marshal(event)
	if err != nil {
		return err
	}

	outboxEntry := Outbox{
		AggregateType: string(proto.MessageName(event).Name()),
		AggregateID:   id,
		Payload:       jsonBytes,
	}

	db := p.db.Begin()
	if res := db.Create(&outboxEntry); res.Error != nil {
		return err
	}

	db.Delete(&outboxEntry)

	fmt.Println("Event published successfully")
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
