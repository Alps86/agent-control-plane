package aktivitaet

import (
	"net/http"

	appaktivitaet "agentcontrolplane/app/internal/app/aktivitaet"
	apporganisation "agentcontrolplane/app/internal/app/organisation"
	"agentcontrolplane/ui/bridge"
)

// Handler zeigt den organisationsgebundenen fachlichen Aktivitätsverlauf.
type Handler struct {
	service       *appaktivitaet.Service
	organizations *apporganisation.Service
	bridge        *bridge.Bridge
	mux           *http.ServeMux
}

type listResponse struct {
	Events any `json:"events"`
}

type errorResponse struct {
	Error string `json:"error"`
}
