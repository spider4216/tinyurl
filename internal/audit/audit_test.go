package audit

import (
	"encoding/json"
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
