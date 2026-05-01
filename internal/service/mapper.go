package service

import "github.com/spider4216/tinyurl/internal/models"

type UrlsIds struct {
	Short  string
	Origin string
	CorId  string
}

func (s Service) MapForMapUrlIds(m []models.ShortenBatchReq) []UrlsIds {
	var urls []UrlsIds

	for _, item := range m {
		u := UrlsIds{
			Short:  s.GenerateId(),
			Origin: item.OriginalUrl,
			CorId:  item.CorrelationId,
		}

		urls = append(urls, u)
	}

	return urls
}
