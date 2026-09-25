package organisationswechsel

import (
	"net/http"

	apporganisation "agentcontrolplane/app/internal/app/organisation"
	appregel "agentcontrolplane/app/internal/app/organisationsregel"
	"agentcontrolplane/ui/bridge"
)

// Handler stellt die Organisationswahl und ihre Arbeitsregeln bereit.
type Handler struct {
	organizations *apporganisation.Service
	rules         *appregel.Service
	bridge        *bridge.Bridge
	bindAddress   string
	mux           *http.ServeMux
}

type searchResponse struct {
	Organizations any `json:"organizations"`
}

type selectionResponse struct {
	SelectedOrganizationID string `json:"selected_organization_id"`
}

type errorResponse struct {
	Error string `json:"error"`
}
