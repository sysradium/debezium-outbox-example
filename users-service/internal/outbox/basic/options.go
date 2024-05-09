package basic

import (
	"log/slog"
	"time"
)

type workerOption func(d *Worker)

func WithLogger(l *slog.Logger) workerOption {
	return func(d *Worker) {
		d.logger = l
	}
}

func WithPollingInterval(d time.Duration) workerOption {
	return func(w *Worker) {
		w.pollingInterval = d
	}
}

func WithTopicPrefix(prefix string) workerOption {
	return func(w *Worker) {
		w.topicPrefix = prefix
	}
}
