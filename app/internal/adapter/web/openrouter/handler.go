package openrouter

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"agentcontrolplane/app/internal/app/openrouterverbindung"
)

func NewHandler(service *openrouterverbindung.Service, renderer Renderer, bindAddress string) http.Handler {
	h := &Handler{service: service, renderer: renderer, mux: http.NewServeMux()}
	h.trustedBind = h.numericLoopbackBind(bindAddress)
	h.mux.HandleFunc("GET /settings/modellanbieter/openrouter", h.showHTML)
	h.mux.HandleFunc("POST /settings/modellanbieter/openrouter", h.saveHTML)
	h.mux.HandleFunc("POST /settings/modellanbieter/openrouter/pruefen", h.checkHTML)
	h.mux.HandleFunc("POST /settings/modellanbieter/openrouter/trennen", h.disconnectHTML)
	h.mux.HandleFunc("GET /api/settings/modellanbieter/openrouter", h.showJSON)
	h.mux.HandleFunc("POST /api/settings/modellanbieter/openrouter", h.saveJSON)
	h.mux.HandleFunc("POST /api/settings/modellanbieter/openrouter/pruefen", h.checkJSON)
	h.mux.HandleFunc("DELETE /api/settings/modellanbieter/openrouter", h.disconnectJSON)
	return h
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Referrer-Policy", "same-origin")
	if !h.trustedBind || !h.localRequest(r) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	if r.Method == http.MethodPost || r.Method == http.MethodDelete {
		if !h.originAllowed(r) {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
	}

	h.mux.ServeHTTP(w, r)
}

func (h *Handler) showHTML(w http.ResponseWriter, r *http.Request) {
	view, err := h.service.Status(r.Context())
	h.renderHTML(w, r, view, "", err)
}

func (h *Handler) saveHTML(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 16*1024)
	if r.ParseForm() != nil {
		h.renderHTML(w, r, openrouterverbindung.View{}, "", errors.New("Ungültige Eingabe"))
		return
	}

	view, err := h.service.Save(r.Context(), r.PostForm.Get("key"))
	if err != nil {
		current, statusErr := h.service.Status(r.Context())
		if statusErr == nil {
			view = current
		}
	}

	h.renderHTML(w, r, view, "Verbindung gespeichert", err)
}

func (h *Handler) checkHTML(w http.ResponseWriter, r *http.Request) {
	view, err := h.service.Check(r.Context())
	h.renderHTML(w, r, view, "Status geprüft", err)
}

func (h *Handler) disconnectHTML(w http.ResponseWriter, r *http.Request) {
	view, err := h.service.Disconnect(r.Context())
	h.renderHTML(w, r, view, "Verbindung getrennt", err)
}

func (h *Handler) renderHTML(w http.ResponseWriter, r *http.Request, view openrouterverbindung.View, notice string, err error) {
	status := http.StatusOK
	errorsList := []string{}
	if err != nil {
		status, errorsList = h.publicError(err)
		notice = ""
	}

	data := h.data(view, notice, errorsList)
	name := "modelle/openrouter/page"
	if strings.EqualFold(r.Header.Get("HX-Request"), "true") {
		name = "modelle/openrouter/content"
	}

	h.writeHTML(w, name, data, status)
}

func (h *Handler) writeHTML(w http.ResponseWriter, name string, data map[string]any, status int) {
	if h.renderer == nil {
		http.Error(w, "Ansicht derzeit nicht verfügbar", http.StatusServiceUnavailable)
		return
	}

	var rendered bytes.Buffer
	if h.renderer.Render(&rendered, name, data) != nil {
		http.Error(w, "Ansicht derzeit nicht verfügbar", http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	_, _ = w.Write(rendered.Bytes())
}

func (h *Handler) data(view openrouterverbindung.View, notice string, errorsList []string) map[string]any {
	return map[string]any{
		"PageTitle": "OpenRouter · Modellanbieter · Settings",
		"Notice":    notice,
		"Errors":    errorsList,
		"View": map[string]any{"Settings": map[string]any{"OpenRouter": map[string]any{
			"Connected": view.Connected, "Reference": view.Reference,
			"Status": view.Status, "StatusDetail": view.StatusDetail,
		}}},
	}
}

func (h *Handler) publicError(err error) (int, []string) {
	if errors.Is(err, openrouterverbindung.ErrMissingKey) {
		return http.StatusBadRequest, []string{"Schlüssel fehlt"}
	}

	if errors.Is(err, openrouterverbindung.ErrStorage) {
		return http.StatusServiceUnavailable, []string{"Verbindung derzeit nicht verfügbar"}
	}

	return http.StatusBadRequest, []string{"Ungültige Eingabe"}
}

func (h *Handler) writeJSON(w http.ResponseWriter, view openrouterverbindung.View, notice string, err error) {
	result := response{view.Connected, view.Reference, view.Status, view.StatusDetail, notice, ""}
	status := http.StatusOK
	if err != nil {
		code, public := h.publicError(err)
		status, result = code, response{Error: public[0]}
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(result)
}
