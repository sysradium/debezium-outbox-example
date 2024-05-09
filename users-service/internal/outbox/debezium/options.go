package debezium

import "log/slog"

type option func(d *OutboxPublisher)

func WithLogger(l *slog.Logger) option {
	return func(d *OutboxPublisher) {
		d.logger = l
	}
}
