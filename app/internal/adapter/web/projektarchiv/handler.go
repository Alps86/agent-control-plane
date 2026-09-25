package projektarchiv

import (
	"encoding/json"
	"errors"
	"net/http"

	apporganisation "agentcontrolplane/app/internal/app/organisation"
	appprojektarchiv "agentcontrolplane/app/internal/app/projektarchiv"
	appziel "agentcontrolplane/app/internal/app/ziel"
	domainprojektarchiv "agentcontrolplane/app/internal/domain/projektarchiv"
	"agentcontrolplane/ui/bridge"
)

// NewHandler registriert die Archiv-Routen ohne bestehende Projektrouten zu ändern.
func NewHandler(archive *appprojektarchiv.Service, organizations *apporganisation.Service, goals *appziel.Service, ui *bridge.Bridge, bindAddress string) *Handler {
	h := &Handler{archive: archive, organizations: organizations, goals: goals, bridge: ui, bindAddress: bindAddress, mux: http.NewServeMux()}
	h.mux.HandleFunc("GET /api/organisationen/{id}/projekte/archiv", h.apiList)
	h.mux.HandleFunc("POST /api/organisationen/{id}/projekte/{projektID}/archivieren", h.apiArchive)
	h.mux.HandleFunc("POST /api/organisationen/{id}/projekte/{projektID}/wiederherstellen", h.apiRestore)
	h.mux.HandleFunc("GET /api/organisationen/{id}/projekte/{projektID}/status", h.apiStatus)
	h.mux.HandleFunc("GET /organisationen/{id}/projekte/archiv", h.pageList)
	h.mux.HandleFunc("POST /organisationen/{id}/projekte/{projektID}/archivieren", h.pageArchive)
	h.mux.HandleFunc("POST /organisationen/{id}/projekte/{projektID}/restaurieren", h.pageRestore)
	return h
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

func (h *Handler) apiArchive(w http.ResponseWriter, r *http.Request) {
	if !h.localWrite(r) {
		h.json(w, http.StatusForbidden, errorResponse{Error: "access_denied"})
		return
	}

	if err := h.archive.Archive(r.Context(), r.PathValue("id"), r.PathValue("projektID")); err != nil {
		h.apiError(w, err)
		return
	}

	h.json(w, http.StatusOK, statusResponse{Status: domainprojektarchiv.StatusArchived})
}

func (h *Handler) apiRestore(w http.ResponseWriter, r *http.Request) {
	if !h.localWrite(r) {
		h.json(w, http.StatusForbidden, errorResponse{Error: "access_denied"})
		return
	}

	if err := h.archive.Restore(r.Context(), r.PathValue("id"), r.PathValue("projektID")); err != nil {
		h.apiError(w, err)
		return
	}

	h.json(w, http.StatusOK, statusResponse{Status: domainprojektarchiv.StatusActive})
}

func (h *Handler) apiStatus(w http.ResponseWriter, r *http.Request) {
	status, err := h.archive.Status(r.Context(), r.PathValue("id"), r.PathValue("projektID"))
	if err != nil {
		h.apiError(w, err)
		return
	}

	h.json(w, http.StatusOK, statusResponse{Status: status})
}

func (h *Handler) apiList(w http.ResponseWriter, r *http.Request) {
	projects, err := h.archive.ListArchived(r.Context(), r.PathValue("id"))
	if err != nil {
		h.apiError(w, err)
		return
	}

	h.json(w, http.StatusOK, listResponse{Projects: projects})
}

func (h *Handler) apiError(w http.ResponseWriter, err error) {
	if errors.Is(err, apporganisation.ErrAccessDenied) {
		h.json(w, http.StatusForbidden, errorResponse{Error: "access_denied"})
		return
	}

	if errors.Is(err, apporganisation.ErrNotFound) || errors.Is(err, appprojektarchiv.ErrNotFound) {
		h.json(w, http.StatusNotFound, errorResponse{Error: "not_found"})
		return
	}

	h.apiConflict(w, err)
}

func (h *Handler) apiConflict(w http.ResponseWriter, err error) {
	if errors.Is(err, appprojektarchiv.ErrActiveRun) {
		h.json(w, http.StatusConflict, errorResponse{Error: "active_run"})
		return
	}

	if errors.Is(err, appprojektarchiv.ErrAlreadyArchived) || errors.Is(err, appprojektarchiv.ErrNotArchived) {
		h.json(w, http.StatusConflict, errorResponse{Error: "invalid_status"})
		return
	}

	h.json(w, http.StatusInternalServerError, errorResponse{Error: "internal_error"})
}

func (h *Handler) json(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
