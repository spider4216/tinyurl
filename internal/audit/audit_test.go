package audit

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/spider4216/tinyurl/internal/logger"
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
	logger, err := logger.InitZap("debug")
	require.NoError(t, err)

	event := NewAuditEvent(logger)

	ob := &auditSliceObserver{}

	event.Register(ob)

	data := Body{
		TS:     time.Now().Unix(),
		Action: ShortenAction,
		UserID: "test1",
		URL:    "http://test.loc",
	}

	event.Notify(data)

	time.Sleep(500 * time.Microsecond)

	assert.Len(t, ob.Data, 1)
	raw := ob.Data[0]

	model := Body{}

	err = json.Unmarshal(raw, &model)
	require.NoError(t, err)

	assert.Equal(t, data.Action, model.Action)
}

func ExampleAuditEvent_Notify() {
	logger, _ := logger.InitZap("debug")

	// Создаем событие
	event := NewAuditEvent(logger)

	// Готовим Observer в котором фигурирует логика оповещения
	// Можно создать свой Observer и реализовать свою логику, например
	// Вести аудит в БД. Главное чтобы пользовательский Observer
	// реализовал интерфейс Observer
	ob := &auditSliceObserver{}

	// Регистрируем Observer
	event.Register(ob)

	// Готовим тело аудита
	data := Body{
		TS:     time.Now().Unix(),
		Action: ShortenAction,
		UserID: "test1",
		URL:    "http://test.loc",
	}

	event.Notify(data)
}
