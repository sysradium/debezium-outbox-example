package main

import (
	"context"
	"log"

	"github.com/ThreeDotsLabs/watermill"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"

	"github.com/ThreeDotsLabs/watermill-amazonsqs/sqs"
	"github.com/sysradium/debezium-outbox-example/notifications-service/internal/handlers"
)

func main() {
	ctx := context.Background()
	logger := watermill.NewStdLogger(true, true)

	cfg, err := awsconfig.LoadDefaultConfig(ctx)
	pub, err := sqs.NewPublisher(sqs.PublisherConfig{
		AWSConfig:              cfg,
		CreateQueueIfNotExists: true,
		Marshaler:              sqs.DefaultMarshalerUnmarshaler{},
	}, logger)
	if err != nil {
		log.Fatal(err)
	}
	_ = pub

	sub, err := sqs.NewSubscriber(sqs.SubscriberConfig{
		AWSConfig:                    cfg,
		CreateQueueInitializerConfig: sqs.QueueConfigAtrributes{},
		Unmarshaler:                  sqs.DefaultMarshalerUnmarshaler{},
	}, logger)
	if err != nil {
		log.Fatal(err)
	}

	if err := sub.SubscribeInitialize("events"); err != nil {
		log.Fatal(err)
	}

	messages, err := sub.Subscribe(ctx, "events")
	if err != nil {
		log.Fatal(err)
	}

	handler := handlers.NewUserReigsterdHandler(messages)
	handler.Start(ctx)
}
