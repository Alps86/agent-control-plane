package modellwahl

import (
	"context"
	"errors"

	domain "agentcontrolplane/app/internal/domain/modellwahl"
	portagent "agentcontrolplane/app/internal/port/agent"
	portmodellwahl "agentcontrolplane/app/internal/port/modellwahl"
	portorganisation "agentcontrolplane/app/internal/port/organisation"
)

// NewService injects the existing agent store and an explicit access check.
func NewService(agents portagent.Store, identity portorganisation.Identity, store portmodellwahl.Store, access portmodellwahl.OpenRouterAccess, catalog domain.Catalog) *Service {
	return &Service{agents: agents, identity: identity, store: store, access: access, catalog: catalog}
}

// Get exposes the chosen route and the registered, redacted catalog.
func (s *Service) Get(ctx context.Context, organizationID, agentID string) (View, error) {
	agent, operatorID, err := s.agent(ctx, organizationID, agentID)
	if err != nil {
		return View{}, err
	}

	selection, err := s.store.Get(ctx, organizationID, agentID, operatorID)
	if errors.Is(err, portmodellwahl.ErrNotFound) {
		selection = s.defaultSelection()
		err = nil
	}

	if err != nil {
		return View{}, err
	}

	return View{Agent: agent, ExecutionKind: string(agent.ExecutionKind), Selection: selection,
		Providers: s.availableProviders(ctx, organizationID, agentID)}, nil
}

// Put stores exactly one configured route after the current authorization check.
func (s *Service) Put(ctx context.Context, organizationID, agentID string, selection domain.Selection) (View, error) {
	_, operatorID, err := s.agent(ctx, organizationID, agentID)
	if err != nil {
		return View{}, err
	}

	if err := s.validate(ctx, organizationID, agentID, selection); err != nil {
		return View{}, err
	}

	if err := s.store.Replace(ctx, organizationID, agentID, operatorID, selection); err != nil {
		return View{}, err
	}

	return s.Get(ctx, organizationID, agentID)
}

func (s *Service) validate(ctx context.Context, organizationID, agentID string, selection domain.Selection) error {
	_, _, found := s.catalog.Find(selection)
	if !found {
		return ErrCatalog
	}

	if selection.Provider == "openrouter" {
		return s.openRouterAllowed(ctx, organizationID, agentID, selection.ConnectionReference)
	}

	return nil
}

func (s *Service) openRouterAllowed(ctx context.Context, organizationID, agentID, reference string) error {
	if s.access == nil {
		return ErrGrant
	}

	return s.access.Allowed(ctx, organizationID, agentID, reference)
}

func (s *Service) defaultSelection() domain.Selection {
	for _, provider := range s.catalog.Providers {
		if provider.ID == "codex-abo" && len(provider.Connections) > 0 {
			return domain.Selection{Provider: provider.ID, ConnectionReference: provider.Connections[0].Reference, Model: provider.FirstModel()}
		}
	}

	return domain.Selection{Provider: "codex-abo"}
}

func (s *Service) availableProviders(ctx context.Context, organizationID, agentID string) []domain.Provider {
	providers := append([]domain.Provider(nil), s.catalog.Providers...)
	for i := range providers {
		providers[i] = s.availability(ctx, organizationID, agentID, providers[i])
	}

	return providers
}

func (s *Service) availability(ctx context.Context, organizationID, agentID string, provider domain.Provider) domain.Provider {
	if provider.ID != "openrouter" {
		return s.catalogSelectability(provider)
	}

	if len(provider.Connections) == 0 {
		return provider
	}

	err := s.openRouterAllowed(ctx, organizationID, agentID, provider.Connections[0].Reference)
	provider.Selectable = err == nil && len(provider.Models) > 0
	if err != nil {
		provider.Reason = err.Error()
	}

	if err == nil && len(provider.Models) == 0 {
		provider.Reason = "Kein OpenRouter-Modell konfiguriert"
	}

	return provider
}

func (s *Service) catalogSelectability(provider domain.Provider) domain.Provider {
	provider.Selectable = len(provider.Models) > 0
	if len(provider.Models) == 0 {
		provider.Reason = "Kein Modell konfiguriert"
	}

	return provider
}
