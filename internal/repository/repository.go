package repository

import (
	"encoding/json"
	"errors"

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

type record struct {
	Key         string `json:"uuid"`
	ShortUrl    string `json:"short_url"`
	OriginalUrl string `json:"original_url"`
}

func (r *Repository) Insert(key string, val string) error {
	raw := map[string]string{
		"uuid":         key,
		"short_url":    key,
		"original_url": val,
	}

	b, err := json.Marshal(raw)

	if err != nil {
		return err
	}

	return r.store.Save(b)
}

func (r *Repository) Get(k string) (string, error) {
	rows, err := r.store.Load()

	if err != nil {
		return "", err
	}

	for rows.Next() {
		item, err := rows.Row()

		if err != nil {
			return "", err
		}

		rec := record{}

		if err := json.Unmarshal([]byte(item), &rec); err != nil {
			continue
		}

		if rec.Key == k {
			return string(item), nil
		}
	}

	return "", errors.New("cannot found item")
}

func (r *Repository) Ping() error {
	return r.store.Ping()
}
