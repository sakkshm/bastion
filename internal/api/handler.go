package api

import (
	"log/slog"

	"github.com/sakkshm/bastion/internal/config"
	"github.com/sakkshm/bastion/internal/engine"
)

type Handler struct {
	Engine *engine.Engine
	Logger *slog.Logger
	Config *config.Config
}

func NewHandler(e *engine.Engine, cfg *config.Config, logger *slog.Logger) *Handler {
	return &Handler{
		Engine: e,
		Config: cfg,
		Logger: logger,
	}
}
