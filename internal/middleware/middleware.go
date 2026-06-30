package middleware

import (
	"go.uber.org/zap"

	"github.com/spider4216/tinyurl/internal/config"
	"github.com/spider4216/tinyurl/internal/service"
)

type Middleware struct {
	logger  *zap.SugaredLogger
	cfg     *config.Config
	service service.Service
}

func New(logger *zap.SugaredLogger, cfg *config.Config, service service.Service) Middleware {
	return Middleware{
		logger:  logger,
		cfg:     cfg,
		service: service,
	}
}
