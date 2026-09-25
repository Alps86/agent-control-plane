package agent

import (
	"net/http"

	appagent "agentcontrolplane/app/internal/app/agent"
	appberichtsweg "agentcontrolplane/app/internal/app/berichtsweg"
	"agentcontrolplane/ui/bridge"
)

// Handler exposes only organization-scoped agent operations.
type Handler struct {
	service     *appagent.Service
	reporting   *appberichtsweg.Service
	bridge      *bridge.Bridge
	bindAddress string
	mux         *http.ServeMux
}

type profileResponse struct {
	appagent.Profile
	ParentID   string `json:"parent_id"`
	ParentName string `json:"parent_name"`
}

type listResponse struct {
	Agents any `json:"agents"`
}

type templateResponse struct {
	Templates any `json:"templates"`
}

type teamResponse struct {
	TeamTemplates any `json:"team_templates"`
}

type errorResponse struct {
	Error       string            `json:"error,omitempty"`
	Reason      string            `json:"reason,omitempty"`
	FieldErrors map[string]string `json:"field_errors,omitempty"`
}
