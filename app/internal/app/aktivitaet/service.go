package aktivitaet

import (
	"context"
	"time"

	apporganisation "agentcontrolplane/app/internal/app/organisation"
	domainaktivitaet "agentcontrolplane/app/internal/domain/aktivitaet"
	domainaufgabe "agentcontrolplane/app/internal/domain/aufgabe"
	"agentcontrolplane/app/internal/domain/rechte"
	portaktivitaet "agentcontrolplane/app/internal/port/aktivitaet"
	portorganisation "agentcontrolplane/app/internal/port/organisation"
)

// NewService bindet Eventstore und serverseitige Betreiberidentität.
func NewService(store portaktivitaet.Store, organizations *apporganisation.Service, identity portorganisation.Identity) *Service {
	return &Service{store: store, organizations: organizations, identity: identity}
}

// List gibt nur Ereignisse einer zugänglichen Organisation zurück.
func (s *Service) List(ctx context.Context, organizationID string) ([]domainaktivitaet.Event, error) {
	if s == nil || s.store == nil || s.organizations == nil {
		return nil, ErrAccessDenied
	}

	if _, err := s.organizations.Get(ctx, organizationID); err != nil {
		return nil, err
	}

	events, err := s.store.ListEvents(ctx, organizationID)
	if events == nil {
		events = []domainaktivitaet.Event{}
	}

	return events, err
}

// RecordTaskCreation zeichnet Anlage und anfängliche Zuweisung im übergebenen Transaktionskontext auf.
func (s *Service) RecordTaskCreation(ctx context.Context, task domainaufgabe.Task, assigneeName, source string) error {
	actor, err := s.operatorID()
	if err != nil {
		return err
	}

	if source != domainaktivitaet.SourceAPI && source != domainaktivitaet.SourceBrowser {
		return ErrInvalidSource
	}

	at := time.Now().UTC()
	created := domainaktivitaet.NewTaskEvent(task, domainaktivitaet.KindCreated, actor, source, "", at)
	assigned := domainaktivitaet.NewTaskEvent(task, domainaktivitaet.KindAssigned, actor, source, assigneeName, at)
	if err := s.store.Record(ctx, created); err != nil {
		return err
	}

	return s.store.Record(ctx, assigned)
}

func (s *Service) operatorID() (string, error) {
	if s == nil || s.store == nil || s.identity == nil {
		return "", ErrAccessDenied
	}

	actors := s.identity.Actors()
	if len(actors) != 1 || !actors[0].Valid() || actors[0].Kind != rechte.Operator {
		return "", ErrAccessDenied
	}

	return actors[0].ID, nil
}
