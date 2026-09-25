package berichtsweg

import (
	"context"
	"errors"

	domainberichtsweg "agentcontrolplane/app/internal/domain/berichtsweg"
	"agentcontrolplane/app/internal/domain/rechte"
	portagent "agentcontrolplane/app/internal/port/agent"
	portberichtsweg "agentcontrolplane/app/internal/port/berichtsweg"
	portorganisation "agentcontrolplane/app/internal/port/organisation"
)

// NewService bindet Identität und Speicher an den organisationsgebundenen Service.
func NewService(store portberichtsweg.Store, agents portagent.Store, organizations portorganisation.Store, identity portorganisation.Identity) *Service {
	return &Service{store: store, agents: agents, organizations: organizations, identity: identity}
}

// Chart liest den sichtbaren Berichtsweg innerhalb einer zugänglichen Organisation.
func (s *Service) Chart(ctx context.Context, organizationID string) (domainberichtsweg.Chart, error) {
	operatorID, err := s.authorize(ctx, organizationID)
	if err != nil {
		return domainberichtsweg.Chart{}, err
	}

	lines, err := s.store.ListLines(ctx, organizationID, operatorID)
	if err != nil {
		return domainberichtsweg.Chart{}, err
	}

	return domainberichtsweg.Chart{OrganizationID: organizationID, Lines: lines}, nil
}

// Assign ersetzt nur den unmittelbaren Vorgesetzten und gibt den aktuellen Plan zurück.
func (s *Service) Assign(ctx context.Context, organizationID, agentID, parentID string) (domainberichtsweg.Chart, error) {
	operatorID, err := s.authorize(ctx, organizationID)
	if err != nil {
		return domainberichtsweg.Chart{}, err
	}

	if err := s.store.AssignParent(ctx, organizationID, operatorID, agentID, parentID); err != nil {
		return domainberichtsweg.Chart{}, err
	}

	return s.Chart(ctx, organizationID)
}

func (s *Service) authorize(ctx context.Context, organizationID string) (string, error) {
	if s == nil || s.store == nil || s.agents == nil || s.organizations == nil || s.identity == nil {
		return "", ErrAccessDenied
	}

	actors := s.identity.Actors()
	if len(actors) != 1 || !actors[0].Valid() || actors[0].Kind != rechte.Operator {
		return "", ErrAccessDenied
	}

	_, err := s.organizations.Get(ctx, organizationID, actors[0].ID)
	if errors.Is(err, portorganisation.ErrNotFound) {
		return "", ErrNotFound
	}

	return actors[0].ID, err
}
