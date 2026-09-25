package modellwahl

import (
	"strings"
	"time"
)

// Find returns only an exactly configured provider, connection and model.
func (c Catalog) Find(selection Selection) (Provider, Model, bool) {
	for _, provider := range c.Providers {
		if provider.ID == selection.Provider && provider.hasConnection(selection.ConnectionReference) {
			return provider.findModel(selection.Model)
		}
	}

	return Provider{}, Model{}, false
}

func (p Provider) hasConnection(reference string) bool {
	for _, connection := range p.Connections {
		if connection.Reference == reference && reference != "" {
			return true
		}
	}

	return false
}

func (p Provider) findModel(id string) (Provider, Model, bool) {
	for _, model := range p.Models {
		if model.ID == id && id != "" {
			return p, model, true
		}
	}

	return Provider{}, Model{}, false
}

// FirstModel returns the first configured candidate without claiming availability.
func (p Provider) FirstModel() string {
	if len(p.Models) == 0 {
		return ""
	}

	return p.Models[0].ID
}

// Valid rejects ambiguous or incomplete catalog declarations.
func (c Catalog) Valid() bool {
	seen := map[string]bool{}
	for _, provider := range c.Providers {
		if strings.TrimSpace(provider.ID) == "" || seen[provider.ID] || !provider.valid() {
			return false
		}

		seen[provider.ID] = true
	}

	return true
}

func (p Provider) valid() bool {
	if p.Name == "" || p.AuthType == "" || len(p.Connections) == 0 {
		return false
	}

	return p.validConnections() && p.ValidOpenRouterConnection() && p.validModels()
}

// ValidOpenRouterConnection keeps the one central Story-23 reference explicit.
func (p Provider) ValidOpenRouterConnection() bool {
	if p.ID != "openrouter" {
		return true
	}

	return len(p.Connections) == 1 && p.Connections[0].Reference == "openrouter-central"
}

func (p Provider) validConnections() bool {
	seenConnections := map[string]bool{}
	for _, connection := range p.Connections {
		if connection.Reference == "" || seenConnections[connection.Reference] {
			return false
		}

		seenConnections[connection.Reference] = true
	}

	return true
}

func (p Provider) validModels() bool {
	seenModels := map[string]bool{}
	for _, model := range p.Models {
		if seenModels[model.ID] || !model.Valid() {
			return false
		}

		seenModels[model.ID] = true
	}

	return true
}

// Valid requires a traceable catalog entry and explicit capability states.
func (m Model) Valid() bool {
	if strings.TrimSpace(m.ID) == "" || m.Source == "" || m.CheckStatus == "" {
		return false
	}

	if _, err := time.Parse(time.RFC3339, m.ObservedAt); err != nil {
		return false
	}

	for _, capability := range m.Capabilities {
		if !capability.Valid() {
			return false
		}
	}

	return true
}

func (c Capability) Valid() bool {
	if c.Status == "" {
		return false
	}

	if c.Status != "verified" {
		return true
	}

	if c.Source == "" {
		return false
	}

	_, err := time.Parse(time.RFC3339, c.CheckedAt)
	return err == nil
}
