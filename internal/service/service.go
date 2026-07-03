package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"net/url"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"go.uber.org/zap"

	"github.com/spider4216/tinyurl/internal/audit"
	"github.com/spider4216/tinyurl/internal/models"
	"github.com/spider4216/tinyurl/internal/repository"
)

const (
	ChankSize = 2 // Размер пачки с URL для удаления.
)

// Service основной сервис проекта.
type Service struct {
	repo   *repository.Repository
	logger *zap.SugaredLogger
	audit  audit.Publisher
}

// New конструктор сервиса.
func New(repo *repository.Repository, logger *zap.SugaredLogger, audit audit.Publisher) Service {
	return Service{
		repo:   repo,
		logger: logger,
		audit:  audit,
	}
}

// DeletedUrlError пользовательский тип ошибки, вызывать когда URL удален.
type DeletedUrlError struct{}

// Error реализация метода интерфейса, возвращает кастом сообщение.
func (e DeletedUrlError) Error() string {
	return "Url was deleted"
}

// GenerateId генерирует уникальный ID для оригинального URL.
func (s Service) GenerateId() string {
	key := make([]byte, 6)
	rand.Read(key)

	return base64.URLEncoding.EncodeToString(key)
}

// StoreData создает короткий URL на основе оригинального.
func (s Service) StoreData(ctx context.Context, id string, val string, userId string) error {
	in := s.repo.MapInserData(id, val, userId)

	return s.repo.Insert(ctx, in)
}

// DeleteBatch удаляет множество URL (пачкой).
func (s Service) DeleteBatch(ctx context.Context, ids []string, userId string) error {
	return s.repo.DeleteByIds(ctx, ids, userId)
}

// DeleteBatchAsync асинхронный вариант удаления мнодества URL.
func (s Service) DeleteBatchAsync(ctx context.Context, ids []string, userId string) {
	// канал с данными
	inputCh := s.GenerateChunk(ids, ChankSize, userId)

	// получаем слайс каналов
	channels := s.FanOutDeleteBatch(ctx, inputCh)

	// а теперь объединяем каналы в один
	resCh := s.FanInDeleteBatch(channels)

	// логируем результаты расчетов из канала
	for res := range resCh {
		s.logger.Debug("Result chank delete received")

		if res.Err != nil {
			s.logger.Error("Chank delete error ", zap.Error(res.Err), res.IDs)
		}
	}
}

// StoreDataBatch создания мнодества URL.
func (s Service) StoreDataBatch(ctx context.Context, urls []UrlsIds) error {
	in := []models.InsertData{}

	for _, item := range urls {
		i := models.InsertData{
			Key:    item.Short,
			Value:  item.Origin,
			UserId: item.UserId,
		}

		in = append(in, i)
	}

	return s.repo.InsertBatch(ctx, in)
}

// GetUrl получение оригинального URL по сокращенному.
func (s Service) GetUrl(ctx context.Context, k string) (string, error) {
	item, err := s.repo.GetByShort(ctx, k)
	if err != nil {
		return "", err
	}

	if item.IsDeleted {
		return "", DeletedUrlError{}
	}

	return item.OriginarUrl, nil
}

// GetShortByOrigin получение сокращенного URL по оригинальному.
func (s Service) GetShortByOrigin(ctx context.Context, v string) (string, error) {
	item, err := s.repo.GetByOrigin(ctx, v)
	if err != nil {
		return "", err
	}

	return item.ShortUrl, nil
}

// GetUrlsByUserId получение всех оригинальных URL пользователя.
func (s Service) GetUrlsByUserId(ctx context.Context, userId string, baseUrl string) ([]models.UrlItem, error) {
	items, err := s.repo.GetByUserId(ctx, userId)
	if err != nil {
		return nil, err
	}

	for i := range items {
		fullUrl, err := url.JoinPath(baseUrl, items[i].ShortUrl)
		if err != nil {
			return nil, err
		}

		items[i].ShortUrl = fullUrl
	}

	return items, nil
}

// Ping проверка доступности источника данных.
func (s Service) Ping(ctx context.Context) error {
	return s.repo.Ping(ctx)
}

// IsErrAsDuplicate проверка, является ли переданная ошибка типум дубликата.
func (s Service) IsErrAsDuplicate(err error) bool {
	var pgErr *pgconn.PgError

	if !errors.As(err, &pgErr) {
		return false
	}

	return pgErr.Code == pgerrcode.UniqueViolation
}

// AuditNotify метод аудита. Оповещает всех подписчиков о событии аудита.
func (s Service) AuditNotify(action audit.AuditAction, userID string, url string) error {
	s.logger.Debug("Audit notify with action ", action)

	n := audit.Body{
		TS:     time.Now().Unix(),
		Action: action,
		UserID: userID,
		URL:    url,
	}

	return s.audit.Notify(n)
}
