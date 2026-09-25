package ziel

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	domainziel "agentcontrolplane/app/internal/domain/ziel"
)

func (h *Handler) apiTree(w http.ResponseWriter, r *http.Request) {
	goals, paths, err := h.treeData(r)
	if err != nil {
		h.apiError(w, err)
		return
	}

	h.json(w, http.StatusOK, treeResponse{Goals: goals, ProjectPaths: paths})
}

func (h *Handler) treeData(r *http.Request) ([]domainziel.Goal, []domainziel.ProjectPath, error) {
	goals, err := h.goals.List(r.Context(), r.PathValue("id"))
	if err != nil {
		return nil, nil, err
	}

	paths, err := h.goals.ProjectPaths(r.Context(), r.PathValue("id"))
	if goals == nil {
		goals = []domainziel.Goal{}
	}
	if paths == nil {
		paths = []domainziel.ProjectPath{}
	}
	return goals, paths, err
}

func (h *Handler) apiCreateChild(w http.ResponseWriter, r *http.Request) {
	if !h.localWrite(r) {
		h.json(w, http.StatusForbidden, errorResponse{Error: "access_denied"})
		return
	}

	request, err := h.decode(r)
	if err != nil {
		h.json(w, http.StatusBadRequest, errorResponse{Error: "invalid_json"})
		return
	}

	goal, err := h.goals.CreateChild(r.Context(), r.PathValue("id"), r.PathValue("zielID"), request.Name)
	if err != nil {
		h.apiError(w, err)
		return
	}

	w.Header().Set("Location", "/api/organisationen/"+goal.OrganizationID+"/zielbaum")
	h.json(w, http.StatusCreated, goal)
}

func (h *Handler) apiStatus(w http.ResponseWriter, r *http.Request) {
	if !h.localWrite(r) {
		h.json(w, http.StatusForbidden, errorResponse{Error: "access_denied"})
		return
	}

	request, err := h.decodeStatus(r)
	if err != nil {
		h.json(w, http.StatusBadRequest, errorResponse{Error: "invalid_json"})
		return
	}

	goal, err := h.goals.SetStatus(r.Context(), r.PathValue("id"), r.PathValue("zielID"), request.Status)
	if err != nil {
		h.apiError(w, err)
		return
	}

	h.json(w, http.StatusOK, goal)
}

func (h *Handler) decodeStatus(r *http.Request) (statusRequest, error) {
	var request statusRequest
	decoder := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		return request, err
	}

	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return request, errors.New("multiple JSON values")
	}

	return request, nil
}
