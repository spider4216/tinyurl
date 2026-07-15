package pool

import (
	"testing"

	"github.com/spider4216/tinyurl/internal/models"
	"github.com/stretchr/testify/assert"
)

// TestPool Тест обощенного пулла
func TestPool(t *testing.T) {
	// Создаем pool
	p := New(func() *models.ShortenReq {
		return &models.ShortenReq{}
	})

	// Извлекаем из пулла
	req := p.Get()

	// Проверяем, что извлеченное равняется тому что положили
	assert.IsType(t, models.ShortenReq{}, *req)

	url := "http://example.com"

	// Меняем содердимое структуры
	req.Url = url

	// Проверяем изменения
	assert.Equal(t, url, req.Url)

	// Кладем структуру обратно в pool
	p.Put(req)

	// Снова извлекаем структуру
	diff := p.Get()

	// Проверяем, что извлеченное равняется тому что положили
	assert.IsType(t, models.ShortenReq{}, *diff)

	// Убеждаемся что сброс произошел
	assert.Equal(t, "", diff.Url)
}
