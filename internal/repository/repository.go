package repository

import "sync"

func New(store map[string]string) *Repository {
	return &Repository{
		store: store,
	}
}

type Repository struct {
	store map[string]string
	mu    sync.RWMutex
}

func (r *Repository) Insert(k string, v string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.store[k] = v
}

func (r *Repository) Get(k string) string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	v, ok := r.store[k]

	if !ok {
		return ""
	}

	return v
}
