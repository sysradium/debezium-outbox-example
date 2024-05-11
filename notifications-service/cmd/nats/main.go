package main

import (
	"context"
	"log"
	"log/slog"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-nats/v2/pkg/nats"
	nc "github.com/nats-io/nats.go"
	"github.com/sysradium/debezium-outbox-example/notifications-service/internal/handlers"
)

func main() {
	ctx := context.Background()
	logger := slog.Default()

	subscriber, err := nats.NewSubscriber(
		nats.SubscriberConfig{
			URL: "nats://nats:4222",
			JetStream: nats.JetStreamConfig{
				SubscribeOptions: []nc.SubOpt{
					nc.DeliverLast(),
					nc.AckExplicit(),
				},
				TrackMsgId: true,
				AckAsync:   false,
			},
		}, watermill.NewSlogLogger(logger),
	)
	if err != nil {
		log.Fatal(err)
	}

	messages, err := subscriber.Subscribe(ctx, "outbox.event.UserRegistered")
	if err != nil {
		log.Fatal(err)
	}
	handler := handlers.NewUserReigsterdHandler(messages)
	handler.Start(ctx)
}
