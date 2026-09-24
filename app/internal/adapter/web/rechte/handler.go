package rechte

import (
	"encoding/json"
	"net/http"

	apprechte "agentcontrolplane/app/internal/app/rechte"
	portrechte "agentcontrolplane/app/internal/port/rechte"
)

// NewHandler bindet die serverseitige Quelle an den öffentlichen Beispielweg.
func NewHandler(directory portrechte.Directory) http.Handler {
	handler := &Handler{reader: apprechte.NewReader(directory)}
	mux := http.NewServeMux()
	mux.Handle("GET /api/architektur/rechte/ressourcen/{id}", handler)
	return mux
}

// ServeHTTP gibt Ressourcendaten nur nach der zentralen Rechteprüfung aus.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	resource, outcome := h.reader.Read(r.PathValue("id"))
	w.Header().Set("Content-Type", "application/json")
	if outcome == apprechte.AccessDenied {
		h.writeError(w, http.StatusForbidden, `{"error":"access_denied"}`)
		return
	}

	if outcome != apprechte.Allowed {
		h.writeError(w, http.StatusNotFound, `{"error":"resource_not_found"}`)
		return
	}

	response := resourceResponse{resource.ID, resource.Name, resource.OrganizationID}
	_ = json.NewEncoder(w).Encode(response)
}

func (h *Handler) writeError(w http.ResponseWriter, status int, body string) {
	w.WriteHeader(status)
	_, _ = w.Write([]byte(body))
}
