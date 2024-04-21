package main

import (
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/sysradium/debezium-outbox-example/users-service/events"
	"google.golang.org/protobuf/encoding/protojson"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// EventPublisher is responsible for publishing events to the outbox
type EventPublisher struct {
	db *gorm.DB
}

// Outbox represents the structure of the outbox table
type Outbox struct {
	ID            uuid.UUID `gorm:"column:id;type:uuid;primary_key;"`
	AggregateType string    `gorm:"column:aggregatetype;type:varchar(255);not null"`
	AggregateID   string    `gorm:"column:aggregateid;type:varchar(255);not null"`
	Type          string    `gorm:"column:type;type:varchar(255);not null"`
	Payload       []byte    `gorm:"column:payload;type:jsonb;not null"`
}

// NewEventPublisher creates a new instance of EventPublisher
func NewEventPublisher(db *gorm.DB) *EventPublisher {
	return &EventPublisher{db: db}
}

// Publish marshals the UserRegistered struct and stores it in the outbox
func (p *EventPublisher) Publish(event *events.UserRegistered) error {
	marshaler := protojson.MarshalOptions{
		UseProtoNames: true,
		Multiline:     false,
	}
	jsonBytes, err := marshaler.Marshal(event)
	if err != nil {
		return err
	}

	outboxEntry := Outbox{
		ID:            uuid.New(),
		AggregateType: "UserRegistered",
		//		AggregateID:   uuid.New().String(), // Assuming AggregateID is another UUID
		Type:    "UserRegisteredEvent",
		Payload: jsonBytes,
	}

	// Save the entry to the database
	if err := p.db.Create(&outboxEntry).Error; err != nil {
		return err
	}

	fmt.Println("Event published successfully")
	return nil
}

func main() {
	// Set up the database connection
	dsn := "host=db user=postgres password=some-password dbname=users port=5432 sslmode=disable TimeZone=Europe/Berlin"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		}})
	if err != nil {
		log.Fatal(err)
	}

	// Migrate the Outbox table
	db.AutoMigrate(&Outbox{})

	// Create an event publisher
	publisher := NewEventPublisher(db)

	// Create a UserRegistered event
	userEvent := &events.UserRegistered{
		Username:  "johndoe",
		FirstName: "John",
		LastName:  "Doe",
	}

	// Publish the event
	if err := publisher.Publish(userEvent); err != nil {
		log.Fatal(err)
	}
}
