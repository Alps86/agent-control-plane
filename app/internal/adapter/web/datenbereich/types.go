package datenbereich

import (
	"net/http"

	appdatenbereich "agentcontrolplane/app/internal/app/datenbereich"
	"agentcontrolplane/ui/bridge"
)

// Handler binds the operator's agent data scope to HTTP and the UI bridge.
type Handler struct {
	service     *appdatenbereich.Service
	bridge      *bridge.Bridge
	bindAddress string
	mux         *http.ServeMux
}

type replaceRequest struct {
	Scopes []appdatenbereich.ScopeInput `json:"scopes"`
}

type scopeResponse struct {
	Scopes any `json:"scopes"`
}

type errorResponse struct {
	Error string `json:"error"`
}
