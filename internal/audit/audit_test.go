package audit

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type auditSliceObserver struct {
	Data [][]byte
}

func (aso *auditSliceObserver) GetID() string {
	return "test_audit_slice_observer"
}

func (aso *auditSliceObserver) Update(data Body) error {
	raw, err := json.Marshal(data)
	if err != nil {
		return err
	}

	aso.Data = append(aso.Data, raw)

	return nil
}

func TestAudit(t *testing.T) {
	event := NewAuditEvent()

	ob := &auditSliceObserver{}

	event.Register(ob)

	data := Body{
		TS:     time.Now(),
		Action: ShortenAction,
		UserID: "test1",
		URL:    "http://test.loc",
	}

	err := event.Notify(data)
	require.NoError(t, err)

	assert.Len(t, ob.Data, 1)
	raw := ob.Data[0]

	model := Body{}

	err = json.Unmarshal(raw, &model)
	require.NoError(t, err)

	assert.Equal(t, data.Action, model.Action)
}

func ExampleAuditEvent_Notify() {
	// Создаем событие
	event := NewAuditEvent()

	// Готовим Observer в котором фигурирует логика оповещения
	// Можно создать свой Observer и реализовать свою логику, например
	// Вести аудит в БД. Главное чтобы пользовательский Observer
	// реализовал интерфейс Observer
	ob := &auditSliceObserver{}

	// Регистрируем Observer
	event.Register(ob)

	// Готовим тело аудита
	data := Body{
		TS:     time.Now(),
		Action: ShortenAction,
		UserID: "test1",
		URL:    "http://test.loc",
	}

	if err := event.Notify(data); err != nil {
		// Обработать ошибку, например залогировать
	}

	fmt.Println("Оповещение прошло успешно")
}
