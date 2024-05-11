package main

import (
	"context"
	"encoding/json"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sysradium/debezium-outbox-example/users-service/internal/app"
	"github.com/sysradium/debezium-outbox-example/users-service/internal/app/commands"
	"github.com/sysradium/debezium-outbox-example/users-service/internal/domain"
	"github.com/sysradium/debezium-outbox-example/users-service/internal/outbox/basic"
	"github.com/sysradium/debezium-outbox-example/users-service/internal/outbox/debezium"
	"github.com/sysradium/debezium-outbox-example/users-service/internal/publishers"
	"github.com/sysradium/debezium-outbox-example/users-service/internal/repository"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type UserCreationRequest struct {
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type Server struct {
	a app.Application
}

func (s *Server) CreateUser(w http.ResponseWriter, r *http.Request) {
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

	u, err := s.a.Commands.CreateUser.Handle(ctx, commands.CreateUser{
		Username:  request.Username,
		FirstName: request.FirstName,
		LastName:  request.LastName,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(u); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}

type OutboxType int

const (
	OUTBOX_TYPE_DEBEZIUM OutboxType = iota + 1
	OUTBOX_TYPE_BASIC
)

func main() {
	logger := slog.Default()
	outbox := OUTBOX_TYPE_BASIC

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

	var repo repository.Repository[domain.User]
	if outbox == OUTBOX_TYPE_BASIC {
		publisher, err := publishers.NewNatsPublisher(logger, "outbox")
		if err != nil {
			log.Fatal(err)
		}

		worker := basic.NewWorker(
			db,
			publisher,
			basic.WithLogger(logger),
			basic.WithPollingInterval(time.Millisecond*50),
			basic.WithTopicPrefix("outbox.event"),
		)
		go worker.Start()
		defer func() {
			logger.Info("shutting worker down")
			worker.Stop()
			logger.Info("exiting")
		}()
		repo = repository.NewUserRepository(
			db,
			repository.WithLogger(logger),
			repository.WithOutbox(basic.NewOutbox()),
		)

	} else if outbox == OUTBOX_TYPE_DEBEZIUM {
		repo = repository.NewUserRepository(
			db,
			repository.WithLogger(logger),
			repository.WithOutbox(debezium.NewOutboxPublisher()),
		)
	}

	srv := Server{
		a: app.NewApplication(repo),
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	mux := http.NewServeMux()
	mux.HandleFunc("/users", srv.CreateUser)
	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	go func() {
		server.ListenAndServe()
	}()

	<-sigs

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	server.Shutdown(ctx)
	cancel()
}
