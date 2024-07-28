package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sysradium/debezium-outbox-example/users-service/internal/app"
	"github.com/sysradium/debezium-outbox-example/users-service/internal/domain"
	"github.com/sysradium/debezium-outbox-example/users-service/internal/outbox/basic"
	"github.com/sysradium/debezium-outbox-example/users-service/internal/outbox/debezium"
	"github.com/sysradium/debezium-outbox-example/users-service/internal/ports"
	"github.com/sysradium/debezium-outbox-example/users-service/internal/publishers"
	"github.com/sysradium/debezium-outbox-example/users-service/internal/repository"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type OutboxType int

const (
	OUTBOX_TYPE_DEBEZIUM OutboxType = iota + 1
	OUTBOX_TYPE_BASIC
)

func main() {
	logger := slog.Default()
	// hardcoded just for an example
	outbox := OUTBOX_TYPE_BASIC

	db, err := gorm.Open(postgres.Open(os.Getenv("DB_DSN")), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		}})
	if err != nil {
		logger.Error("failed connecting to db", "error", err.Error())
		os.Exit(1)
	}

	// A bit clumsy, but whatever for now
	if err := db.AutoMigrate(&debezium.Outbox{}, &repository.User{}); err != nil {
		logger.Error("migration failed", "error", err)
		os.Exit(1)
	}

	var repo repository.Repository[domain.User]
	switch outbox {
	case OUTBOX_TYPE_BASIC:
		publisher, err := publishers.NewNatsPublisher(logger, os.Getenv("NATS_URL"), "outbox")
		if err != nil {
			log.Fatal(err)
		}

		sdb, err := db.DB()
		if err != nil {
			logger.Error("can not retrieve db", "error", err)
			os.Exit(1)
		}

		worker := basic.NewWorker(
			sdb,
			publisher,
			basic.WithLogger(logger),
			basic.WithPollingInterval(time.Millisecond*50),
			basic.WithTopicPrefix("outbox.event"),
		)

		go func() {
			if err := worker.Start(); err != nil {
				logger.Error("unable to start a worker", "error", err)
				os.Exit(1)
			}
		}()
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

	case OUTBOX_TYPE_DEBEZIUM:
		repo = repository.NewUserRepository(
			db,
			repository.WithLogger(logger),
			//repository.WithOutbox(debezium.NewOutboxPublisher()),
		)
	}

	srv := ports.NewHTTP(app.NewApplication(repo))
	mux := http.NewServeMux()
	mux.HandleFunc("/users", srv.CreateUser)
	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	if err := server.Shutdown(ctx); err != nil {
		log.Fatal(err)
	}
	cancel()
}
