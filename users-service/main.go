package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-nats/v2/pkg/nats"
	"github.com/ThreeDotsLabs/watermill/message"
	nc "github.com/nats-io/nats.go"
	natsJS "github.com/nats-io/nats.go/jetstream"
	"github.com/sysradium/debezium-outbox-example/users-service/internal/app"
	"github.com/sysradium/debezium-outbox-example/users-service/internal/app/commands"
	"github.com/sysradium/debezium-outbox-example/users-service/internal/outbox/basic"
	"github.com/sysradium/debezium-outbox-example/users-service/internal/outbox/debezium"
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

type Marshaler struct{}

func (*Marshaler) Marshal(topic string, m *message.Message) (*nc.Msg, error) {
	natsMsg := nc.NewMsg(topic)
	natsMsg.Data = m.Payload

	return natsMsg, nil
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

	publisher, err := newNatsPublisher(logger, "outbox.event")
	if err != nil {
		log.Fatal(err)
	}

	pub := basic.NewWorker(
		db,
		publisher,
		basic.WithLogger(logger),
		basic.WithPollingInterval(time.Millisecond*50),
		basic.WithTopicPrefix("outbox.event"),
	)
	go pub.Start()
	defer pub.Stop()

	srv := Server{
		a: app.NewApplication(repository.NewUserRepository(
			db,
			repository.WithLogger(logger),
			repository.WithOutbox(basic.NewOutbox()),
		))}

	http.HandleFunc("/users", srv.CreateUser)
	http.ListenAndServe(":8080", nil)

}

func newNatsPublisher(logger *slog.Logger, prefix string) (*nats.Publisher, error) {
	conn, err := nc.Connect(
		"nats://nats:4222",
		nc.RetryOnFailedConnect(true),
		nc.Timeout(30*time.Second),
		nc.ReconnectWait(1*time.Second),
	)

	if err != nil {
		return nil, err
	}
	js, err := natsJS.New(conn)
	if err != nil {
		return nil, err
	}

	if _, err := js.CreateOrUpdateStream(context.Background(),
		natsJS.StreamConfig{
			Name:     "DebeziumEvents",
			Subjects: []string{fmt.Sprintf("%s.>", prefix)},
		},
	); err != nil {
		return nil, err
	}

	publisher, err := nats.NewPublisherWithNatsConn(
		conn,
		nats.PublisherPublishConfig{
			Marshaler:         &Marshaler{},
			SubjectCalculator: nats.DefaultSubjectCalculator,
			JetStream: nats.JetStreamConfig{
				ConnectOptions: nil,
				SubscribeOptions: []nc.SubOpt{
					nc.DeliverAll(),
					nc.AckExplicit(),
				},
				PublishOptions: nil,
				DurablePrefix:  "",
			},
		},
		watermill.NewSlogLogger(logger),
	)

	return publisher, err
}
