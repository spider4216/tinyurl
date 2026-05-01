package db

type DbConfig struct {
	DbDsn string `env:"DATABASE_DSN"` // Connection string для БД
}
