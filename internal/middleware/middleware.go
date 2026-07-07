// Middleware пакет позволяет прогнать запрос через прослойки до момента попадания
// его в Handler. В настоящий момент сервис поддерживает следующие Middlewares:
//   - logger - логирует информацию о запросе
//   - gzip   - сжимает данные перед отправкой клиенту
//   - auth   - аутентификация и авторизация пользователя
package middleware

import (
	"go.uber.org/zap"

	"github.com/spider4216/tinyurl/internal/config"
	"github.com/spider4216/tinyurl/internal/service"
)

// Middleware основной тип Middleware. Расширяется новыми прослойками
type Middleware struct {
	logger  *zap.SugaredLogger
	cfg     *config.Config
	service service.Service
}

// New конструктор Middleware
func New(logger *zap.SugaredLogger, cfg *config.Config, service service.Service) Middleware {
	return Middleware{
		logger:  logger,
		cfg:     cfg,
		service: service,
	}
}
