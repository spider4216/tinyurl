package service

import (
	"crypto/rand"
	"encoding/base64"
)

type Store interface {
	Insert(k string, v string)
	Get(k string) string
}

func New(repo Store) Service {
	return Service{
		repo: repo,
	}
}

type Service struct {
	repo Store
}

func (s Service) GenerateId() string {
	key := make([]byte, 6)
	rand.Read(key)

	return base64.URLEncoding.EncodeToString(key)
}

func (s Service) StoreData(id string, val string) {
	s.repo.Insert(id, val)
}

func (s Service) GetUrl(k string) string {
	return s.repo.Get(k)
}
