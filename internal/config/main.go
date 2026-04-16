package config

import "github.com/caarlos0/env/v11"

type Config struct {
	ServerAddress string `env:"SERVER_ADDRESS"` // Адрес запуска HTTP-сервера
	BaseUrl       string `env:"BASE_URL"`       // Базовый адрес результирующего сокращённого URL
	LogLvl        string `env:"LOG_LEVEL"`      // Уровень логирования
	StoreDriver   string `env:"STORE_DRIVER"`   // Драйвер хранилища
}

func New() Config {
	cfg := Config{}
	env.Parse(&cfg)

	return cfg
}
