package models

// ShortenBatchReq модель запроса на сокращение мнодества URL.
type ShortenBatchReq struct {
	CorrelationId string `json:"correlation_id"`
	OriginalUrl   string `json:"original_url"`
}
