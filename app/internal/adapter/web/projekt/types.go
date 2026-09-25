package projekt

import (
	"net/http"

	appaufgabe "agentcontrolplane/app/internal/app/aufgabe"
	apporganisation "agentcontrolplane/app/internal/app/organisation"
	appprojekt "agentcontrolplane/app/internal/app/projekt"
	appziel "agentcontrolplane/app/internal/app/ziel"
	domainprojekt "agentcontrolplane/app/internal/domain/projekt"
	"agentcontrolplane/ui/bridge"
)

// Handler bindet Projekte an die öffentlichen JSON- und HTML-Routen.
type Handler struct {
	projects      *appprojekt.Service
	tasks         *appaufgabe.Service
	organizations *apporganisation.Service
	goals         *appziel.Service
	bridge        *bridge.Bridge
	bindAddress   string
	mux           *http.ServeMux
}

type createRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	GoalID      string `json:"goal_id"`
}

type listResponse struct {
	Projects any `json:"projects"`
}

type detailResponse struct {
	domainprojekt.Project
	Tasks any `json:"tasks"`
}

type errorResponse struct {
	Error       string            `json:"error,omitempty"`
	FieldErrors map[string]string `json:"field_errors,omitempty"`
}
