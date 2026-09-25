package organisationswechsel

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	apporganisation "agentcontrolplane/app/internal/app/organisation"
	appregel "agentcontrolplane/app/internal/app/organisationsregel"
)

func (h *Handler) decode(w http.ResponseWriter, r *http.Request) (appregel.Change, error) {
	var change appregel.Change
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&change); err != nil {
		return change, err
	}

	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return change, errors.New("multiple JSON values")
	}

	return change, nil
}

func (h *Handler) json(w http.ResponseWriter, status int, value any) {
	var output bytes.Buffer
	if err := json.NewEncoder(&output).Encode(value); err != nil {
		http.Error(w, "internal_error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_, _ = output.WriteTo(w)
}

func (h *Handler) apiError(w http.ResponseWriter, err error) {
	status, code := h.errorStatus(err)
	h.json(w, status, errorResponse{Error: code})
}

func (h *Handler) errorStatus(err error) (int, string) {
	if errors.Is(err, apporganisation.ErrNotFound) || errors.Is(err, appregel.ErrNotFound) {
		return http.StatusNotFound, "not_found"
	}

	if errors.Is(err, apporganisation.ErrAccessDenied) || errors.Is(err, appregel.ErrAccessDenied) {
		return http.StatusForbidden, "access_denied"
	}

	return h.ruleErrorStatus(err)
}

func (h *Handler) ruleErrorStatus(err error) (int, string) {
	if errors.Is(err, appregel.ErrRevisionConflict) {
		return http.StatusConflict, "revision_conflict"
	}

	if errors.Is(err, appregel.ErrInvalidTransition) {
		return http.StatusUnprocessableEntity, "invalid_transition"
	}

	if errors.Is(err, appregel.ErrApproverRequired) {
		return http.StatusUnprocessableEntity, "approver_required"
	}

	return h.otherRuleErrorStatus(err)
}

func (h *Handler) otherRuleErrorStatus(err error) (int, string) {
	if errors.Is(err, appregel.ErrRightsExpansion) {
		return http.StatusUnprocessableEntity, "rights_expansion"
	}

	if errors.Is(err, appregel.ErrWorkflowDSL) {
		return http.StatusUnprocessableEntity, "workflow_dsl"
	}

	if errors.Is(err, appregel.ErrRuleNotFound) {
		return http.StatusUnprocessableEntity, "rule_not_found"
	}

	if errors.Is(err, appregel.ErrInvalidOperation) {
		return http.StatusUnprocessableEntity, "invalid_operation"
	}

	return http.StatusInternalServerError, "internal_error"
}
