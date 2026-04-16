package storage

import (
	"encoding/json"
	"os"
	"sync"
)

type MapStorage struct {
	store []map[string]string
	mu    sync.RWMutex
}

func NewMapStorage() MapStorage {
	return MapStorage{
		store: []map[string]string{},
	}
}

func (ms *MapStorage) Save(key string, data []byte) error {
	record := record{}

	if err := json.Unmarshal(data, &record); err != nil {
		return err
	}

	record.Key = key

	raw := map[string]string{
		"uuid":         key,
		"short_url":    record.ShortUrl,
		"original_url": record.OriginalUrl,
	}

	ms.mu.Lock()
	defer ms.mu.Unlock()

	ms.store = append(ms.store, raw)

	return nil
}

func (ms *MapStorage) Load(key string) ([]byte, error) {
	for _, r := range ms.store {
		k, ok := r["uuid"]

		if !ok {
			continue
		}

		if k == key {
			b, err := json.Marshal(r)

			if err != nil {
				return nil, err
			}

			return b, nil
		}
	}

	return nil, os.ErrNotExist
}
