package ziel

import (
	"errors"
	"net/http"

	appziel "agentcontrolplane/app/internal/app/ziel"
	domainziel "agentcontrolplane/app/internal/domain/ziel"
)

func (h *Handler) treePageData(r *http.Request) (map[string]any, error) {
	organization, err := h.organizations.Get(r.Context(), r.PathValue("id"))
	if err != nil {
		return nil, err
	}

	goals, paths, err := h.treeData(r)
	if err != nil {
		return nil, err
	}

	data := h.pageData(organization, "list")
	view := data["View"].(map[string]any)
	view["Tree"] = h.goalNodes(goals, "")
	view["ProjectPaths"] = paths
	return data, nil
}

func (h *Handler) goalNodes(goals []domainziel.Goal, parent string) []goalNode {
	nodes := []goalNode{}
	for _, goal := range goals {
		if h.parentID(goal) != parent {
			continue
		}

		nodes = append(nodes, goalNode{ID: goal.ID, OrganizationID: goal.OrganizationID, Name: goal.Name, Status: goal.Status, StatusLabel: h.statusLabel(goal.Status), Children: h.goalNodes(goals, goal.ID)})
	}
	return nodes
}

func (h *Handler) parentID(goal domainziel.Goal) string {
	if goal.ParentGoalID == nil {
		return ""
	}

	return *goal.ParentGoalID
}

func (h *Handler) statusLabel(status string) string {
	labels := map[string]string{"planned": "Geplant", "active": "Aktiv", "achieved": "Erreicht", "cancelled": "Abgebrochen"}
	return labels[status]
}

func (h *Handler) pageCreateChild(w http.ResponseWriter, r *http.Request) {
	if !h.localWrite(r) {
		h.pageFailure(w, r, http.StatusForbidden, "Zugriff verweigert")
		return
	}

	if !h.parseActionForm(w, r) {
		return
	}

	_, err := h.goals.CreateChild(r.Context(), r.PathValue("id"), r.PathValue("zielID"), r.PostFormValue("name"))
	if err != nil {
		h.pageActionError(w, r, err)
		return
	}

	h.redirectTree(w, r)
}

func (h *Handler) pageStatus(w http.ResponseWriter, r *http.Request) {
	if !h.localWrite(r) {
		h.pageFailure(w, r, http.StatusForbidden, "Zugriff verweigert")
		return
	}

	if !h.parseActionForm(w, r) {
		return
	}

	_, err := h.goals.SetStatus(r.Context(), r.PathValue("id"), r.PathValue("zielID"), r.PostFormValue("status"))
	if err != nil {
		h.pageActionError(w, r, err)
		return
	}

	h.redirectTree(w, r)
}

func (h *Handler) parseActionForm(w http.ResponseWriter, r *http.Request) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	if err := r.ParseForm(); err != nil {
		h.pageFailure(w, r, http.StatusBadRequest, "Formulardaten konnten nicht gelesen werden.")
		return false
	}

	return true
}

func (h *Handler) pageActionError(w http.ResponseWriter, r *http.Request, err error) {
	message := h.actionMessage(err)
	if message == "" {
		h.pageError(w, r, err)
		return
	}

	data, dataErr := h.treePageData(r)
	if dataErr != nil {
		h.pageError(w, r, dataErr)
		return
	}

	if errors.Is(err, appziel.ErrNameRequired) {
		view := data["View"].(map[string]any)
		view["Tree"] = h.markChildError(view["Tree"].([]goalNode), r.PathValue("zielID"), r.PostFormValue("name"), message)
		h.render(w, r, http.StatusUnprocessableEntity, data)
		return
	}

	data["Errors"] = []string{message}
	h.render(w, r, http.StatusUnprocessableEntity, data)
}

func (h *Handler) markChildError(nodes []goalNode, id, name, message string) []goalNode {
	for index := range nodes {
		if nodes[index].ID == id {
			nodes[index].ChildName = name
			nodes[index].ChildError = message
		}

		nodes[index].Children = h.markChildError(nodes[index].Children, id, name, message)
	}
	return nodes
}

func (h *Handler) actionMessage(err error) string {
	if errors.Is(err, appziel.ErrNameRequired) {
		return "Bitte geben Sie einen Namen ein."
	}
	if errors.Is(err, appziel.ErrInvalidParent) {
		return "Bitte wählen Sie ein Ziel dieser Organisation aus."
	}
	if errors.Is(err, appziel.ErrInvalidStatus) {
		return "Bitte wählen Sie einen gültigen Status aus."
	}
	return ""
}

func (h *Handler) redirectTree(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/organisationen/"+r.PathValue("id")+"/ziele", http.StatusSeeOther)
}
