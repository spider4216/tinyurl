package interceptors

import (
	"go.uber.org/zap"

	"github.com/spider4216/tinyurl/internal/config"
	"github.com/spider4216/tinyurl/internal/service"
)

// Interceptor прослойка для gRPC сервера.
// Методы структуры - это отдельные интерцепторы
type Interceptor struct {
	logger  *zap.SugaredLogger
	cfg     *config.Config
	service service.Service
}

func New(cfg *config.Config, service service.Service, logger *zap.SugaredLogger) *Interceptor {
	return &Interceptor{
		cfg:     cfg,
		service: service,
		logger:  logger,
	}
}
