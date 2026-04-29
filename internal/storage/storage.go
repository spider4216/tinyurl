package storage

import (
	"context"
	"fmt"

	"github.com/spider4216/tinyurl/internal/config"
)

const (
	FileDriver     = "file"
	MapDriver      = "map"
	PostgresDriver = "pgx"
)

type Storage interface {
	Save(ctx context.Context, data []byte) error
	SaveBatch(ctx context.Context, data [][]byte) error
	Load(ctx context.Context) (Iterator, error)
	Ping(ctx context.Context) error
	Source() any
	StoreName() string
}

type Iterator interface {
	Next() bool
	Row() ([]byte, error)
	Err() error
	Close() error
}

func New(driver string, cfg *config.Config) (Storage, error) {
	switch driver {
	case FileDriver:
		fileStore, err := NewFileStorage(cfg.FileStorePath)

		if err != nil {
			return nil, err
		}

		return fileStore, nil
	case MapDriver:
		mapStore := NewMapStorage()
		return mapStore, nil
	case PostgresDriver:
		pgxStore, err := NewPgxStorage(cfg.DbDsn)

		if err != nil {
			return nil, err
		}

		return pgxStore, nil
	}

	return nil, fmt.Errorf("unsupported driver %s", driver)
}
