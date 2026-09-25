package projektort

import (
	"net/http"

	apporganisation "agentcontrolplane/app/internal/app/organisation"
	appprojekt "agentcontrolplane/app/internal/app/projekt"
	apport "agentcontrolplane/app/internal/app/projektort"
	"agentcontrolplane/ui/bridge"
)

// Handler stellt die gebundene Projektordneransicht und Startprüfung bereit.
type Handler struct {
	locations     *apport.Service
	projects      *appprojekt.Service
	organizations *apporganisation.Service
	bridge        *bridge.Bridge
	bindAddress   string
	mux           *http.ServeMux
}

type bindRequest struct {
	Kind string `json:"kind"`
}

type errorResponse struct {
	Error  string `json:"error"`
	Reason string `json:"reason,omitempty"`
}
