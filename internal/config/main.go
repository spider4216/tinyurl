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
}

func New() (*Config, error) {
	cfg := &Config{}

	if err := env.Parse(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
