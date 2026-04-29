package handler

import (
	"fmt"

	"github.com/spider4216/tinyurl/internal/models"
	"github.com/spider4216/tinyurl/internal/service"
)

func (h Handler) MapGenUrlsResp(urls []service.UrlsIds, baseUrl string) []models.ShortenBatchResp {
	res := []models.ShortenBatchResp{}

	for _, url := range urls {
		full := fmt.Sprintf("%s/%s", baseUrl, url.Short)

		v := models.ShortenBatchResp{
			CorrelationId: url.CorId,
			ShortUrl:      full,
		}

		res = append(res, v)
	}

	return res
}
