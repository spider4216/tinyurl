package audit

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
)

type AuditAction string

const (
	AuditFileObserverID   string      = "audit_file_onserver"
	AuditServerObserverID string      = "audit_server_observer"
	ShortenAction         AuditAction = "shorten"
	FollowAction          AuditAction = "follow"
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
	TS     int64       `json:"ts"`
	Action AuditAction `json:"action"`
	UserID string      `json:"user_id"`
	URL    string      `json:"url"`
}

type AuditFileObserver struct {
	ID     string
	Logger *zap.SugaredLogger
	File   *os.File
	mu     sync.RWMutex
}

func NewAuditFileObserver(path string, logger *zap.SugaredLogger) (*AuditFileObserver, error) {
	logger.Debug("Audit file observer was created")

	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, err
	}

	return &AuditFileObserver{
		ID:     AuditFileObserverID,
		Logger: logger,
		File:   file,
	}, nil
}

func (afo *AuditFileObserver) GetID() string {
	return afo.ID
}

func (afo *AuditFileObserver) Update(data Body) error {
	afo.mu.Lock()
	defer afo.mu.Unlock()

	afo.Logger.Debug("Update in audit file observer")

	raw, err := json.Marshal(data)
	if err != nil {
		return err
	}

	if _, err = afo.File.Write(raw); err != nil {
		afo.Logger.Error("cannot write in audit file", zap.Error(err))
		return err
	}

	afo.Logger.Debug("File audit append ", string(raw))

	return nil
}

type AuditServerObserver struct {
	ID     string
	URL    string
	Logger *zap.SugaredLogger
	Cli    *resty.Client
}

func NewAuditServerObserver(host string, url string, logger *zap.SugaredLogger) *AuditServerObserver {
	logger.Debug("Audit server observer was created")

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
	aso.Logger.Debug("Update in audit server observer")

	resp, err := aso.Cli.R().SetBody(data).Post(aso.URL)

	if err != nil {
		return err
	}

	if resp.IsError() {
		return fmt.Errorf("Audit server error. Code: %v", resp.StatusCode())
	}

	return nil
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
