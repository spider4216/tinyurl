package middleware

import (
	"github.com/spider4216/tinyurl/internal/config"
	"go.uber.org/zap"
)

type Middleware struct {
	logger *zap.SugaredLogger
	cfg    *config.Config
}

func New(logger *zap.SugaredLogger, cfg *config.Config) Middleware {
	return Middleware{
		logger: logger,
		cfg:    cfg,
	}
}
