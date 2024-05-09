package main

import (
	"context"
	"log"
	"strconv"

	"github.com/google/uuid"
	"github.com/sysradium/debezium-outbox-example/users-service/events"
	"github.com/sysradium/debezium-outbox-example/users-service/internal/domain"
	"github.com/sysradium/debezium-outbox-example/users-service/internal/outbox/debezium"
	"github.com/sysradium/debezium-outbox-example/users-service/internal/repository"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type Application struct {
	userRepo repository.Repository
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

	db.AutoMigrate(&debezium.Outbox{}, &repository.User{})

	app := Application{
		userRepo: repository.NewUserRepository(db),
	}

	if _, err := app.userRepo.Atomic(
		func(r repository.Repository) (domain.User, error) {
			u, err := r.Create(
				domain.User{
					Username:  "johndoe",
					FirstName: "John",
					LastName:  "Doe",
				},
			)
			if err != nil {
				return u, err
			}

			event := &events.UserRegistered{
				Id:        strconv.FormatUint(uint64(u.ID), 10),
				Username:  u.Username,
				FirstName: u.FirstName,
				LastName:  u.LastName,
			}

			if err := r.Outbox().Store(context.Background(), uuid.NewString(), event); err != nil {
				return u, err
			}

			return u, nil
		},
	); err != nil {
		log.Fatal(err)
	}
}
