package basic

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/sysradium/debezium-outbox-example/users-service/internal/outbox"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Outbox struct {
	ID            string `gorm:"column:id;type:uuid;primary_key;default:uuid_generate_v4()"`
	AggregateType string `gorm:"column:aggregatetype;type:varchar(255);not null"`
	AggregateID   string `gorm:"column:aggregateid;type:varchar(255);not null"`
	Type          string `gorm:"column:type;type:varchar(255);not null"`
	Payload       []byte `gorm:"column:payload;type:jsonb;not null"`
}

type Marshaler interface {
	Marshal(proto.Message) ([]byte, error)
}

type Worker struct {
	db              *gorm.DB
	logger          *slog.Logger
	marshaler       Marshaler
	pollingInterval time.Duration
	publisher       message.Publisher
	batchSize       int
	topicPrefix     string
	done            <-chan struct{}
	ctx             context.Context
	cancel          context.CancelFunc
}

func NewWorker(db *gorm.DB, publisher message.Publisher, opts ...workerOption) *Worker {
	ctx, cancel := context.WithCancel(context.Background())
	w := &Worker{
		db:     db,
		logger: slog.Default(),
		marshaler: protojson.MarshalOptions{
			UseProtoNames: true,
			Multiline:     false,
		},
		publisher: publisher,
		batchSize: 10000,
		ctx:       ctx,
		cancel:    cancel,
	}

	for _, o := range opts {
		o(w)
	}

	return w
}

func (p *Worker) Start() {
	ticker := time.NewTicker(p.pollingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-p.ctx.Done():
			break
		case <-ticker.C:
			p.processMessages(p.ctx, p.batchSize)
		}
	}
}

func (p *Worker) Stop() {
	p.cancel()
	<-p.done
}

func (p *Worker) processMessages(ctx context.Context, batchSize int) error {
	db := p.db.WithContext(ctx).Begin()
	defer db.Commit()

	var messages []Outbox
	res := db.Limit(batchSize).Clauses(
		clause.Locking{
			Strength: clause.LockingStrengthUpdate,
			Options:  clause.LockingOptionsSkipLocked,
		},
	).Find(&messages)
	if res.Error != nil {
		return res.Error
	}

	if len(messages) == 0 {
		return nil
	}

	for _, m := range messages {
		if err := p.publisher.Publish(fmt.Sprintf("outbox.event.%v", m.AggregateType), message.NewMessage(m.ID, m.Payload)); err != nil {
			p.logger.Error("unable to publish message", "error", err)
		}
	}

	if res := db.Delete(&messages); res.Error != nil {
		p.logger.Error("unable to delete message batch")
	}

	return nil

}

type OutboxStorage struct {
	tx        *gorm.DB
	logger    *slog.Logger
	marshaler Marshaler
}

func NewOutbox() func(*gorm.DB) outbox.Storer {
	return func(tx *gorm.DB) outbox.Storer {
		pub := &OutboxStorage{
			tx:     tx,
			logger: slog.Default(),
			marshaler: protojson.MarshalOptions{
				UseProtoNames: true,
				Multiline:     false,
			},
		}

		return pub
	}
}

func (p *OutboxStorage) Store(ctx context.Context, id string, event proto.Message) error {
	jsonBytes, err := p.marshaler.Marshal(event)
	if err != nil {
		return err
	}

	outboxEntry := Outbox{
		AggregateType: string(proto.MessageName(event).Name()),
		AggregateID:   id,
		Payload:       jsonBytes,
	}

	tx := p.tx.WithContext(ctx)
	if res := tx.Create(&outboxEntry); res.Error != nil {
		return err
	}

	return nil
}
