package storage

import (
	"io"
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

func (fs *FileStorage) Save(data []byte) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	data = append(data, '\n')

	if _, err := fs.file.Write(data); err != nil {
		return err
	}

	return nil
}

func (fs *FileStorage) Load() (io.Reader, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	// Вернуть курсор вначало
	if _, err := fs.file.Seek(0, 0); err != nil {
		return nil, err
	}

	return fs.file, nil
}

func (fs *FileStorage) Ping() error {
	return nil
}
