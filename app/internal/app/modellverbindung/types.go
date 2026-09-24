package modellverbindung

import (
	"context"
	"errors"
	"sync"
	"time"

	"agentcontrolplane/app/internal/port/credentials"
)

var ErrDeviceCodeDisabled = errors.New("device code disabled")
var ErrReauthenticationRequired = errors.New("reauthentication required")
var ErrAccessUnavailable = errors.New("codex access unavailable")
var ErrExchangeFailed = errors.New("device code exchange failed")

const CredentialKey = "codex-chatgpt"

type Challenge struct {
	LoginID         string
	VerificationURL string
	UserCode        string
	DeviceAuthID    string
	IntervalSeconds int
	ExpiresAt       time.Time
}

type Account struct {
	Type string
}

type Completion struct {
	LoginID string
	Success bool
	Error   string
}

type TokenBundle struct {
	IDToken      string    `json:"idToken"`
	AccessToken  string    `json:"accessToken"`
	RefreshToken string    `json:"refreshToken"`
	AccountID    string    `json:"accountId"`
	ExpiresAt    time.Time `json:"expiresAt"`
}

type PollResult struct {
	State  string
	Tokens TokenBundle
}

type Client interface {
	Start(context.Context) (Challenge, error)
	Poll(context.Context, Challenge) (PollResult, error)
	Refresh(context.Context, TokenBundle) (TokenBundle, error)
}

type Status struct {
	State           string `json:"state"`
	Reason          string `json:"reason,omitempty"`
	VerificationURL string `json:"verificationUrl,omitempty"`
	UserCode        string `json:"userCode,omitempty"`
}

type DeviceFlow interface {
	Start(context.Context) (Status, error)
	Status(context.Context) (Status, error)
	Cancel(context.Context) (Status, error)
}

type Service struct {
	client   Client
	store    credentials.Store
	mu       sync.Mutex
	status   Status
	pending  *Challenge
	nextPoll time.Time
	invalid  bool
}

var _ credentials.AccessResolver = (*Service)(nil)
