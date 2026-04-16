package storage

import (
	"errors"
	"sync"
)

type MapStorage struct {
	store map[string][]byte
	mu    sync.RWMutex
}

func NewMapStorage() MapStorage {
	return MapStorage{
		store: map[string][]byte{},
	}
}

func (ms *MapStorage) Save(key string, data []byte) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	ms.store[key] = data

	return nil
}

func (ms *MapStorage) Load(key string) ([]byte, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	v, ok := ms.store[key]

	if !ok {
		return nil, errors.New("Cannot fined value")
	}

	return v, nil
}
