package modellverbindung

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"agentcontrolplane/app/internal/port/credentials"
)

func New(client Client, store credentials.Store) *Service {
	return &Service{client: client, store: store, status: Status{State: "idle"}}
}

func (s *Service) Start(ctx context.Context) (Status, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.pending != nil {
		return s.status, nil
	}

	connected, err := s.connectedBeforeStart(ctx)
	if err != nil || connected {
		return s.status, err
	}

	return s.begin(ctx)
}

func (s *Service) begin(ctx context.Context) (Status, error) {
	challenge, err := s.client.Start(ctx)
	if errors.Is(err, ErrDeviceCodeDisabled) {
		s.status = Status{State: "unavailable", Reason: "device_code_disabled"}
		return s.status, nil
	}

	if err != nil {
		return Status{}, err
	}

	return s.acceptChallenge(challenge), nil
}

func (s *Service) connectedBeforeStart(ctx context.Context) (bool, error) {
	_, _, err := s.access(ctx)
	if err == nil {
		s.status = Status{State: "connected"}
		return true, nil
	}

	if errors.Is(err, credentials.ErrNotFound) || errors.Is(err, ErrReauthenticationRequired) {
		return false, nil
	}

	return false, err
}

func (s *Service) acceptChallenge(challenge Challenge) Status {
	s.pending = &challenge
	s.nextPoll = time.Time{}
	s.status = Status{State: "pending", VerificationURL: challenge.VerificationURL, UserCode: challenge.UserCode}
	return s.status
}

func (s *Service) Status(ctx context.Context) (Status, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.pending != nil {
		return s.poll(ctx, false)
	}

	if s.status.State == "expired" || s.status.State == "denied" || s.status.State == "cancelled" {
		return s.status, nil
	}

	_, _, err := s.access(ctx)
	return s.statusFromAccess(err)
}

func (s *Service) Cancel(ctx context.Context) (Status, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.pending == nil {
		return s.status, nil
	}

	status, err := s.poll(ctx, true)
	if err != nil || status.State == "connected" {
		return status, err
	}

	s.pending = nil
	s.status = Status{State: "cancelled"}
	return s.status, nil
}

func (s *Service) poll(ctx context.Context, force bool) (Status, error) {
	if time.Now().After(s.pending.ExpiresAt) {
		return s.finish("expired", "code_expired"), nil
	}

	if !force && time.Now().Before(s.nextPoll) {
		return s.status, nil
	}

	result, err := s.client.Poll(ctx, *s.pending)
	if errors.Is(err, ErrExchangeFailed) {
		return s.finish("unavailable", "login_failed"), nil
	}

	if err != nil {
		return Status{}, err
	}

	return s.applyPoll(ctx, result)
}

func (s *Service) applyPoll(ctx context.Context, result PollResult) (Status, error) {
	if result.State == "pending" {
		s.nextPoll = time.Now().Add(time.Duration(s.pending.IntervalSeconds) * time.Second)
		return s.status, nil
	}

	if result.State == "expired" {
		return s.finish("expired", "code_expired"), nil
	}

	if result.State == "denied" {
		return s.finish("denied", "authorization_denied"), nil
	}

	return s.saveConnected(ctx, result)
}

func (s *Service) saveConnected(ctx context.Context, result PollResult) (Status, error) {
	if result.State != "connected" || !s.complete(result.Tokens) {
		return Status{}, ErrAccessUnavailable
	}

	if err := s.save(ctx, result.Tokens); err != nil {
		return Status{}, err
	}

	s.invalid = false
	return s.finish("connected", ""), nil
}

func (s *Service) finish(state, reason string) Status {
	s.pending = nil
	s.status = Status{State: state, Reason: reason}
	return s.status
}

func (s *Service) Access(ctx context.Context) (string, string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.access(ctx)
}

func (s *Service) access(ctx context.Context) (string, string, error) {
	if s.invalid {
		return "", "", ErrReauthenticationRequired
	}

	bundle, err := s.load(ctx)
	if err != nil {
		return "", "", err
	}

	if time.Now().Add(5 * time.Minute).Before(bundle.ExpiresAt) {
		return bundle.AccessToken, bundle.AccountID, nil
	}

	return s.refresh(ctx, bundle)
}

func (s *Service) refresh(ctx context.Context, bundle TokenBundle) (string, string, error) {
	next, err := s.client.Refresh(ctx, bundle)
	if errors.Is(err, ErrReauthenticationRequired) {
		s.invalid = true
		_ = s.store.Delete(ctx, CredentialKey)
		return "", "", ErrReauthenticationRequired
	}

	if err != nil || !s.complete(next) {
		return "", "", ErrAccessUnavailable
	}

	if err := s.save(ctx, next); err != nil {
		return "", "", err
	}

	return next.AccessToken, next.AccountID, nil
}

func (s *Service) load(ctx context.Context) (TokenBundle, error) {
	data, err := s.store.Load(ctx, CredentialKey)
	if err != nil {
		return TokenBundle{}, err
	}

	var bundle TokenBundle
	if json.Unmarshal(data, &bundle) != nil || !s.complete(bundle) {
		return TokenBundle{}, ErrAccessUnavailable
	}

	return bundle, nil
}

func (s *Service) save(ctx context.Context, bundle TokenBundle) error {
	data, err := json.Marshal(bundle)
	if err != nil {
		return ErrAccessUnavailable
	}

	return s.store.Save(ctx, CredentialKey, data)
}

func (s *Service) statusFromAccess(err error) (Status, error) {
	if err == nil {
		s.status = Status{State: "connected"}
		return s.status, nil
	}

	if errors.Is(err, credentials.ErrNotFound) && s.status.State != "connected" {
		return s.status, nil
	}

	if errors.Is(err, ErrReauthenticationRequired) || errors.Is(err, credentials.ErrNotFound) {
		s.status = Status{State: "reauthentication_required", Reason: "session_invalid"}
		return s.status, nil
	}

	return Status{}, err
}

func (s *Service) complete(bundle TokenBundle) bool {
	return bundle.IDToken != "" && bundle.AccessToken != "" && bundle.RefreshToken != "" && bundle.AccountID != "" && !bundle.ExpiresAt.IsZero()
}
