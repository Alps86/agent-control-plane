package openrouterverbindung

import (
	"context"
	"errors"
	"sync"

	"agentcontrolplane/app/internal/port/credentials"
)

const Reference = "openrouter-central"

const (
	CheckReady       CheckResult = "ready"
	CheckInvalid     CheckResult = "invalid"
	CheckUnavailable CheckResult = "unavailable"
)

var ErrMissingKey = errors.New("Schlüssel fehlt")
var ErrStorage = errors.New("Verbindung derzeit nicht verfügbar")

type CheckResult string

type Probe interface {
	Check(context.Context, string) CheckResult
}

type Service struct {
	mu    sync.Mutex
	store credentials.Store
	probe Probe
}

type View struct {
	Connected    bool
	Reference    string
	Status       string
	StatusDetail string
}

type record struct {
	Key    string `json:"key"`
	Status string `json:"status"`
}
