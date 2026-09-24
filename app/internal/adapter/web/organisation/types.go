package organisation

import (
	"net/http"

	apporganisation "agentcontrolplane/app/internal/app/organisation"
	"agentcontrolplane/ui/bridge"
)

// Handler bindet den Organisationsanwendungsfall an HTTP und die UI-Bridge.
type Handler struct {
	service *apporganisation.Service
	bridge  *bridge.Bridge
	mux     *http.ServeMux
}

type createRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type listResponse struct {
	Organizations any `json:"organizations"`
}

type errorResponse struct {
	Error       string            `json:"error,omitempty"`
	FieldErrors map[string]string `json:"field_errors,omitempty"`
}
