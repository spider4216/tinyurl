package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
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
		if err := ms.Save(ctx, item); err != nil {
			return err
		}
	}

	return nil
}

func (ms *MapStorage) Save(ctx context.Context, data []byte) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	select {
	case <-ctx.Done():
		return fmt.Errorf("timeout in slice save")
	default:
		ms.store = append(ms.store, string(data))
	}

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

func (ms *MapStorage) DeleteBatch(ctx context.Context, ids []string, userId string) error {
	// Здесь реализована внутренняя специфика
	// связанная с конкретным хранилищем, в данном случае с слайсом, а именно
	// специфика условий для удаления из слайса

	ms.mu.Lock()
	defer ms.mu.Unlock()

	for i, item := range ms.store {
		line := map[string]string{}

		if err := json.Unmarshal([]byte(item), &line); err != nil {
			return err
		}

		sourceId, ok := line["user_id"]

		if !ok {
			continue
		}

		short, ok := line["short_url"]

		if !ok {
			continue
		}

		if sourceId == userId && slices.Contains(ids, short) {
			// Обновляем is_deleted
			line["is_deleted"] = "true"
		}

		updLine, err := json.Marshal(line)

		if err != nil {
			return err
		}

		ms.store[i] = string(updLine)
	}

	return nil
}
