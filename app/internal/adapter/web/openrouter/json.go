package openrouter

import (
	"encoding/json"
	"errors"
	"net/http"

	"agentcontrolplane/app/internal/app/openrouterverbindung"
)

func (h *Handler) showJSON(w http.ResponseWriter, r *http.Request) {
	view, err := h.service.Status(r.Context())
	h.writeJSON(w, view, "", err)
}

func (h *Handler) saveJSON(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 16*1024)
	var request saveRequest
	if json.NewDecoder(r.Body).Decode(&request) != nil {
		h.writeJSON(w, openrouterverbindung.View{}, "", errors.New("invalid JSON"))
		return
	}

	view, err := h.service.Save(r.Context(), request.Key)
	h.writeJSON(w, view, "Verbindung gespeichert", err)
}

func (h *Handler) checkJSON(w http.ResponseWriter, r *http.Request) {
	view, err := h.service.Check(r.Context())
	h.writeJSON(w, view, "Status geprüft", err)
}

func (h *Handler) disconnectJSON(w http.ResponseWriter, r *http.Request) {
	view, err := h.service.Disconnect(r.Context())
	h.writeJSON(w, view, "Verbindung getrennt", err)
}
