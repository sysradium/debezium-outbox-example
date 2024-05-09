package commands

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
)

type commandLoggingDecorator[C any, R any] struct {
	base   CommandHandler[C, R]
	logger *slog.Logger
}

func (l commandLoggingDecorator[C, R]) Handle(ctx context.Context, cmd C) (R, error) {
	l.logger.Info("executing command", "command", generateActionName(cmd))
	res, err := l.base.Handle(ctx, cmd)
	if err != nil {
		l.logger.Error("unable to process command", "command", generateActionName(cmd), "error", err)
	}
	return res, err
}

func ApplyCommandDecorators[C any, R any](handler CommandHandler[C, R], logger *slog.Logger) CommandHandler[C, R] {
	return commandLoggingDecorator[C, R]{
		base:   handler,
		logger: logger,
	}
}

func generateActionName(handler any) string {
	return strings.Split(fmt.Sprintf("%T", handler), ".")[1]
}
