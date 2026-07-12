package models

import "fmt"

// UrlItem единая модель на которую мапятся данные из разных Store.
type UrlItem struct {
	ShortUrl    string `json:"short_url"`
	OriginarUrl string `json:"original_url"`
	IsDeleted   bool   `json:"deleted_at"`
	UserId      string `json:"user_id"`
}

func (v *UrlItem) Reset() {
	fmt.Println("reset")
}
