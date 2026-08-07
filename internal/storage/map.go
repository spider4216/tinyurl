package storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"sync"

	"go.uber.org/zap"

	"github.com/spider4216/tinyurl/internal/models"
)

type recordSlice struct {
	Origin    string `json:"original_url"`
	Short     string `json:"short_url"`
	UserId    string `json:"user_id"`
	IsDeleted string `json:"is_deleted"`
}

// MapStorage хранилище, где данные складываются в Slice.
type MapStorage struct {
	store  []string
	mu     sync.RWMutex
	logger *zap.SugaredLogger
}

// NewMapStorage создание хранилища основанном на Slice.
func NewMapStorage(logger *zap.SugaredLogger) *MapStorage {
	return &MapStorage{
		store:  []string{},
		logger: logger,
	}
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

func (ms *MapStorage) GetByShort(ctx context.Context, short string) (*models.UrlItem, error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	for _, item := range ms.store {
		var line recordSlice

		if err := json.Unmarshal([]byte(item), &line); err != nil {
			return nil, err
		}

		if line.Short != short {
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

func (ms *MapStorage) CreateUrl(ctx context.Context, data models.InsertData) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	item := recordSlice{
		Origin:    data.Value,
		Short:     data.Key,
		UserId:    data.UserId,
		IsDeleted: "false",
	}

	b, err := json.Marshal(item)
	if err != nil {
		return err
	}

	select {
	case <-ctx.Done():
		return fmt.Errorf("timeout in slice save")
	default:
		ms.store = append(ms.store, string(b))
	}

	ms.logger.Debug("Added, now store: ", ms.store)

	return nil
}

func (ms *MapStorage) CreateUrls(ctx context.Context, data []models.InsertData) error {
	for _, item := range data {
		if err := ms.CreateUrl(ctx, item); err != nil {
			return err
		}
	}

	return nil
}

// CountUsers количество пользователей в сервисе
func (ms *MapStorage) CountUsers(ctx context.Context) (int, error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	var res []string

	for _, item := range ms.store {
		var line recordSlice

		if err := json.Unmarshal([]byte(item), &line); err != nil {
			return 0, err
		}

		if line.IsDeleted == "true" {
			continue
		}

		res = append(res, line.UserId)
	}

	ms.logger.Debug("Count users before unique", len(res))

	slices.Sort(res)
	res = slices.Compact(res)

	ms.logger.Debug("Count users after unique", len(res))

	return len(res), nil
}

// CountUrls количество сокращённых URL в сервисе
func (ms *MapStorage) CountUrls(ctx context.Context) (int, error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	var res []recordSlice

	for _, item := range ms.store {
		var line recordSlice

		if err := json.Unmarshal([]byte(item), &line); err != nil {
			return 0, err
		}

		if line.IsDeleted == "true" {
			continue
		}

		res = append(res, line)
	}

	return len(res), nil
}
