package modellfreigabe

import (
	"context"
	"errors"

	"agentcontrolplane/app/internal/app/openrouterverbindung"
	domainagent "agentcontrolplane/app/internal/domain/agent"
	domainmodell "agentcontrolplane/app/internal/domain/modellfreigabe"
	"agentcontrolplane/app/internal/domain/rechte"
	portagent "agentcontrolplane/app/internal/port/agent"
	portmodell "agentcontrolplane/app/internal/port/modellfreigabe"
	"agentcontrolplane/app/internal/port/modellschluessel"
	portorganisation "agentcontrolplane/app/internal/port/organisation"
)

func NewService(store portmodell.Store, organizations portorganisation.Store, agents portagent.Store, identity portorganisation.Identity, connection portmodell.ConnectionStatus) *Service {
	return &Service{store: store, organizations: organizations, agents: agents, identity: identity, connection: connection}
}

func (s *Service) GrantOrganization(ctx context.Context, organizationID string) error {
	operatorID, err := s.organization(ctx, organizationID)
	if err != nil {
		return err
	}
	binding, err := s.binding(ctx)
	if err != nil {
		return err
	}
	return s.store.GrantOrganization(ctx, operatorID, organizationID, binding)
}

func (s *Service) RevokeOrganization(ctx context.Context, organizationID string) error {
	operatorID, err := s.organization(ctx, organizationID)
	if err != nil {
		return err
	}
	return s.store.RevokeOrganization(ctx, operatorID, organizationID)
}

func (s *Service) GrantAgent(ctx context.Context, organizationID, agentID string) error {
	operatorID, err := s.einoAgent(ctx, organizationID, agentID)
	if err != nil {
		return err
	}
	binding, err := s.binding(ctx)
	if err != nil {
		return err
	}
	return s.store.GrantAgent(ctx, operatorID, organizationID, agentID, binding)
}

func (s *Service) RevokeAgent(ctx context.Context, organizationID, agentID string) error {
	operatorID, err := s.einoAgent(ctx, organizationID, agentID)
	if err != nil {
		return err
	}
	return s.store.RevokeAgent(ctx, operatorID, organizationID, agentID)
}

func (s *Service) OrganizationStatus(ctx context.Context, organizationID string) (domainmodell.Status, error) {
	operatorID, err := s.organization(ctx, organizationID)
	if err != nil {
		return domainmodell.Status{}, err
	}
	return s.status(ctx, operatorID, organizationID, "")
}

func (s *Service) AgentStatus(ctx context.Context, organizationID, agentID string) (domainmodell.Status, error) {
	operatorID, err := s.einoAgent(ctx, organizationID, agentID)
	if err != nil {
		return domainmodell.Status{}, err
	}
	return s.status(ctx, operatorID, organizationID, agentID)
}

func (s *Service) Authorize(ctx context.Context, organizationID, agentID, reference string) (domainmodell.Decision, error) {
	decision, _, err := s.authorizeBinding(ctx, organizationID, agentID, reference)
	return decision, err
}

func (s *Service) authorizeBinding(ctx context.Context, organizationID, agentID, reference string) (domainmodell.Decision, modellschluessel.Binding, error) {
	if reference != openrouterverbindung.Reference {
		return domainmodell.Decision{Reason: domainmodell.ReasonReference}, modellschluessel.Binding{}, nil
	}
	operatorID, err := s.einoAgent(ctx, organizationID, agentID)
	if err != nil {
		return domainmodell.Decision{}, modellschluessel.Binding{}, err
	}
	binding, err := s.binding(ctx)
	if errors.Is(err, ErrConnection) {
		return domainmodell.Decision{Reason: domainmodell.ReasonConnection}, modellschluessel.Binding{}, nil
	}
	if err != nil {
		return domainmodell.Decision{}, modellschluessel.Binding{}, err
	}
	decision, err := s.currentDecision(ctx, operatorID, organizationID, agentID, binding)
	return decision, binding, err
}

func (s *Service) currentDecision(ctx context.Context, operatorID, organizationID, agentID string, binding modellschluessel.Binding) (domainmodell.Decision, error) {
	status, err := s.statusForBinding(ctx, operatorID, organizationID, agentID, binding)
	if err != nil {
		return domainmodell.Decision{}, err
	}
	current, err := s.binding(ctx)
	if errors.Is(err, ErrConnection) {
		return domainmodell.Decision{Reason: domainmodell.ReasonConnection}, nil
	}
	if err != nil {
		return domainmodell.Decision{}, err
	}
	if current != binding {
		return domainmodell.Decision{Reason: domainmodell.ReasonConnection}, nil
	}
	return status.Decide(), nil
}

// Invoke prüft unmittelbar vor dem noch nicht ausgeführten Provideraufruf erneut.
func (s *Service) Invoke(ctx context.Context, organizationID, agentID, reference string, caller portmodell.Caller) error {
	decision, binding, err := s.authorizeBinding(ctx, organizationID, agentID, reference)
	if err != nil {
		return err
	}
	if !decision.Allowed {
		return decision
	}
	if caller == nil {
		return ErrAccessDenied
	}
	return caller.Call(ctx, organizationID, agentID, &access{
		service: s, organizationID: organizationID, agentID: agentID, binding: binding})
}

// Resolve prüft dieselbe Bindung und beide Freigaben erneut vor dem Transport.
func (a *access) Resolve(ctx context.Context) (string, error) {
	if a == nil || a.service == nil {
		return "", ErrAccessDenied
	}
	operatorID, err := a.service.einoAgent(ctx, a.organizationID, a.agentID)
	if err != nil {
		return "", err
	}
	decision, err := a.service.currentDecision(ctx, operatorID, a.organizationID, a.agentID, a.binding)
	if err != nil {
		return "", err
	}
	if !decision.Allowed {
		return "", decision
	}
	return a.service.connection.Resolve(ctx, a.binding)
}

func (s *Service) operatorID() (string, error) {
	if s == nil || s.store == nil || s.organizations == nil || s.agents == nil || s.identity == nil || s.connection == nil {
		return "", ErrAccessDenied
	}
	actors := s.identity.Actors()
	if len(actors) != 1 || !actors[0].Valid() || actors[0].Kind != rechte.Operator {
		return "", ErrAccessDenied
	}
	return actors[0].ID, nil
}

func (s *Service) organization(ctx context.Context, organizationID string) (string, error) {
	operatorID, err := s.operatorID()
	if err != nil {
		return "", err
	}
	_, err = s.organizations.Get(ctx, organizationID, operatorID)
	if errors.Is(err, portorganisation.ErrNotFound) {
		return "", ErrNotFound
	}
	return operatorID, err
}

func (s *Service) einoAgent(ctx context.Context, organizationID, agentID string) (string, error) {
	operatorID, err := s.organization(ctx, organizationID)
	if err != nil {
		return "", err
	}
	agent, err := s.agents.FindAgent(ctx, organizationID, agentID, operatorID)
	if errors.Is(err, portagent.ErrNotFound) {
		return "", ErrNotFound
	}
	if err != nil {
		return "", err
	}
	if agent.ExecutionKind != domainagent.Eino {
		return "", ErrNotFound
	}
	return operatorID, nil
}

func (s *Service) binding(ctx context.Context) (modellschluessel.Binding, error) {
	binding, err := s.connection.Binding(ctx)
	if errors.Is(err, modellschluessel.ErrMissing) || errors.Is(err, modellschluessel.ErrNotReady) || errors.Is(err, modellschluessel.ErrRevoked) {
		return modellschluessel.Binding{}, ErrConnection
	}
	if err != nil {
		return modellschluessel.Binding{}, err
	}
	if binding.Reference != openrouterverbindung.Reference || binding.Generation == "" {
		return modellschluessel.Binding{}, ErrConnection
	}
	return binding, nil
}
