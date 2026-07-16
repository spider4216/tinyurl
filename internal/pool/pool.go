package pool

import (
	"sync"

	"github.com/spider4216/tinyurl/internal/models"
)

// ReqPools Пул запросов которе будет использовать основной Handler.
// Можно масштабировать на новые структуры.
type ReqPools struct {
	ShortenReq *Pool[*models.ShortenReq]
}

// Resetter интерфейс для пула, т.е. пулл будет содержать и
// извлекать структуру имплементирующую данный интерфейс.
type Resetter interface {
	Reset()
}

// Pool Обобщенная структура пулла, т.е. в sync.Pool
// Будет ложиться и извлекаться обобщенный тип реализующий интерфейс Resetter.
type Pool[T Resetter] struct {
	pool sync.Pool
}

// New конструктор моего пулла. Также обощенный. Создает структуру Pool
// в которую ложит sync.Pool в котором непосредственно будет оежать
// структура реализующая интерфейс Resetter.
func New[T Resetter](newFn func() T) *Pool[T] {
	return &Pool[T]{
		pool: sync.Pool{
			New: func() any {
				return newFn()
			},
		},
	}
}

// Get извлечение структуры из пулла.
func (p *Pool[T]) Get() T {
	return p.pool.Get().(T)
}

// Put возвращение структуры в pool, но перед возвратом
// осуществляется сброс этой структуры.
func (p *Pool[T]) Put(v T) {
	v.Reset()
	p.pool.Put(v)
}
