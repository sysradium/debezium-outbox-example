package basic

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/ThreeDotsLabs/watermill/message"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Publisher interface {
	Publish(topic string, messages ...*message.Message) error
	Close() error
}

type Worker struct {
	db              *gorm.DB
	logger          *slog.Logger
	pollingInterval time.Duration
	publisher       Publisher
	batchSize       int
	topicPrefix     string
	done            chan struct{}
	ctx             context.Context
	cancel          context.CancelFunc
	tableName       string
}

func NewWorker(db *gorm.DB, publisher message.Publisher, opts ...workerOption) *Worker {
	ctx, cancel := context.WithCancel(context.Background())
	w := &Worker{
		db:        db.Table("my-outbox"),
		logger:    slog.Default(),
		publisher: publisher,
		batchSize: 10000,
		ctx:       ctx,
		cancel:    cancel,
		tableName: "my-outbox",
		done:      make(chan struct{}),
	}

	for _, o := range opts {
		o(w)
	}

	return w
}

func (p *Worker) Start() {
	p.db.Table(p.tableName).AutoMigrate(&Outbox{})
	ticker := time.NewTicker(p.pollingInterval)
	defer func() {
		ticker.Stop()
		close(p.done)
	}()

	for {
		select {
		case <-p.ctx.Done():
			return
		case <-ticker.C:
			if err := p.processMessages(p.ctx, p.batchSize); err != nil {
				p.logger.Error("failed to process messages", "error", err)
			}
		}
	}

}

func (p *Worker) Stop() {
	p.publisher.Close()
	p.cancel()
	<-p.done
}

func (p *Worker) processMessages(ctx context.Context, batchSize int) (rErr error) {
	db := p.db.WithContext(ctx).Begin()
	defer func() {
		if p := recover(); p != nil {
			_ = db.Rollback()
			panic(p)
		}
		if rErr != nil {
			xerr := db.Rollback().Error
			if xerr != nil {
				rErr = fmt.Errorf("%s: %w", xerr.Error(), rErr)
			}
			return
		}
		rErr = db.Commit().Error
	}()

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
		if err := p.publisher.Publish(
			fmt.Sprintf("%v.%v", p.topicPrefix, m.AggregateType),
			message.NewMessage(m.ID, m.Payload),
		); err != nil {
			return fmt.Errorf("unable to publish event: %w", err)
		}
	}

	if res := db.Delete(&messages); res.Error != nil {
		p.logger.Error("unable to delete message batch")
	}

	return nil
}
