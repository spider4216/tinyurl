package repository

import "github.com/spider4216/tinyurl/internal/models"

func (r *Repository) MapInserData(k string, v string, userId string) models.InsertData {
	return models.InsertData{
		Key:    k,
		Value:  v,
		UserId: userId,
	}
}
