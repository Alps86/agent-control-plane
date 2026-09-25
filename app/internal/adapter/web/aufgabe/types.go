package aufgabe

import (
	"net/http"

	appagent "agentcontrolplane/app/internal/app/agent"
	appaufgabe "agentcontrolplane/app/internal/app/aufgabe"
	apporganisation "agentcontrolplane/app/internal/app/organisation"
	appprojekt "agentcontrolplane/app/internal/app/projekt"
	appprojektarchiv "agentcontrolplane/app/internal/app/projektarchiv"
	"agentcontrolplane/ui/bridge"
)

// Handler verbindet Aufgaben-Anwendungsfälle mit JSON und Browserseiten.
type Handler struct {
	tasks         *appaufgabe.Service
	projects      *appprojekt.Service
	archive       *appprojektarchiv.Service
	agents        *appagent.Service
	organizations *apporganisation.Service
	bridge        *bridge.Bridge
	bindAddress   string
	mux           *http.ServeMux
}

type createRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Priority    string `json:"priority"`
	AssigneeID  string `json:"assignee_id"`
}

type listResponse struct {
	Tasks any `json:"tasks"`
}

type errorResponse struct {
	Error       string            `json:"error,omitempty"`
	FieldErrors map[string]string `json:"field_errors,omitempty"`
}
