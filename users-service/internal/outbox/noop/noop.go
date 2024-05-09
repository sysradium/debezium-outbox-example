package noop

import (
	"context"
	"log/slog"

	"github.com/spewerspew/spew"
	"github.com/sysradium/debezium-outbox-example/users-service/internal/outbox"
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"
)

var _ outbox.Storer = (*NoopPublisher)(nil)

func New(logger *slog.Logger, tx *gorm.DB) *NoopPublisher {
	return &NoopPublisher{
		logger: logger,
	}
}

type NoopPublisher struct {
	logger *slog.Logger
}

func (n *NoopPublisher) Store(_ context.Context, id string, event proto.Message) error {
	n.logger.Info("publishing event", "event", spew.Sdump(event))
	return nil
}
