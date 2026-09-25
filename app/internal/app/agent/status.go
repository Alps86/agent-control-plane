package agent

import (
	"context"

	domainagent "agentcontrolplane/app/internal/domain/agent"
	portagent "agentcontrolplane/app/internal/port/agent"
)

// SetStatus ändert den Agentenstatus nur für die zugeordnete Betreiberorganisation.
func (s *Service) SetStatus(ctx context.Context, organizationID, agentID string, status domainagent.Status) (Profile, error) {
	if !status.Valid() {
		return Profile{}, ErrInvalidStatus
	}

	operatorID, err := s.authorize(ctx, organizationID)
	if err != nil {
		return Profile{}, err
	}

	store, ok := s.store.(portagent.StatusStore)
	if !ok {
		return Profile{}, ErrAccessDenied
	}

	if err := store.UpdateAgentStatus(ctx, organizationID, agentID, operatorID, status); err != nil {
		return Profile{}, err
	}

	return s.Get(ctx, organizationID, agentID)
}
