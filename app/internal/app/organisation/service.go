package organisation

import (
	"context"

	domainorganisation "agentcontrolplane/app/internal/domain/organisation"
	"agentcontrolplane/app/internal/domain/rechte"
	portorganisation "agentcontrolplane/app/internal/port/organisation"
	"github.com/google/uuid"
)

// NewService bindet Speicher und lokale Identitätsquelle an den Anwendungsfall.
func NewService(store portorganisation.Store, identity portorganisation.Identity) *Service {
	return &Service{store: store, identity: identity}
}

// Create legt eine Organisation samt restriktiver Regel für den Betreiber an.
func (s *Service) Create(ctx context.Context, name, description string) (domainorganisation.Organization, error) {
	actorID, err := s.operatorID()
	if err != nil {
		return domainorganisation.Organization{}, err
	}

	organization, err := domainorganisation.NewOrganization(uuid.NewString(), name, description)
	if err != nil {
		return domainorganisation.Organization{}, err
	}

	if err := s.store.Create(ctx, actorID, organization); err != nil {
		return domainorganisation.Organization{}, err
	}

	return organization, nil
}

// List zeigt nur dem lokalen Betreiber zugeordnete Organisationen.
func (s *Service) List(ctx context.Context) ([]domainorganisation.Organization, error) {
	actorID, err := s.operatorID()
	if err != nil {
		return nil, err
	}

	return s.store.List(ctx, actorID)
}

// Get gibt fremde und unbekannte Kennungen mit demselben Fehler zurück.
func (s *Service) Get(ctx context.Context, id string) (domainorganisation.Organization, error) {
	actorID, err := s.operatorID()
	if err != nil {
		return domainorganisation.Organization{}, err
	}

	return s.store.Get(ctx, id, actorID)
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
