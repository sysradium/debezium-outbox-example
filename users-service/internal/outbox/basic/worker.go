package basic

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/ThreeDotsLabs/watermill/message"
)

type Publisher interface {
	Publish(topic string, messages ...*message.Message) error
	Close() error
}

type Worker struct {
	db              *sql.DB
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

func NewWorker(db *sql.DB, publisher message.Publisher, opts ...workerOption) *Worker {
	ctx, cancel := context.WithCancel(context.Background())
	w := &Worker{
		db:        db,
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

func (p *Worker) Start() error {
	if _, err := p.db.Exec(createSQLStatement); err != nil {
		return err
	}

	ticker := time.NewTicker(p.pollingInterval)
	defer func() {
		ticker.Stop()
		close(p.done)
	}()

	for {
		select {
		case <-p.ctx.Done():
			return nil
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
	db, err := p.db.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			_ = db.Rollback()
			panic(p)
		}
		if rErr != nil {
			xerr := db.Rollback()
			if xerr != nil {
				rErr = errors.Join(xerr, rErr)
			}
			return
		}
		rErr = db.Commit()
	}()

	rows, err := db.Query("SELECT id, aggregatetype, aggregateid, attempts, payload FROM \"my-outbox\" LIMIT 1000 FOR UPDATE SKIP LOCKED")
	if err != nil {
		return err
	}

	var messages []Outbox
	for rows.Next() {
		var msg Outbox
		if err := rows.Scan(&msg.ID, &msg.AggregateType, &msg.AggregateID, &msg.Attempts, &msg.Payload); err != nil {
			return err
		}
		messages = append(messages, msg)
	}

	if err := rows.Err(); err != nil {
		return err
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
		_, err := db.ExecContext(ctx, "DELETE FROM \"my-outbox\" WHERE id = $1", m.ID)
		if err != nil {
			return fmt.Errorf("unable to remove event from DB: %w", err)
		}
	}

	return nil
}
