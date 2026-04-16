package repository

import (
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

func (r *Repository) Insert(k string, v string) error {
	return r.store.Save(k, []byte(v))
}

func (r *Repository) Get(k string) (string, error) {
	b, err := r.store.Load(k)

	if err != nil {
		return "", err
	}

	return string(b), nil
}
