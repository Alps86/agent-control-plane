package modellfreigabe

import (
	"bytes"
	"encoding/json"
	"net/http"
)

func (h *Handler) apiShow(w http.ResponseWriter, r *http.Request) {
	view, err := h.service.Overview(r.Context(), r.PathValue("id"))
	if err != nil {
		h.apiFailure(w, err)
		return
	}

	h.json(w, http.StatusOK, view)
}

func (h *Handler) apiAgent(w http.ResponseWriter, r *http.Request) {
	status, err := h.service.AgentStatus(r.Context(), r.PathValue("id"), r.PathValue("agentID"))
	if err != nil {
		h.apiFailure(w, err)
		return
	}

	h.json(w, http.StatusOK, status)
}

func (h *Handler) apiGrantOrganization(w http.ResponseWriter, r *http.Request) {
	h.apiOrganizationWrite(w, r, true)
}

func (h *Handler) apiRevokeOrganization(w http.ResponseWriter, r *http.Request) {
	h.apiOrganizationWrite(w, r, false)
}

func (h *Handler) apiOrganizationWrite(w http.ResponseWriter, r *http.Request, grant bool) {
	var err error
	if grant {
		err = h.service.GrantOrganization(r.Context(), r.PathValue("id"))
	}
	if !grant {
		err = h.service.RevokeOrganization(r.Context(), r.PathValue("id"))
	}
	if err != nil {
		h.apiFailure(w, err)
		return
	}

	h.apiShow(w, r)
}

func (h *Handler) apiGrantAgent(w http.ResponseWriter, r *http.Request) {
	h.apiAgentWrite(w, r, true)
}

func (h *Handler) apiRevokeAgent(w http.ResponseWriter, r *http.Request) {
	h.apiAgentWrite(w, r, false)
}

func (h *Handler) apiAgentWrite(w http.ResponseWriter, r *http.Request, grant bool) {
	var err error
	if grant {
		err = h.service.GrantAgent(r.Context(), r.PathValue("id"), r.PathValue("agentID"))
	}
	if !grant {
		err = h.service.RevokeAgent(r.Context(), r.PathValue("id"), r.PathValue("agentID"))
	}
	if err != nil {
		h.apiFailure(w, err)
		return
	}

	h.apiAgent(w, r)
}

func (h *Handler) apiFailure(w http.ResponseWriter, err error) {
	status, code, _ := h.failure(err)
	h.json(w, status, errorResponse{Error: code})
}

func (h *Handler) json(w http.ResponseWriter, status int, value any) {
	var output bytes.Buffer
	if json.NewEncoder(&output).Encode(value) != nil {
		http.Error(w, "internal_error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_, _ = output.WriteTo(w)
}
