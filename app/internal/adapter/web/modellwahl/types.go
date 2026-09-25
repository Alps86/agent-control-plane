package modellwahl

import (
	"net/http"

	app "agentcontrolplane/app/internal/app/modellwahl"
	"agentcontrolplane/ui/bridge"
)

// Handler exposes one organization's Eino agent selection.
type Handler struct {
	service     *app.Service
	ui          *bridge.Bridge
	bindAddress string
	mux         *http.ServeMux
}

type errorResponse struct {
	Error  string `json:"error"`
	Reason string `json:"reason,omitempty"`
}
