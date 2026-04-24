package storage

import (
	"bytes"
	"io"
	"sync"
)

type MapStorage struct {
	store []string
	mu    sync.RWMutex
}

func NewMapStorage() *MapStorage {
	return &MapStorage{
		store: []string{},
	}
}

func (ms *MapStorage) Save(data []byte) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	ms.store = append(ms.store, string(data))

	return nil
}

func (ms *MapStorage) Load() (io.Reader, error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	var buf bytes.Buffer

	for _, line := range ms.store {
		buf.WriteString(line)
	}

	return &buf, nil
}

func (ms *MapStorage) Ping() error {
	return nil
}
