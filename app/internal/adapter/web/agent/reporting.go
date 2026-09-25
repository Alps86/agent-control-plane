package agent

import (
	"context"
	"net/http"

	appagent "agentcontrolplane/app/internal/app/agent"
)

func (h *Handler) projectReporting(r *http.Request, profile appagent.Profile, data map[string]any) error {
	if h.reporting == nil {
		return nil
	}

	parentID, parentName, err := h.reportingParent(r.Context(), r.PathValue("id"), profile.ID)
	if err != nil {
		return err
	}

	agent := data["View"].(map[string]any)["Agent"].(map[string]any)
	agent["ParentID"], agent["ParentName"] = parentID, parentName
	return nil
}

func (h *Handler) reportingParent(ctx context.Context, organizationID, agentID string) (string, string, error) {
	chart, err := h.reporting.Chart(ctx, organizationID)
	if err != nil {
		return "", "", err
	}

	parentID := ""
	for _, line := range chart.Lines {
		if line.AgentID == agentID {
			parentID = line.ParentID
			break
		}
	}

	for _, line := range chart.Lines {
		if line.AgentID == parentID {
			return parentID, line.Name, nil
		}
	}

	return parentID, "", nil
}
