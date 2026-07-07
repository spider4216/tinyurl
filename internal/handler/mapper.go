package handler

import (
	"github.com/spider4216/tinyurl/internal/models"
	"github.com/spider4216/tinyurl/internal/service"
)

func (h Handler) MapGenUrlsResp(urls []service.UrlsIds, baseUrl string) []models.ShortenBatchResp {
	res := make([]models.ShortenBatchResp, 0, len(urls))

	for _, url := range urls {
		full := baseUrl + "/" + url.Short

		v := models.ShortenBatchResp{
			CorrelationId: url.CorId,
			ShortUrl:      full,
		}

		res = append(res, v)
	}

	return res
}
