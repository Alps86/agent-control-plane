package projektarchiv

import (
	"net/http"

	apporganisation "agentcontrolplane/app/internal/app/organisation"
	appprojektarchiv "agentcontrolplane/app/internal/app/projektarchiv"
	appziel "agentcontrolplane/app/internal/app/ziel"
	domainprojektarchiv "agentcontrolplane/app/internal/domain/projektarchiv"
	"agentcontrolplane/ui/bridge"
)

// Handler bindet Archivaktionen und Archivansicht an HTTP.
type Handler struct {
	archive       *appprojektarchiv.Service
	organizations *apporganisation.Service
	goals         *appziel.Service
	bridge        *bridge.Bridge
	bindAddress   string
	mux           *http.ServeMux
}

type statusResponse struct {
	Status domainprojektarchiv.Status `json:"status"`
}

type listResponse struct {
	Projects []domainprojektarchiv.ArchivedProject `json:"projects"`
}

type errorResponse struct {
	Error string `json:"error"`
}
