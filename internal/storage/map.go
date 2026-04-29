package storage

import (
	"context"
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

type SliceIterator struct {
	data []string
	idx  int
}

func (i *SliceIterator) Next() bool {
	if i.idx >= len(i.data) {
		return false
	}

	i.idx++

	return true
}

func (i *SliceIterator) Row() ([]byte, error) {
	v := i.data[i.idx-1]

	return []byte(v), nil
}

func (i *SliceIterator) Err() error {
	return nil
}

func (i *SliceIterator) Close() error {
	return nil
}

func (ms *MapStorage) SaveBatch(ctx context.Context, data [][]byte) error {
	for _, item := range data {
		ms.Save(ctx, item)
	}

	return nil
}

func (ms *MapStorage) Save(ctx context.Context, data []byte) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	ms.store = append(ms.store, string(data))

	return nil
}

func (ms *MapStorage) Load(ctx context.Context) (Iterator, error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	return &SliceIterator{
		data: ms.store,
		idx:  0,
	}, nil
}

func (ms *MapStorage) Ping(ctx context.Context) error {
	return nil
}

func (ms *MapStorage) Source() any {
	return nil
}

func (ms *MapStorage) StoreName() string {
	return MapDriver
}
