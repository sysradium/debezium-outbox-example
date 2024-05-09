package app

import (
	"log/slog"

	"github.com/sysradium/debezium-outbox-example/users-service/internal/app/commands"
	"github.com/sysradium/debezium-outbox-example/users-service/internal/domain"
	"github.com/sysradium/debezium-outbox-example/users-service/internal/repository"
)

type Application struct {
	Commands struct {
		CreateUser commands.CreateUserHandler
	}
}

func NewApplication(
	userRepo repository.Repository[domain.User],

) Application {
	a := Application{}
	a.Commands.CreateUser = commands.ApplyCommandDecorators[commands.CreateUser](
		commands.NewUserCreateHandler(userRepo),
		slog.Default(),
	)
	return a
}
