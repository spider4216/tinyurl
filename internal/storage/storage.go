package storage

import (
	"fmt"
)

const (
	MapDriver  = "map"
	FileDriver = "file"
)

type Storage interface {
	Save(key string, data []byte) error
	Load(key string) ([]byte, error)
}

func New(driver string) (Storage, error) {
	switch driver {
	case MapDriver:
		mapStore := NewMapStorage()
		return &mapStore, nil
	}

	return nil, fmt.Errorf("unsupported driver %s", driver)
}
