package ziel

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	appziel "agentcontrolplane/app/internal/app/ziel"
)

func (h *Handler) apiMove(w http.ResponseWriter, r *http.Request) {
	if !h.localWrite(r) {
		h.json(w, http.StatusForbidden, errorResponse{Error: "access_denied"})
		return
	}

	request, err := h.decodeMove(r)
	if err != nil {
		h.json(w, http.StatusBadRequest, errorResponse{Error: "invalid_json"})
		return
	}

	goal, err := h.goals.Move(r.Context(), r.PathValue("id"), r.PathValue("zielID"), *request.ParentGoalID)
	if err != nil {
		h.apiError(w, err)
		return
	}

	h.json(w, http.StatusOK, goal)
}

func (h *Handler) decodeMove(r *http.Request) (moveRequest, error) {
	var request moveRequest
	decoder := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		return request, err
	}

	var extra any
	if err := decoder.Decode(&extra); err != io.EOF || request.ParentGoalID == nil {
		return request, errors.New("ungültige Zielkante")
	}

	return request, nil
}

func (h *Handler) pageMove(w http.ResponseWriter, r *http.Request) {
	if !h.localWrite(r) {
		h.pageFailure(w, r, http.StatusForbidden, "Zugriff verweigert")
		return
	}

	if !h.parseActionForm(w, r) {
		return
	}

	values := r.PostForm["parent_goal_id"]
	if len(values) != 1 {
		h.pageMoveError(w, r, appziel.ErrInvalidParent)
		return
	}

	_, err := h.goals.Move(r.Context(), r.PathValue("id"), r.PathValue("zielID"), values[0])
	if err != nil {
		h.pageMoveError(w, r, err)
		return
	}

	h.redirectTree(w, r)
}

func (h *Handler) pageMoveError(w http.ResponseWriter, r *http.Request, err error) {
	if !errors.Is(err, appziel.ErrInvalidParent) && !errors.Is(err, appziel.ErrGoalCycle) {
		h.pageError(w, r, err)
		return
	}

	data, dataErr := h.treePageData(r)
	if dataErr != nil {
		h.pageError(w, r, dataErr)
		return
	}

	view := data["View"].(map[string]any)
	view["Tree"] = h.markParentError(view["Tree"].([]goalNode), r.PathValue("zielID"), r.PostFormValue("parent_goal_id"), h.actionMessage(err))
	h.render(w, r, http.StatusUnprocessableEntity, data)
}

func (h *Handler) markParentError(nodes []goalNode, id, attempted, message string) []goalNode {
	for index := range nodes {
		if nodes[index].ID == id {
			nodes[index].ParentError = message
			h.selectParentOption(&nodes[index], attempted)
		}

		nodes[index].Children = h.markParentError(nodes[index].Children, id, attempted, message)
	}

	return nodes
}

func (h *Handler) selectParentOption(node *goalNode, attempted string) {
	for index := range node.ParentOptions {
		node.ParentOptions[index].Selected = node.ParentOptions[index].ID == attempted
	}
}
