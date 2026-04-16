package storage

import (
	"fmt"

	"github.com/spider4216/tinyurl/internal/config"
)

const (
	FileDriver = "file"
	MapDriver  = "map"
)

type record struct {
	Key         string `json:"uuid"`
	ShortUrl    string `json:"short_url"`
	OriginalUrl string `json:"original_url"`
}

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
	case MapDriver:
		mapStore := NewMapStorage()
		return &mapStore, nil
	}

	return nil, fmt.Errorf("unsupported driver %s", driver)
}
