package handlers

import (
	"hafh-server/internal/database"

	"go.uber.org/zap"
)

type handlerConfig struct {
	db  *database.Database
	log *zap.SugaredLogger
}

var config *handlerConfig

// Init initializes the handler configuration with the provided database and logger.
func Init(db *database.Database, log *zap.SugaredLogger) {
	config = &handlerConfig{
		db:  db,
		log: log,
	}
}
