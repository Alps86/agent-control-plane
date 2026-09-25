package verbindungsstatus

import (
	"context"

	"agentcontrolplane/app/internal/app/modellverbindung"
	"agentcontrolplane/app/internal/app/openrouterverbindung"
	"agentcontrolplane/app/internal/domain/modellwahl"
)

type CodexStatus interface {
	Status(context.Context) (modellverbindung.Status, error)
}

type OpenRouterStatus interface {
	Status(context.Context) (openrouterverbindung.View, error)
}

// StatusPort reads a mounted provider without exposing its credentials.
type StatusPort interface {
	Read(context.Context) (State, error)
}

type State struct {
	Status string
	Ready  bool
}

// Source is an explicitly mounted status source, separate from model dispatch.
type Source struct {
	ProviderID string
	Name       string
	AuthType   string
	Reference  string
	Port       StatusPort
}

type Provider struct {
	ID                  string `json:"id"`
	Name                string `json:"name"`
	AuthType            string `json:"auth_type"`
	Status              string `json:"status"`
	StatusDetail        string `json:"status_detail"`
	ConnectionReference string `json:"connection_reference"`
	Ready               bool   `json:"ready"`
}

type View struct {
	Providers []Provider `json:"providers"`
}

type Service struct {
	catalog modellwahl.Catalog
	sources []Source
}

type codexPort struct{ service CodexStatus }
type openRouterPort struct{ service OpenRouterStatus }
