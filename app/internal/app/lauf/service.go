package lauf

import (
	"context"

	apprechte "agentcontrolplane/app/internal/app/rechte"
	domainlauf "agentcontrolplane/app/internal/domain/lauf"
	portlauf "agentcontrolplane/app/internal/port/lauf"
)

// NewService bindet den Laufstore an die vorhandene zentrale Rechtegrenze.
func NewService(store portlauf.Store, reader *apprechte.Reader) *Service {
	return &Service{store: store, reader: reader}
}

// Reserve legt höchstens eine aktive Reservierung je Aufgabe an.
func (s *Service) Reserve(ctx context.Context, taskID string) (domainlauf.Run, bool, error) {
	organizationID, err := s.authorize(taskID)
	if err != nil {
		return domainlauf.Run{}, false, err
	}

	return s.store.Reserve(ctx, organizationID, taskID)
}

// Lookup liest einen Lauf erst nach der zentralen Rechteentscheidung.
func (s *Service) Lookup(ctx context.Context, taskID, runID string) (domainlauf.Run, error) {
	organizationID, err := s.authorize(taskID)
	if err != nil {
		return domainlauf.Run{}, err
	}

	run, found, err := s.store.Lookup(ctx, organizationID, taskID, runID)
	if err != nil {
		return domainlauf.Run{}, err
	}

	if !found {
		return domainlauf.Run{}, ErrUnavailable
	}

	return run, nil
}

// List zeigt nur Läufe einer lesbaren Aufgabe.
func (s *Service) List(ctx context.Context, taskID string) ([]domainlauf.Run, error) {
	organizationID, err := s.authorize(taskID)
	if err != nil {
		return nil, err
	}

	return s.store.List(ctx, organizationID, taskID)
}

// FailPreparation persistiert einen Vorbereitungsfehler vor Adapterstart.
func (s *Service) FailPreparation(ctx context.Context, taskID, runID, reason string) (domainlauf.Run, error) {
	run, err := s.Lookup(ctx, taskID, runID)
	if err != nil {
		return domainlauf.Run{}, err
	}

	return s.store.FailPreparation(ctx, run.OrganizationID, taskID, runID, reason)
}

// Complete weist einen unbelegten Erfolg ausdrücklich ab.
func (s *Service) Complete(ctx context.Context, taskID, runID string) (domainlauf.Run, error) {
	if _, err := s.Lookup(ctx, taskID, runID); err != nil {
		return domainlauf.Run{}, err
	}

	return domainlauf.Run{}, ErrAdapterEvidenceRequired
}

func (s *Service) authorize(taskID string) (string, error) {
	if s == nil || s.store == nil || s.reader == nil {
		return "", ErrUnavailable
	}

	resource, outcome := s.reader.Read(taskID)
	if outcome != apprechte.Allowed {
		return "", ErrUnavailable
	}

	if resource.OrganizationID == "" {
		return "", ErrUnavailable
	}

	return resource.OrganizationID, nil
}
