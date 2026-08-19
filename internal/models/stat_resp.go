package models

// StatResp структура ответа для эндопинта статистики сервиса.
type StatResp struct {
	UrlsCount  int `json:"urls"`
	UsersCount int `json:"users"`
}
