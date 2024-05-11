package publishers

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-nats/v2/pkg/nats"
	"github.com/ThreeDotsLabs/watermill/message"
	nc "github.com/nats-io/nats.go"
	natsJS "github.com/nats-io/nats.go/jetstream"
)

func NewNatsPublisher(logger *slog.Logger, prefix string) (*nats.Publisher, error) {
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

type Marshaler struct{}

func (*Marshaler) Marshal(topic string, m *message.Message) (*nc.Msg, error) {
	natsMsg := nc.NewMsg(topic)
	natsMsg.Data = m.Payload

	return natsMsg, nil
}
