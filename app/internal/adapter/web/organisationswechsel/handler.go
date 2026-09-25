package organisationswechsel

import (
	"net/http"

	apporganisation "agentcontrolplane/app/internal/app/organisation"
	appregel "agentcontrolplane/app/internal/app/organisationsregel"
	"agentcontrolplane/ui/bridge"
)

// NewHandler konstruiert die eigenständigen Such-, Wechsel- und Regelrouten.
func NewHandler(organizations *apporganisation.Service, rules *appregel.Service, ui *bridge.Bridge, bindAddress string) *Handler {
	h := &Handler{organizations: organizations, rules: rules, bridge: ui, bindAddress: bindAddress, mux: http.NewServeMux()}
	h.mux.HandleFunc("GET /api/organisationswechsel", h.apiSearch)
	h.mux.HandleFunc("POST /api/organisationen/{id}/wechsel", h.apiSelect)
	h.mux.HandleFunc("GET /api/organisationen/{id}/arbeitsregeln", h.apiRules)
	h.mux.HandleFunc("PUT /api/organisationen/{id}/arbeitsregeln", h.apiSave)
	h.mux.HandleFunc("DELETE /api/organisationen/{id}/arbeitsregeln", h.apiRevoke)
	h.mux.HandleFunc("POST /api/organisationen/{id}/arbeitsregeln/vorschau", h.apiPreview)
	h.mux.HandleFunc("GET /organisationswechsel", h.pageSearch)
	h.mux.HandleFunc("POST /organisationswechsel/{id}", h.pageSelect)
	h.mux.HandleFunc("GET /organisationen/{id}/arbeitsregeln", h.pageRules)
	h.mux.HandleFunc("POST /organisationen/{id}/arbeitsregeln", h.pageSave)
	h.mux.HandleFunc("POST /organisationen/{id}/arbeitsregeln/widerruf", h.pageRevoke)
	h.mux.HandleFunc("POST /organisationen/{id}/arbeitsregeln/vorschau", h.pagePreview)
	return h
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

func (h *Handler) apiSearch(w http.ResponseWriter, r *http.Request) {
	organizations, err := h.search(r)
	if err != nil {
		h.apiError(w, err)
		return
	}

	h.json(w, http.StatusOK, searchResponse{Organizations: organizations})
}

func (h *Handler) apiSelect(w http.ResponseWriter, r *http.Request) {
	if !h.localWrite(r) {
		h.json(w, http.StatusForbidden, errorResponse{Error: "access_denied"})
		return
	}

	organization, err := h.organizations.Get(r.Context(), r.PathValue("id"))
	if err != nil {
		h.apiError(w, err)
		return
	}

	h.selectOrganization(w, r, organization.ID)
	h.json(w, http.StatusOK, selectionResponse{SelectedOrganizationID: organization.ID})
}

func (h *Handler) apiRules(w http.ResponseWriter, r *http.Request) {
	snapshot, err := h.rules.Read(r.Context(), r.PathValue("id"))
	if err != nil {
		h.apiError(w, err)
		return
	}

	h.json(w, http.StatusOK, snapshot)
}

func (h *Handler) apiPreview(w http.ResponseWriter, r *http.Request) {
	if !h.localWrite(r) {
		h.json(w, http.StatusForbidden, errorResponse{Error: "access_denied"})
		return
	}

	change, err := h.decode(w, r)
	if err != nil {
		h.json(w, http.StatusBadRequest, errorResponse{Error: "invalid_json"})
		return
	}

	h.previewJSON(w, r, change)
}

func (h *Handler) previewJSON(w http.ResponseWriter, r *http.Request, change appregel.Change) {
	preview, err := h.rules.Preview(r.Context(), r.PathValue("id"), change)
	if err != nil {
		h.apiError(w, err)
		return
	}

	h.json(w, http.StatusOK, preview)
}

func (h *Handler) apiSave(w http.ResponseWriter, r *http.Request) {
	if !h.localWrite(r) {
		h.json(w, http.StatusForbidden, errorResponse{Error: "access_denied"})
		return
	}

	change, err := h.decode(w, r)
	if err != nil {
		h.json(w, http.StatusBadRequest, errorResponse{Error: "invalid_json"})
		return
	}

	change.Operation = "allow"
	h.saveJSON(w, r, change)
}

func (h *Handler) apiRevoke(w http.ResponseWriter, r *http.Request) {
	if !h.localWrite(r) {
		h.json(w, http.StatusForbidden, errorResponse{Error: "access_denied"})
		return
	}

	change, err := h.decode(w, r)
	if err != nil {
		h.json(w, http.StatusBadRequest, errorResponse{Error: "invalid_json"})
		return
	}

	change.Operation = "revoke"
	h.saveJSON(w, r, change)
}

func (h *Handler) saveJSON(w http.ResponseWriter, r *http.Request, change appregel.Change) {
	snapshot, err := h.rules.Save(r.Context(), r.PathValue("id"), change)
	if err != nil {
		h.apiError(w, err)
		return
	}

	h.json(w, http.StatusOK, snapshot)
}
