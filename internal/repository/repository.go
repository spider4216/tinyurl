package repository

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/spider4216/tinyurl/internal/models"
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
	UserId      string `json:"user_id"`
	IsDeleted   string `json:"is_deleted"`
}

func (r *Repository) DeleteByIds(ctx context.Context, ids []string, userId string) error {
	return r.store.DeleteBatch(ctx, ids, userId)
}

func (r *Repository) InsertBatch(ctx context.Context, items []models.InsertData) error {
	raw := [][]byte{}

	for _, item := range items {
		i := map[string]string{
			"uuid":         item.Key,
			"short_url":    item.Key,
			"original_url": item.Value,
			"user_id":      item.UserId,
			"is_deleted":   "false",
		}

		b, err := json.Marshal(i)
		if err != nil {
			return err
		}

		raw = append(raw, b)
	}

	return r.store.SaveBatch(ctx, raw)
}

func (r *Repository) Insert(ctx context.Context, data models.InsertData) error {
	raw := map[string]string{
		"uuid":         data.Key,
		"short_url":    data.Key,
		"original_url": data.Value,
		"user_id":      data.UserId,
		"is_deleted":   "false",
	}

	b, err := json.Marshal(raw)
	if err != nil {
		return err
	}

	return r.store.Save(ctx, b)
}

func (r *Repository) Get(ctx context.Context, k string) (string, error) {
	rows, err := r.store.Load(ctx)
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

func (r *Repository) GetByValue(ctx context.Context, v string) (string, error) {
	rows, err := r.store.Load(ctx)
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

		if rec.OriginalUrl == v {
			return string(item), nil
		}
	}

	return "", errors.New("cannot found item")
}

func (r *Repository) Ping(ctx context.Context) error {
	return r.store.Ping(ctx)
}

func (r *Repository) GetByUserId(ctx context.Context, userId string) ([]models.UrlItem, error) {
	return r.store.GetByUserId(ctx, userId)
}
