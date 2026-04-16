package storage

import (
	"fmt"

	"github.com/spider4216/tinyurl/internal/config"
)

const (
	FileDriver = "file"
)

type Storage interface {
	Save(key string, data []byte) error
	Load(key string) ([]byte, error)
}

func New(driver string, cfg config.Config) (Storage, error) {
	switch driver {
	case FileDriver:
		fileStore, err := NewFileStorage(cfg.FileStorePath)

		if err != nil {
			return nil, err
		}

		return fileStore, nil

	}

	return nil, fmt.Errorf("unsupported driver %s", driver)
}
