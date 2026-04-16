package service

import (
	"crypto/rand"
	"encoding/base64"

	"github.com/spider4216/tinyurl/internal/repository"
)

func New(repo *repository.Repository) Service {
	return Service{
		repo: repo,
	}
}

type Service struct {
	repo *repository.Repository
}

func (s Service) GenerateId() string {
	key := make([]byte, 6)
	rand.Read(key)

	return base64.URLEncoding.EncodeToString(key)
}

func (s Service) StoreData(id string, val string) error {
	return s.repo.Insert(id, val)
}

func (s Service) GetUrl(k string) (string, error) {
	return s.repo.Get(k)
}
