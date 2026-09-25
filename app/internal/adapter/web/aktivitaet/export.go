package aktivitaet

import (
	"bytes"
	"encoding/csv"
	"net/http"
	"time"

	domainaktivitaet "agentcontrolplane/app/internal/domain/aktivitaet"
)

func (h *Handler) apiExport(w http.ResponseWriter, r *http.Request) {
	events, err := h.events(r)
	if err != nil {
		h.apiError(w, err)
		return
	}

	output, err := h.csvEvents(events)
	if err != nil {
		h.apiError(w, err)
		return
	}

	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="aktivitaet.csv"`)
	w.WriteHeader(http.StatusOK)
	_, _ = output.WriteTo(w)
}

func (h *Handler) csvEvents(events []domainaktivitaet.Event) (*bytes.Buffer, error) {
	output := &bytes.Buffer{}
	writer := csv.NewWriter(output)
	if err := writer.Write([]string{"id", "organization_id", "project_id", "task_id", "kind", "actor", "source", "occurred_at", "object_title", "deep_link", "assignee_id", "assignee_name"}); err != nil {
		return nil, err
	}

	for _, event := range events {
		if err := writer.Write(h.csvRecord(event)); err != nil {
			return nil, err
		}
	}

	writer.Flush()
	return output, writer.Error()
}

func (h *Handler) csvRecord(event domainaktivitaet.Event) []string {
	return []string{event.ID, event.OrganizationID, event.ProjectID, event.TaskID,
		event.Kind, event.Actor, event.Source, event.OccurredAt.Format(time.RFC3339Nano),
		event.ObjectTitle, event.DeepLink, event.AssigneeID, event.AssigneeName}
}
