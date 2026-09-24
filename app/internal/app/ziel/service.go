package ziel

import (
	"context"

	domainziel "agentcontrolplane/app/internal/domain/ziel"
	portziel "agentcontrolplane/app/internal/port/ziel"
	"github.com/google/uuid"
)

// NewService bindet Zielstore und Organisationszugriff an den Anwendungsfall.
func NewService(store portziel.GoalStore, organizations portziel.OrganizationReader) *Service {
	return &Service{store: store, organizations: organizations}
}

// Create legt ein neues Stammziel in einer zugänglichen Organisation an.
func (s *Service) Create(ctx context.Context, organizationID, name string) (domainziel.Goal, error) {
	if err := s.checkOrganization(ctx, organizationID); err != nil {
		return domainziel.Goal{}, err
	}

	goal, err := (domainziel.Goal{}).NewRootGoal(uuid.NewString(), organizationID, name)
	if err != nil {
		return domainziel.Goal{}, err
	}

	if err := s.store.CreateGoal(ctx, goal); err != nil {
		return domainziel.Goal{}, err
	}

	return goal, nil
}

// List zeigt nur Ziele einer zugänglichen Organisation.
func (s *Service) List(ctx context.Context, organizationID string) ([]domainziel.Goal, error) {
	if err := s.checkOrganization(ctx, organizationID); err != nil {
		return nil, err
	}

	return s.store.ListGoals(ctx, organizationID)
}

func (s *Service) checkOrganization(ctx context.Context, organizationID string) error {
	if s == nil || s.store == nil || s.organizations == nil {
		return ErrAccessDenied
	}

	_, err := s.organizations.Get(ctx, organizationID)
	return err
}
