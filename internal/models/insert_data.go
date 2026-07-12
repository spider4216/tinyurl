package models

// InsertData модель для вставки сокращенного URL.
// generate:reset
type InsertData struct {
	Key      string
	Value    string
	UserId   string
	MyCustom *string
	Gen      int
}
