package storage

import (
	"encoding/json"
	"log"
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

func (fs *FileStorage) Save(key string, data []byte) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	enc := json.NewEncoder(fs.file)

	record := record{}

	if err := json.Unmarshal(data, &record); err != nil {
		return err
	}

	record.Key = key

	return enc.Encode(record)
}

func (fs *FileStorage) Load(key string) ([]byte, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	// Вернуть курсор вначало
	fs.file.Seek(0, 0)

	dec := json.NewDecoder(fs.file)

	for dec.More() {
		item := record{}

		if err := dec.Decode(&item); err != nil {
			continue
		}

		log.Println(item.Key)

		if item.Key == key {
			return json.Marshal(item)
		}
	}

	return nil, os.ErrNotExist
}
