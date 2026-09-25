package modellfreigabe

import (
	"io"
	"net/http"

	appfreigabe "agentcontrolplane/app/internal/app/modellfreigabe"
)

type Renderer interface {
	Render(io.Writer, string, map[string]any) error
}

// Handler hält nur die öffentliche Bediengrenze und den redigierten Anwendungsfall.
type Handler struct {
	service     *appfreigabe.Service
	renderer    Renderer
	bindAddress string
	mux         *http.ServeMux
}

type errorResponse struct {
	Error string `json:"error"`
}
