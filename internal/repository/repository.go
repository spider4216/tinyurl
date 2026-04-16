package repository

import (
	"encoding/json"

	"github.com/spider4216/tinyurl/internal/storage"
)

func New(store storage.Storage) *Repository {
	return &Repository{
		store: store,
	}
}

type Repository struct {
	store storage.Storage
}

func (r *Repository) Insert(key string, val string) error {
	raw := map[string]string{
		"short_url":    key,
		"original_url": val,
	}

	b, err := json.Marshal(raw)

	if err != nil {
		return err
	}

	return r.store.Save(key, b)
}

func (r *Repository) Get(k string) (string, error) {
	b, err := r.store.Load(k)

	if err != nil {
		return "", err
	}

	return string(b), nil
}
