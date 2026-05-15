package storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"sync"

	"github.com/spider4216/tinyurl/internal/models"
)

type recordSlice struct {
	Origin    string `json:"original_url"`
	Short     string `json:"short_url"`
	UserId    string `json:"user_id"`
	IsDeleted string `json:"is_deleted"`
}

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

// Здесь реализована внутренняя специфика
// связанная с конкретным хранилищем, в данном случае с слайсом, а именно
// специфика условий для удаления из слайса
func (ms *MapStorage) DeleteBatch(ctx context.Context, ids []string, userId string) error {
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

func (ms *MapStorage) GetByUserId(ctx context.Context, userId string) ([]models.UrlItem, error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	var urls []models.UrlItem

	for _, item := range ms.store {
		var line recordSlice

		if err := json.Unmarshal([]byte(item), &line); err != nil {
			return nil, err
		}

		if line.UserId != userId {
			continue
		}

		b, err := strconv.ParseBool(line.IsDeleted)

		if err != nil {
			return nil, err
		}

		urls = append(urls, models.UrlItem{
			OriginarUrl: line.Origin,
			ShortUrl:    line.Short,
			IsDeleted:   b,
			UserId:      line.UserId,
		})
	}

	return urls, nil
}

func (ms *MapStorage) GetByOrigin(ctx context.Context, origin string) (*models.UrlItem, error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	for _, item := range ms.store {
		var line recordSlice

		if err := json.Unmarshal([]byte(item), &line); err != nil {
			return nil, err
		}

		if line.Origin != origin {
			continue
		}

		b, err := strconv.ParseBool(line.IsDeleted)

		if err != nil {
			return nil, err
		}

		return &models.UrlItem{
			OriginarUrl: line.Origin,
			ShortUrl:    line.Short,
			IsDeleted:   b,
			UserId:      line.UserId,
		}, nil
	}

	return nil, errors.New("cannot find url")
}
