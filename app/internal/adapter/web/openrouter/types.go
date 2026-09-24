package openrouter

import (
	"io"
	"net/http"

	"agentcontrolplane/app/internal/app/openrouterverbindung"
)

type Renderer interface {
	Render(io.Writer, string, map[string]any) error
}

type Handler struct {
	service     *openrouterverbindung.Service
	renderer    Renderer
	mux         *http.ServeMux
	trustedBind bool
}

type response struct {
	Connected    bool   `json:"connected"`
	Reference    string `json:"reference,omitempty"`
	Status       string `json:"status"`
	StatusDetail string `json:"status_detail,omitempty"`
	Notice       string `json:"notice,omitempty"`
	Error        string `json:"error,omitempty"`
}

type saveRequest struct {
	Key string `json:"key"`
}
