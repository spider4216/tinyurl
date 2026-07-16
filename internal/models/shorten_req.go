package models

// ShortenReq модель на запрос сокращения URL.
// generate:reset
type ShortenReq struct {
	Url string `json:"url"`
}
