package repository

import (
	"context"
	"encoding/json"

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
	return r.store.CreateUrl(ctx, data)
}

func (r *Repository) GetByShort(ctx context.Context, k string) (*models.UrlItem, error) {
	return r.store.GetByShort(ctx, k)
}

func (r *Repository) GetByOrigin(ctx context.Context, v string) (*models.UrlItem, error) {
	return r.store.GetByOrigin(ctx, v)
}

func (r *Repository) Ping(ctx context.Context) error {
	return r.store.Ping(ctx)
}

func (r *Repository) GetByUserId(ctx context.Context, userId string) ([]models.UrlItem, error) {
	return r.store.GetByUserId(ctx, userId)
}
