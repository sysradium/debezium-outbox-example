package publishers

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-nats/v2/pkg/nats"
	nc "github.com/nats-io/nats.go"
	natsJS "github.com/nats-io/nats.go/jetstream"
)

func NewNatsPublisher(logger *slog.Logger, url, prefix string) (*nats.Publisher, error) {
	conn, err := nc.Connect(
		url,
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
			Name:     "OUTBOX",
			Subjects: []string{fmt.Sprintf("%s.>", prefix)},
		},
	); err != nil {
		return nil, err
	}

	publisher, err := nats.NewPublisherWithNatsConn(
		conn,
		nats.PublisherPublishConfig{
			Marshaler:         &nats.NATSMarshaler{},
			SubjectCalculator: nats.DefaultSubjectCalculator,
			JetStream: nats.JetStreamConfig{
				ConnectOptions: nil,
				SubscribeOptions: []nc.SubOpt{
					nc.DeliverAll(),
					nc.AckExplicit(),
				},
				PublishOptions: nil,
				DurablePrefix:  "",
				TrackMsgId:     true,
			},
		},
		watermill.NewSlogLogger(logger),
	)

	return publisher, err
}
