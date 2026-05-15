package storage

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"slices"
	"strconv"
	"sync"

	"github.com/spider4216/tinyurl/internal/models"
	"go.uber.org/zap"
)

type FileStorage struct {
	file   *os.File
	mu     sync.RWMutex
	logger *zap.SugaredLogger
}

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
	IsDeleted string `json:"is_deleted"`
}

type FileIterator struct {
	scanner *bufio.Scanner
	file    *os.File
}

func (i *FileIterator) Next() bool {
	return i.scanner.Scan()
}

func (i *FileIterator) Row() ([]byte, error) {
	v := i.scanner.Text()

	return []byte(v), nil
}

func (i *FileIterator) Err() error {
	return nil
}

func (i *FileIterator) Close() error {
	return i.file.Close()
}

func (fs *FileStorage) SaveBatch(ctx context.Context, data [][]byte) error {
	for _, item := range data {
		if err := fs.Save(ctx, item); err != nil {
			return err
		}
	}

	return nil
}

func (fs *FileStorage) Save(ctx context.Context, data []byte) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	select {
	case <-ctx.Done():
		return fmt.Errorf("timeout in file save")
	default:
		data = append(data, '\n')

		if _, err := fs.file.Write(data); err != nil {
			return err
		}

		return nil
	}
}

func (fs *FileStorage) Load(ctx context.Context) (Iterator, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	// Вернуть курсор вначало
	if _, err := fs.file.Seek(0, 0); err != nil {
		return nil, err
	}

	return &FileIterator{
		scanner: bufio.NewScanner(fs.file),
		file:    fs.file,
	}, nil
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

// Поскольку pgx при работе с базой требует условия
// приходится идти на компромис и делать реализацию поиска
// в самом store слое, чтобы соблюсти единый интерфейс
// Здесь реализована внутренняя специфика
// связанная с конкретным хранилищем, в данном случае с файлом
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
		if err := os.Remove(tmpFile.Name()); err != nil {
			fs.logger.Warn("cannot remove tmp file", zap.Error(err))
		}
	}()

	// Произвожу поиск
	for scanner.Scan() {
		rec := map[string]string{}

		line := scanner.Bytes()

		if err := json.Unmarshal(line, &rec); err != nil {
			continue
		}

		sourceId, ok := rec["user_id"]

		if !ok {
			continue
		}

		short, ok := rec["short_url"]

		if !ok {
			continue
		}

		if sourceId == userId && slices.Contains(ids, short) {
			// Обновляем is_deleted
			rec["is_deleted"] = "true"
		}

		updLine, err := json.Marshal(rec)
		if err != nil {
			return err
		}

		if _, err := tmpFile.Write(updLine); err != nil {
			return err
		}

		if _, err := tmpFile.Write([]byte("\n")); err != nil {
			return err
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	originPath := fs.file.Name()

	// Закрываем старый файл
	if err := fs.file.Close(); err != nil {
		return err
	}

	// Заменяем файл
	if err := os.Rename(tmpFile.Name(), originPath); err != nil {
		return err
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

		b, err := strconv.ParseBool(item.IsDeleted)

		if err != nil {
			return nil, err
		}

		urls = append(urls, models.UrlItem{
			OriginarUrl: item.Origin,
			ShortUrl:    item.Short,
			IsDeleted:   b,
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

		b, err := strconv.ParseBool(item.IsDeleted)

		if err != nil {
			return nil, err
		}

		return &models.UrlItem{
			OriginarUrl: item.Origin,
			ShortUrl:    item.Short,
			IsDeleted:   b,
			UserId:      item.UserId,
		}, nil
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return nil, errors.New("cannot find url")

}
