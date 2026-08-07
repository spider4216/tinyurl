package config

import (
	"time"

	"github.com/caarlos0/env/v11"

	"github.com/spider4216/tinyurl/internal/config/db"
)

type Config struct {
	db.DbConfig
	ServerAddress string        `env:"SERVER_ADDRESS" json:"server_address"`           // Адрес запуска HTTP-сервера
	BaseUrl       string        `env:"BASE_URL" json:"base_url"`                       // Базовый адрес результирующего сокращённого URL
	LogLvl        string        `env:"LOG_LEVEL" json:"log_level"`                     // Уровень логирования
	StoreDriver   string        `env:"STORE_DRIVER" json:"store_driver"`               // Драйвер хранилища
	FileStorePath string        `env:"FILE_STORAGE_PATH" json:"file_storage_path"`     // В случае store driver file - путь до файла
	CtxTimeout    time.Duration `env:"CTX_TIMEOUT" envDefault:"3s" json:"ctx_timeout"` // Таймаут контекста в секундах
	ReadTimeout   time.Duration `env:"READ_TIMEOUT" envDefault:"5s" json:"read_timeout"`
	WriteTimeout  time.Duration `env:"WRITE_TIMEOUT" envDefault:"10s" json:"write_timeout"`
	IdleTimeout   time.Duration `env:"IDLE_TIMEOUT" envDefault:"30s" json:"idle_timeout"`
	MaxBodySize   int64         `env:"MAX_BODY_SIZE" envDefault:"2048" json:"max_body_size"`
	CookieTTL     time.Duration `env:"COOKIE_TTL" envDefault:"24h" json:"cookie_ttl"`
	SignKey       string        `env:"SIGN_KEY" json:"sign_key"`                              // Ключ для подписи значения куки
	DeleteMaxPool int           `env:"DEL_MAX_POOL" envDefault:"10" json:"del_max_pool"`      // Кол-во одновременно вып-мых задач на удаление
	AuditFile     string        `env:"AUDIT_FILE" json:"audit_file"`                          // Аудит в файл
	AuditURL      string        `env:"AUDIT_URL" json:"audit_url"`                            // Аудит на сервер по HTTP
	ProfileHost   string        `env:"PROFILE_HOST" envDefault:":6060" json:"profile_host"`   // Хост для профилирования
	CrtPath       string        `env:"CRT_PATH" envDefault:"certs/cert.pem" json:"crt_path"`  // Путь до сертификата для режимо HTTPS
	PKPath        string        `env:"PK_PATH" envDefault:"certs/private.pem" json:"pk_path"` // Путь до приватного ключа для режимо HTTPS
	Https         bool          `env:"ENABLE_HTTPS" json:"enable_https"`                      // Режим HTTPS
	CfgFile       string        `env:"CONFIG"`                                                // Путь до файла конфигурации
}

func New() (*Config, error) {
	cfg := &Config{}

	if err := env.Parse(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
