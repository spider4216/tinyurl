package audit

import (
	"encoding/json"
	"os"
	"time"

	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
)

const (
	AuditFileObserverID   string = "audit_file_onserver"
	AuditServerObserverID string = "audit_server_observer"
)

type Publisher interface {
	Register(ob Observer)
	Notify(data Body) error
}

type Observer interface {
	GetID() string
	Update(data Body) error
}

type AuditEvent struct {
	Observers map[string]Observer
}

func NewAuditEvent() *AuditEvent {
	return &AuditEvent{
		Observers: make(map[string]Observer),
	}
}

type Body struct {
	TS     time.Time
	Action string
	UserID string
	URL    string
}

type AuditFileObserver struct {
	ID     string
	Path   string
	Logger *zap.SugaredLogger
}

func NewAuditFileObserver(path string, logger *zap.SugaredLogger) *AuditFileObserver {
	return &AuditFileObserver{
		ID:     AuditFileObserverID,
		Path:   path,
		Logger: logger,
	}
}

func (afo *AuditFileObserver) GetID() string {
	return afo.ID
}

func (afo *AuditFileObserver) Update(data Body) error {
	raw, err := json.Marshal(data)

	if err != nil {
		return err
	}

	file, err := os.OpenFile(afo.Path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	if err != nil {
		return err
	}

	defer func() {
		if err := file.Close(); err != nil {
			afo.Logger.Warn("Cannot close file correctly")
		}
	}()

	if _, err = file.Write(raw); err != nil {
		return err
	}

	return nil
}

type AuditServerObserver struct {
	ID     string
	URL    string
	Logger *zap.SugaredLogger
	Cli    *resty.Client
}

func NewAuditServerObserver(host string, url string, logger *zap.SugaredLogger) *AuditServerObserver {
	cli := resty.New().
		SetBaseURL(host).
		SetHeader("Content-Type", "application/json")

	return &AuditServerObserver{
		ID:     AuditServerObserverID,
		URL:    url,
		Logger: logger,
		Cli:    cli,
	}
}

func (aso *AuditServerObserver) GetID() string {
	return aso.ID
}

func (aso *AuditServerObserver) Update(data Body) error {
	_, err := aso.Cli.R().SetBody(data).Post(aso.URL)

	return err
}

func (ae *AuditEvent) Register(o Observer) {
	ae.Observers[o.GetID()] = o
}

func (ae *AuditEvent) Notify(data Body) error {
	for _, o := range ae.Observers {
		if err := o.Update(data); err != nil {
			return err
		}
	}

	return nil
}
