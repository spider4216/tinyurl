package service

import "github.com/spider4216/tinyurl/internal/models"

type UrlsIds struct {
	Short  string
	Origin string
	CorId  string
	UserId string
}

func (s Service) MapForMapUrlIds(m []models.ShortenBatchReq, userId string) []UrlsIds {
	urls := make([]UrlsIds, 0, len(m))

	for _, item := range m {
		u := UrlsIds{
			Short:  s.GenerateId(),
			Origin: item.OriginalUrl,
			CorId:  item.CorrelationId,
			UserId: userId,
		}

		urls = append(urls, u)
	}

	return urls
}
