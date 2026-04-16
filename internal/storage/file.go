package storage

import (
	"encoding/json"
	"os"
)

type FileStorage struct {
	file *os.File
}

func NewFileStorage(filename string) (*FileStorage, error) {
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0666)

	if err != nil {
		return nil, err
	}

	return &FileStorage{
		file: file,
	}, nil
}

// Загружаем содержимое всего файла
func (fs *FileStorage) loadAll() ([]record, error) {
	if _, err := fs.file.Seek(0, 0); err != nil {
		return nil, err
	}

	var records []record

	stat, _ := fs.file.Stat()
	if stat.Size() == 0 {
		return records, nil
	}

	if err := json.NewDecoder(fs.file).Decode(&records); err != nil {
		return nil, err
	}

	return records, nil
}

// Перезаписываем файл
func (fs *FileStorage) writeAll(records []record) error {
	if err := fs.file.Truncate(0); err != nil {
		return err
	}

	if _, err := fs.file.Seek(0, 0); err != nil {
		return err
	}

	return json.NewEncoder(fs.file).Encode(records)
}

func (fs *FileStorage) Save(key string, data []byte) error {
	records, err := fs.loadAll()
	if err != nil {
		return err
	}

	record := record{}

	if err = json.Unmarshal(data, &record); err != nil {
		return err
	}

	record.Key = key

	records = append(records, record)

	return fs.writeAll(records)
}

func (fs *FileStorage) Load(key string) ([]byte, error) {
	records, err := fs.loadAll()
	if err != nil {
		return nil, err
	}

	for _, r := range records {
		if r.Key == key {
			b, err := json.Marshal(r)

			if err != nil {
				return nil, err
			}

			return b, nil
		}
	}

	return nil, os.ErrNotExist
}
