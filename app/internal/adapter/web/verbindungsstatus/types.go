package verbindungsstatus

import (
	"net/http"

	app "agentcontrolplane/app/internal/app/verbindungsstatus"
	"agentcontrolplane/ui/bridge"
)

type Handler struct {
	service *app.Service
	ui      *bridge.Bridge
	mux     *http.ServeMux
}
