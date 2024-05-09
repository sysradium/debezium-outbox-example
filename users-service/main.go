package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"

	"net/http"

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

type UserCreationRequest struct {
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

func (a *Application) CreateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "only POST is allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()

	var request UserCreationRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	u, err := a.userRepo.Atomic(
		ctx,
		func(ctx context.Context, r repository.Repository) (domain.User, error) {
			u, err := r.Create(
				ctx,
				domain.User{
					Username:  request.Username,
					FirstName: request.FirstName,
					LastName:  request.LastName,
				},
			)
			if err != nil {
				return u, err
			}

			event := &events.UserRegistered{
				Id:        u.ID.String(),
				Username:  u.Username,
				FirstName: u.FirstName,
				LastName:  u.LastName,
			}

			if err := r.Outbox().Store(ctx, uuid.NewString(), event); err != nil {
				return u, err
			}

			return u, nil
		},
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(u); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}

func main() {
	logger := slog.Default()

	dsn := "host=db user=postgres password=some-password dbname=users port=5432 sslmode=disable TimeZone=Europe/Berlin"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		}})
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	// A bit clumsy, but whatever for now
	db.AutoMigrate(&debezium.Outbox{}, &repository.User{})

	app := Application{
		userRepo: repository.NewUserRepository(
			db,
			repository.WithLogger(logger),
		),
	}

	http.HandleFunc("/users", app.CreateUser)
	http.ListenAndServe(":8080", nil)
}
