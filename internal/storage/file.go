package storage

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"slices"
	"sync"

	"go.uber.org/zap"

	"github.com/spider4216/tinyurl/internal/models"
)

// FileStorage хранит данные в файле.
type FileStorage struct {
	file   *os.File
	mu     sync.RWMutex
	logger *zap.SugaredLogger
}

// NewFileStorage функция создания хранилища на файле.
func NewFileStorage(filename string, logger *zap.SugaredLogger) (*FileStorage, error) {
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0o666)
	if err != nil {
		return nil, err
	}

	return &FileStorage{
		file:   file,
		logger: logger,
	}, nil
}

type recordFile struct {
	Origin    string `json:"original_url"`
	Short     string `json:"short_url"`
	UserId    string `json:"user_id"`
	IsDeleted bool   `json:"is_deleted"`
}

func (fs *FileStorage) Ping(ctx context.Context) error {
	return nil
}

func (fs *FileStorage) Source() any {
	return nil
}

func (fs *FileStorage) StoreName() string {
	return FileDriver
}

func (fs *FileStorage) DeleteBatch(ctx context.Context, ids []string, userId string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	// Вернуть курсор вначало
	if _, err := fs.file.Seek(0, 0); err != nil {
		return err
	}

	scanner := bufio.NewScanner(fs.file)

	// Создать временный файл, туда будет записываться
	// измененные данные
	tmpFile, err := os.CreateTemp("", "storage-tmp.json")
	if err != nil {
		return err
	}

	// Удаляем временный файл после завершения обновления
	defer func() {
		if rmErr := os.Remove(tmpFile.Name()); rmErr != nil {
			fs.logger.Warn("cannot remove tmp file", zap.Error(rmErr))
		}
	}()

	// Произвожу поиск
	for scanner.Scan() {
		rec := recordFile{}

		line := scanner.Bytes()

		if umErr := json.Unmarshal(line, &rec); umErr != nil {
			continue
		}

		if rec.UserId == userId && slices.Contains(ids, rec.Short) {
			// Обновляем is_deleted
			rec.IsDeleted = true
		}

		updLine, merr := json.Marshal(rec)
		if merr != nil {
			return merr
		}

		if _, wErr := tmpFile.Write(updLine); wErr != nil {
			return wErr
		}

		if _, wErr := tmpFile.Write([]byte("\n")); wErr != nil {
			return wErr
		}
	}

	if scanErr := scanner.Err(); scanErr != nil {
		return scanErr
	}

	originPath := fs.file.Name()

	// Закрываем старый файл
	if closeErr := fs.file.Close(); closeErr != nil {
		return closeErr
	}

	// Заменяем файл
	if renErr := os.Rename(tmpFile.Name(), originPath); renErr != nil {
		return renErr
	}

	// Открываем заново
	file, err := os.OpenFile(originPath, os.O_RDWR|os.O_CREATE, 0o666)
	if err != nil {
		return err
	}

	fs.file = file

	return nil
}

func (fs *FileStorage) GetByUserId(ctx context.Context, userId string) ([]models.UrlItem, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	// Вернуть курсор вначало
	if _, err := fs.file.Seek(0, 0); err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(fs.file)

	var urls []models.UrlItem

	for scanner.Scan() {
		var item recordFile

		if err := json.Unmarshal(scanner.Bytes(), &item); err != nil {
			continue
		}

		if item.UserId != userId {
			continue
		}

		urls = append(urls, models.UrlItem{
			OriginarUrl: item.Origin,
			ShortUrl:    item.Short,
			IsDeleted:   item.IsDeleted,
			UserId:      item.UserId,
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return urls, nil
}

func (fs *FileStorage) GetByOrigin(ctx context.Context, origin string) (*models.UrlItem, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	// Вернуть курсор вначало
	if _, err := fs.file.Seek(0, 0); err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(fs.file)

	for scanner.Scan() {
		var item recordFile

		if err := json.Unmarshal(scanner.Bytes(), &item); err != nil {
			continue
		}

		if item.Origin != origin {
			continue
		}

		return &models.UrlItem{
			OriginarUrl: item.Origin,
			ShortUrl:    item.Short,
			IsDeleted:   item.IsDeleted,
			UserId:      item.UserId,
		}, nil
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return nil, errors.New("cannot find url")
}

func (fs *FileStorage) GetByShort(ctx context.Context, short string) (*models.UrlItem, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	// Вернуть курсор вначало
	if _, err := fs.file.Seek(0, 0); err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(fs.file)

	for scanner.Scan() {
		var item recordFile

		if err := json.Unmarshal(scanner.Bytes(), &item); err != nil {
			continue
		}

		if item.Short != short {
			continue
		}

		return &models.UrlItem{
			OriginarUrl: item.Origin,
			ShortUrl:    item.Short,
			IsDeleted:   item.IsDeleted,
			UserId:      item.UserId,
		}, nil
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return nil, errors.New("cannot find url")
}

func (fs *FileStorage) CreateUrl(ctx context.Context, data models.InsertData) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	item := recordFile{
		Origin:    data.Value,
		Short:     data.Key,
		UserId:    data.UserId,
		IsDeleted: false,
	}

	b, err := json.Marshal(item)
	if err != nil {
		return err
	}

	select {
	case <-ctx.Done():
		return fmt.Errorf("timeout in file save")
	default:
		b = append(b, '\n')

		if _, err := fs.file.Write(b); err != nil {
			return err
		}

		return nil
	}
}

func (fs *FileStorage) CreateUrls(ctx context.Context, data []models.InsertData) error {
	for _, item := range data {
		if err := fs.CreateUrl(ctx, item); err != nil {
			return err
		}
	}

	return nil
}

// CountUsers количество пользователей в сервисе.
func (fs *FileStorage) CountUsers(ctx context.Context) (int, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	// Вернуть курсор вначало
	if _, err := fs.file.Seek(0, 0); err != nil {
		return 0, err
	}

	scanner := bufio.NewScanner(fs.file)

	var res []string

	for scanner.Scan() {
		var item recordFile

		if err := json.Unmarshal(scanner.Bytes(), &item); err != nil {
			continue
		}

		if item.IsDeleted {
			continue
		}

		res = append(res, item.UserId)
	}

	if err := scanner.Err(); err != nil {
		return 0, err
	}

	fs.logger.Debug("Count users before unique", len(res))

	slices.Sort(res)
	res = slices.Compact(res)

	fs.logger.Debug("Count users after unique", len(res))

	return len(res), nil
}

// CountUrls количество сокращённых URL в сервисе.
func (fs *FileStorage) CountUrls(ctx context.Context) (int, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	// Вернуть курсор вначало
	if _, err := fs.file.Seek(0, 0); err != nil {
		return 0, err
	}

	scanner := bufio.NewScanner(fs.file)

	var res []recordFile

	for scanner.Scan() {
		var item recordFile

		if err := json.Unmarshal(scanner.Bytes(), &item); err != nil {
			continue
		}

		if item.IsDeleted {
			continue
		}

		res = append(res, item)
	}

	if err := scanner.Err(); err != nil {
		return 0, err
	}

	return len(res), nil
}
