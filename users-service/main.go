package main

import (
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/sysradium/debezium-outbox-example/users-service/events"
	"github.com/sysradium/debezium-outbox-example/users-service/internal/repository"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// OutboxPublisher is responsible for publishing events to the outbox
type OutboxPublisher struct {
	db *gorm.DB
}

// Outbox represents the structure of the outbox table
type Outbox struct {
	ID            string `gorm:"column:id;type:uuid;primary_key"`
	AggregateType string `gorm:"column:aggregatetype;type:varchar(255);not null"`
	AggregateID   string `gorm:"column:aggregateid;type:varchar(255);not null"`
	Type          string `gorm:"column:type;type:varchar(255);not null"`
	Payload       []byte `gorm:"column:payload;type:jsonb;not null"`
}

func NewOutboxPublisher(db *gorm.DB) *OutboxPublisher {
	return &OutboxPublisher{db: db}
}

func (p *OutboxPublisher) Publish(event proto.Message) error {
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
		AggregateID:   uuid.New().String(),
		Payload:       jsonBytes,
	}

	db := p.db.Begin()
	res := db.Create(&outboxEntry)
	if res.Error != nil {
		return err
	}

	db.Delete(&outboxEntry)
	db.Commit()

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
		ID:            uuid.NewString(),
		AggregateType: string(proto.MessageName(event).Name()),
		AggregateID:   uuid.New().String(),
		Payload:       jsonBytes,
	}, nil

}

func main() {
	dsn := "host=db user=postgres password=some-password dbname=users port=5432 sslmode=disable TimeZone=Europe/Berlin"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		}})
	if err != nil {
		log.Fatal(err)
	}

	if err := db.AutoMigrate(&Outbox{}, &repository.User{}); err != nil {
		log.Fatal(err)
	}

	userRepo := repository.NewUserRepository(db)
	if _, err := userRepo.Atomic(
		func(r *repository.UserRepository) (repository.User, error) {
			u, err := r.Create(
				repository.User{
					Username:  "johndoe",
					FirstName: "John",
					LastName:  "Doe",
				},
			)
			if err != nil {
				return u, err
			}

			outboxEvent, err := newOutboxMessageFromEvent(&events.UserRegistered{
				Username:  u.Username,
				FirstName: u.FirstName,
				LastName:  u.LastName,
			})

			if err != nil {
				return u, err
			}

			if err := r.DB.Create(&outboxEvent).Error; err != nil {
				return u, err
			}

			return u, nil
		},
	); err != nil {
		log.Fatal(err)
	}
}
