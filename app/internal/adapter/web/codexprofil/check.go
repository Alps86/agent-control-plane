package codexprofil

import (
	"encoding/json"
	"errors"
	"net/http"
)

func (h *Handler) apiCheck(w http.ResponseWriter, r *http.Request) {
	if !h.localWrite(r) {
		h.json(w, http.StatusForbidden, errorResponse{Error: "access_denied"})
		return
	}

	if err := h.decodeCheck(r); err != nil {
		h.json(w, http.StatusBadRequest, errorResponse{Error: "invalid_json"})
		return
	}

	if h.runner == nil {
		h.json(w, http.StatusServiceUnavailable, errorResponse{Error: "runtime_unavailable"})
		return
	}

	h.runCheck(w, r)
}

func (h *Handler) runCheck(w http.ResponseWriter, r *http.Request) {
	result, err := h.runner.Run(r.Context(), r.PathValue("id"), r.PathValue("agentID"))
	if err != nil {
		h.apiError(w, err)
		return
	}

	h.json(w, http.StatusOK, result)
}

func (h *Handler) decodeCheck(r *http.Request) error {
	body, err := h.readBody(r)
	if err != nil {
		return err
	}

	var object map[string]json.RawMessage
	if err := json.Unmarshal(body, &object); err != nil {
		return err
	}

	if object == nil || len(object) != 0 {
		return errors.New("empty JSON object required")
	}

	return nil
}
