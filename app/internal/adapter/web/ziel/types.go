package ziel

import (
	"net/http"

	apporganisation "agentcontrolplane/app/internal/app/organisation"
	appziel "agentcontrolplane/app/internal/app/ziel"
	domainziel "agentcontrolplane/app/internal/domain/ziel"
	"agentcontrolplane/ui/bridge"
)

// Handler bindet Ziele an die öffentlichen JSON- und HTML-Grenzen.
type Handler struct {
	goals         *appziel.Service
	organizations *apporganisation.Service
	bridge        *bridge.Bridge
	bindAddress   string
	mux           *http.ServeMux
}

type createRequest struct {
	Name string `json:"name"`
}

type listResponse struct {
	Goals any `json:"goals"`
}

type treeResponse struct {
	Goals        []domainziel.Goal        `json:"goals"`
	ProjectPaths []domainziel.ProjectPath `json:"project_paths"`
}

type statusRequest struct {
	Status string `json:"status"`
}

type goalNode struct {
	ID             string
	OrganizationID string
	Name           string
	Status         string
	StatusLabel    string
	ChildName      string
	ChildError     string
	Children       []goalNode
}

type errorResponse struct {
	Error       string            `json:"error,omitempty"`
	FieldErrors map[string]string `json:"field_errors,omitempty"`
}
