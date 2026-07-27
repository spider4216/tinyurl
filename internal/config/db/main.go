package db

type DbConfig struct {
	DbDsn string `env:"DATABASE_DSN" json:"database_dsn"` // Connection string для БД
}
