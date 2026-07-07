package models

// ShortenBatchResp модель ответа на запрос по сокращению мнодества URL.
type ShortenBatchResp struct {
	CorrelationId string `json:"correlation_id"`
	ShortUrl      string `json:"short_url"`
}
