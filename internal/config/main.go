package config

import (
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/spider4216/tinyurl/internal/config/db"
)

type Config struct {
	db.DbConfig
	ServerAddress string        `env:"SERVER_ADDRESS"`              // Адрес запуска HTTP-сервера
	BaseUrl       string        `env:"BASE_URL"`                    // Базовый адрес результирующего сокращённого URL
	LogLvl        string        `env:"LOG_LEVEL"`                   // Уровень логирования
	StoreDriver   string        `env:"STORE_DRIVER"`                // Драйвер хранилища
	FileStorePath string        `env:"FILE_STORAGE_PATH"`           // В случае store driver file - путь до файла
	CtxTimeout    time.Duration `env:"CTX_TIMEOUT" envDefault:"3s"` // Таймаут контекста в секундах
	ReadTimeout   time.Duration `env:"READ_TIMEOUT" envDefault:"5s"`
	WriteTimeout  time.Duration `env:"WRITE_TIMEOUT" envDefault:"10s"`
	IdleTimeout   time.Duration `env:"IDLE_TIMEOUT" envDefault:"30s"`
	MaxBodySize   int64         `env:"MAX_BODY_SIZE" envDefault:"2048"`
	CookieTTL     time.Duration `env:"COOKIE_TTL" envDefault:"24h"`
	SignKey       string        `env:"SIGN_KEY" envDefault:"qwerty"`    // Ключ для подписи значения куки
	DeleteMaxPool int           `env:"DEL_MAX_POOL" envDefault:"10"`    // Кол-во одновременно вып-мых задач на удаление
	AuditFile     string        `env:"AUDIT_FILE"`                      // Аудит в файл
	AuditURL      string        `env:"AUDIT_URL"`                       // Аудит на сервер по HTTP
	ProfileHost   string        `env:"PROFILE_HOST" envDefault:":6060"` // Хост для профилирования
}

func New() (*Config, error) {
	cfg := &Config{}

	if err := env.Parse(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
