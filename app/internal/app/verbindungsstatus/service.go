package verbindungsstatus

import (
	"context"

	"agentcontrolplane/app/internal/app/modellverbindung"
	"agentcontrolplane/app/internal/app/openrouterverbindung"
	"agentcontrolplane/app/internal/domain/modellwahl"
)

func NewService(catalog modellwahl.Catalog, sources ...Source) *Service {
	return &Service{catalog: catalog, sources: sources}
}

func NewCodexSource(service CodexStatus) Source {
	return Source{ProviderID: "codex-abo", Name: "Codex-Abo", AuthType: "device_code",
		Reference: modellverbindung.CredentialKey, Port: codexPort{service}}
}

func NewOpenRouterSource(service OpenRouterStatus) Source {
	return Source{ProviderID: "openrouter", Name: "OpenRouter", AuthType: "api_key",
		Reference: openrouterverbindung.Reference, Port: openRouterPort{service}}
}

func (s *Service) Get(ctx context.Context) View {
	view := View{Providers: make([]Provider, 0, len(s.sources))}
	seen := map[string]bool{}
	for _, source := range s.sources {
		if !s.valid(source) || seen[source.ProviderID] {
			continue
		}

		seen[source.ProviderID] = true
		view.Providers = append(view.Providers, s.provider(ctx, source))
	}

	return view
}

func (s *Service) Available(ctx context.Context, providerID, reference string) (bool, error) {
	for _, source := range s.sources {
		if s.valid(source) && source.ProviderID == providerID && source.Reference == reference {
			provider := s.provider(ctx, source)
			return provider.Ready && provider.ConnectionReference == reference, nil
		}
	}

	return false, nil
}

func (s *Service) valid(source Source) bool {
	return source.ProviderID != "" && source.AuthType != "" && source.Reference != "" && source.Port != nil
}

func (s *Service) provider(ctx context.Context, source Source) Provider {
	provider := Provider{ID: source.ProviderID, Name: s.name(source), AuthType: source.AuthType, Status: "unavailable"}
	state, err := source.Port.Read(ctx)
	if err != nil {
		return provider
	}

	provider.Status = s.publicState(state.Status)
	provider.Ready = state.Ready && provider.Status == "connected"
	if provider.Status == "connected" && !provider.Ready {
		provider.Status = "not_ready"
	}

	provider.StatusDetail = s.publicDetail(provider.Status)
	if provider.Ready || provider.Status == "not_ready" {
		provider.ConnectionReference = source.Reference
	}

	return provider
}

func (s *Service) name(source Source) string {
	for _, candidate := range s.catalog.Providers {
		if candidate.ID == source.ProviderID && candidate.Name != "" {
			return candidate.Name
		}
	}

	return source.Name
}

func (s *Service) publicState(state string) string {
	if state == "connected" || state == "pending" || state == "reauthentication_required" {
		return state
	}

	if state == "disconnected" || state == "not_ready" || state == "unavailable" {
		return state
	}

	return "unavailable"
}

func (s *Service) publicDetail(status string) string {
	if status == "reauthentication_required" {
		return "Erneute Anmeldung erforderlich"
	}

	if status == "not_ready" {
		return "Verbindung ist nicht einsatzbereit"
	}

	return ""
}
