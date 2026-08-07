// Слой репозитория сервиса по сокращению URL
package repository

import (
	"context"

	"github.com/spider4216/tinyurl/internal/models"
	"github.com/spider4216/tinyurl/internal/storage"
)

// New конструктор репощитория.
func New(store storage.Storage) *Repository {
	return &Repository{
		store: store,
	}
}

// Repository тип репозитория.
type Repository struct {
	store storage.Storage
}

// DeleteByIds удаление сокращенных URL по идентификаторам.
func (r *Repository) DeleteByIds(ctx context.Context, ids []string, userId string) error {
	return r.store.DeleteBatch(ctx, ids, userId)
}

// InsertBatch создание множества сокращенных URL.
func (r *Repository) InsertBatch(ctx context.Context, items []models.InsertData) error {
	return r.store.CreateUrls(ctx, items)
}

// Insert создание сокращенного URL.
func (r *Repository) Insert(ctx context.Context, data models.InsertData) error {
	return r.store.CreateUrl(ctx, data)
}

// GetByShort получение оригинального URL по сокращенному.
func (r *Repository) GetByShort(ctx context.Context, k string) (*models.UrlItem, error) {
	return r.store.GetByShort(ctx, k)
}

// GetByOrigin получение сокращенного URL по оригинальному.
func (r *Repository) GetByOrigin(ctx context.Context, v string) (*models.UrlItem, error) {
	return r.store.GetByOrigin(ctx, v)
}

// Ping проверка доступности источника данных.
func (r *Repository) Ping(ctx context.Context) error {
	return r.store.Ping(ctx)
}

// GetByUserId получение ссылок пользователя по его идентификатору
func (r *Repository) GetByUserId(ctx context.Context, userId string) ([]models.UrlItem, error) {
	return r.store.GetByUserId(ctx, userId)
}

// CountUsers возвращает кол-во униникальных пользователей
func (r *Repository) CountUsers(ctx context.Context) (int, error) {
	return r.store.CountUsers(ctx)
}

// CountUrls возвращает общее кол-во сокращенных URL
func (r *Repository) CountUrls(ctx context.Context) (int, error) {
	return r.store.CountUrls(ctx)
}
