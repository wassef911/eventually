package commands

import (
	"context"

	"github.com/wassef911/astore/internal/infrastructure/es/store"
	"github.com/wassef911/astore/pkg/config"
	"github.com/wassef911/astore/pkg/logger"
)

type commandHandler[T any] interface {
	Handle(ctx context.Context, command T) error
}
type baseCommandHandler struct {
	log logger.Logger
	cfg *config.Config
	es  store.AggregateStore
}
