package handlers

import (
	"database/sql"

	"go.uber.org/zap"
)

type Handlers struct {
	db     *sql.DB
	logger *zap.SugaredLogger
}

func New(db *sql.DB, logger *zap.SugaredLogger) *Handlers {
	return &Handlers{
		db:     db,
		logger: logger,
	}
}
