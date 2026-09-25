package modellfreigabe

import (
	"context"
	"errors"

	"agentcontrolplane/app/internal/app/openrouterverbindung"
	domainagent "agentcontrolplane/app/internal/domain/agent"
	domainmodell "agentcontrolplane/app/internal/domain/modellfreigabe"
	"agentcontrolplane/app/internal/port/modellschluessel"
)

func (s *Service) status(ctx context.Context, operatorID, organizationID, agentID string) (domainmodell.Status, error) {
	view, err := s.connection.Status(ctx)
	if err != nil {
		return domainmodell.Status{}, err
	}
	status := domainmodell.Status{Reference: openrouterverbindung.Reference, Connected: view.Connected, ConnectionStatus: view.Status}
	binding, err := s.binding(ctx)
	if errors.Is(err, ErrConnection) {
		status.Reason = domainmodell.ReasonConnection
		return status, nil
	}
	if err != nil {
		return domainmodell.Status{}, err
	}
	return s.grantedStatus(ctx, operatorID, organizationID, agentID, binding, status)
}

func (s *Service) statusForBinding(ctx context.Context, operatorID, organizationID, agentID string, binding modellschluessel.Binding) (domainmodell.Status, error) {
	view, err := s.connection.Status(ctx)
	if err != nil {
		return domainmodell.Status{}, err
	}
	status := domainmodell.Status{Reference: openrouterverbindung.Reference, Connected: view.Connected, ConnectionStatus: view.Status}
	return s.grantedStatus(ctx, operatorID, organizationID, agentID, binding, status)
}

func (s *Service) grantedStatus(ctx context.Context, operatorID, organizationID, agentID string, binding modellschluessel.Binding, status domainmodell.Status) (domainmodell.Status, error) {
	var err error
	status.OrganizationGranted, err = s.store.OrganizationGranted(ctx, operatorID, organizationID, binding)
	if err != nil {
		return domainmodell.Status{}, err
	}
	if agentID != "" {
		status.AgentGranted, err = s.store.AgentGranted(ctx, operatorID, organizationID, agentID, binding)
	}
	if err != nil {
		return domainmodell.Status{}, err
	}
	decision := status.Decide()
	status.Allowed, status.Reason = decision.Allowed, decision.Reason
	return status, nil
}

// Overview liest nur Eino-Agenten der zugeordneten Organisation.
func (s *Service) Overview(ctx context.Context, organizationID string) (Overview, error) {
	operatorID, err := s.organization(ctx, organizationID)
	if err != nil {
		return Overview{}, err
	}
	status, err := s.status(ctx, operatorID, organizationID, "")
	if err != nil {
		return Overview{}, err
	}
	organization, err := s.organizations.Get(ctx, organizationID, operatorID)
	if err != nil {
		return Overview{}, err
	}
	agents, err := s.agents.ListAgents(ctx, organizationID, operatorID)
	if err != nil {
		return Overview{}, err
	}
	overview, err := s.collect(ctx, operatorID, organizationID, status, agents)
	overview.OrganizationName = organization.Name
	return overview, err
}

func (s *Service) collect(ctx context.Context, operatorID, organizationID string, status domainmodell.Status, agents []domainagent.Agent) (Overview, error) {
	overview := Overview{OrganizationID: organizationID, Status: status, Agents: []AgentStatus{}}
	for _, agent := range agents {
		if agent.ExecutionKind != domainagent.Eino {
			continue
		}
		item, err := s.agentItem(ctx, operatorID, organizationID, agent)
		if err != nil {
			return Overview{}, err
		}
		overview.Agents = append(overview.Agents, item)
	}
	return overview, nil
}

func (s *Service) agentItem(ctx context.Context, operatorID, organizationID string, agent domainagent.Agent) (AgentStatus, error) {
	status, err := s.status(ctx, operatorID, organizationID, agent.ID)
	if err != nil {
		return AgentStatus{}, err
	}
	return AgentStatus{ID: agent.ID, Name: agent.Name, Status: status}, nil
}
