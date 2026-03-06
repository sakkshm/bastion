package api

import (
	"github.com/sakkshm/bastion/internal/engine"
)

type Handler struct {
	Engine *engine.Engine
}

func NewHandler(e *engine.Engine) *Handler {
	return &Handler{
		Engine: e,
	}
}
