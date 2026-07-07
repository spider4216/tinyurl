package models

// ShortenReq модель на запрос сокращения URL.
type ShortenReq struct {
	Url string `json:"url"`
}
