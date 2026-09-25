package modellwahl

import (
	"encoding/json"
	"fmt"
	"os"

	domain "agentcontrolplane/app/internal/domain/modellwahl"
)

// NewCatalogLoader constructs the non-secret configuration loader.
func NewCatalogLoader() *CatalogLoader { return &CatalogLoader{} }

// Load reads non-secret registered model candidates from a local file.
func (l *CatalogLoader) Load(path string) (domain.Catalog, error) {
	if path == "" {
		return domain.Catalog{Providers: []domain.Provider{{
			ID: "codex-abo", Name: "Codex-Abo", AuthType: "device_code",
			Connections: []domain.Connection{{Reference: "codex-chatgpt", Label: "Codex-Abo"}},
		}}}, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return domain.Catalog{}, fmt.Errorf("Modellkatalog lesen: %w", err)
	}

	return l.decode(data)
}

func (l *CatalogLoader) decode(data []byte) (domain.Catalog, error) {
	var catalog domain.Catalog
	if err := json.Unmarshal(data, &catalog); err != nil {
		return domain.Catalog{}, fmt.Errorf("Modellkatalog dekodieren: %w", err)
	}

	for i := range catalog.Providers {
		catalog.Providers[i] = l.normalizeProvider(catalog.Providers[i])
	}

	if err := l.validate(catalog); err != nil {
		return domain.Catalog{}, err
	}

	return catalog, nil
}

func (l *CatalogLoader) validate(catalog domain.Catalog) error {
	for _, provider := range catalog.Providers {
		if provider.ID == "openrouter" && !provider.ValidOpenRouterConnection() {
			return ErrOpenRouterCatalogConnection
		}
	}

	if !catalog.Valid() {
		return ErrCatalog
	}

	return nil
}

func (l *CatalogLoader) normalizeProvider(provider domain.Provider) domain.Provider {
	if provider.ID == "codex-abo" && len(provider.Connections) == 0 {
		provider.Connections = []domain.Connection{{Reference: "codex-chatgpt", Label: "Codex-Abo"}}
	}

	for i := range provider.Models {
		provider.Models[i] = l.normalizeModel(provider.Models[i])
	}

	return provider
}

func (l *CatalogLoader) normalizeModel(model domain.Model) domain.Model {
	model.AccountVerified = false
	model.RouteVerified = false
	model.CheckStatus = "unverified"

	if model.Capabilities == nil {
		model.Capabilities = map[string]domain.Capability{}
	}

	for _, name := range []string{"text", "tools", "audio_input", "audio_output", "realtime"} {
		if model.Capabilities[name].Status == "" {
			model.Capabilities[name] = domain.Capability{Status: "unverified"}
		}
	}

	return model
}
