package storage

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"sync"
)

type FileStorage struct {
	file *os.File
	mu   sync.RWMutex
}

func NewFileStorage(filename string) (*FileStorage, error) {
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)

	if err != nil {
		return nil, err
	}

	return &FileStorage{
		file: file,
	}, nil
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
		fs.Save(ctx, item)
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
