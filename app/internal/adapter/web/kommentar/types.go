package kommentar

import (
	"net/http"

	appaufgabe "agentcontrolplane/app/internal/app/aufgabe"
	appkommentar "agentcontrolplane/app/internal/app/kommentar"
	apporganisation "agentcontrolplane/app/internal/app/organisation"
	appprojekt "agentcontrolplane/app/internal/app/projekt"
	"agentcontrolplane/ui/bridge"
)

// Handler verbindet den Aufgabenthread mit der öffentlichen HTTP-Grenze.
type Handler struct {
	comments      *appkommentar.Service
	tasks         *appaufgabe.Service
	projects      *appprojekt.Service
	organizations *apporganisation.Service
	bridge        *bridge.Bridge
	bindAddress   string
	mux           *http.ServeMux
}

type errorResponse struct {
	Error       string            `json:"error,omitempty"`
	FieldErrors map[string]string `json:"field_errors,omitempty"`
}

type createRequest struct {
	Content   string            `json:"content"`
	Reference *referenceRequest `json:"reference,omitempty"`
}

type referenceRequest struct {
	Type           string `json:"type"`
	ID             string `json:"id"`
	OrganizationID string `json:"organization_id"`
}

type updateRequest struct {
	Content string `json:"content"`
}
