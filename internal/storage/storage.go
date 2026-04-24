package storage

import (
	"fmt"
	"io"

	"github.com/spider4216/tinyurl/internal/config"
)

const (
	FileDriver     = "file"
	MapDriver      = "map"
	PostgresDriver = "pgx"
)

type Storage interface {
	Save(data []byte) error
	Load() (io.Reader, error)
	Ping() error
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
