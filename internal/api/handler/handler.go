package handler

import (
	"ekken/internal/config"
	"ekken/internal/db"
)

type Handler struct {
	Config config.Config
	DB     *db.DB
}

func New(cfg config.Config, database *db.DB) *Handler {
	return &Handler{
		Config: cfg,
		DB:     database,
	}
}
