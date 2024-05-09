package repository

import "log/slog"

type option func(d *UserRepository)

func WithLogger(l *slog.Logger) option {
	return func(u *UserRepository) {
		u.logger = l
	}
}
