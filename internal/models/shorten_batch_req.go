package models

// ShortenBatchReq модель запроса на сокращение мнодества URL.
// generate:reset
type ShortenBatchReq struct {
	CorrelationId string `json:"correlation_id"`
	OriginalUrl   string `json:"original_url"`
}
