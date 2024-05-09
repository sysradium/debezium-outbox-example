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

type Worker struct {
	db              *gorm.DB
	logger          *slog.Logger
	pollingInterval time.Duration
	publisher       message.Publisher
	batchSize       int
	topicPrefix     string
	done            chan struct{}
	ctx             context.Context
	cancel          context.CancelFunc
}

func NewWorker(db *gorm.DB, publisher message.Publisher, opts ...workerOption) *Worker {
	ctx, cancel := context.WithCancel(context.Background())
	w := &Worker{
		db:        db,
		logger:    slog.Default(),
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
	defer func() {
		ticker.Stop()
		p.done <- struct{}{}
	}()

	for {
		select {
		case <-p.ctx.Done():
			return
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
		if err := p.publisher.Publish(
			fmt.Sprintf("%v.event.%v", p.topicPrefix, m.AggregateType),
			message.NewMessage(m.ID, m.Payload),
		); err != nil {
			p.logger.Error("unable to publish message", "error", err)
		}
	}

	if res := db.Delete(&messages); res.Error != nil {
		p.logger.Error("unable to delete message batch")
	}

	return nil
}
