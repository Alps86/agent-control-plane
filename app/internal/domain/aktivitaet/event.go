package aktivitaet

import (
	"time"

	domainaufgabe "agentcontrolplane/app/internal/domain/aufgabe"
	"github.com/google/uuid"
)

// NewTaskEvent erstellt eine unveränderliche Momentaufnahme des Aufgabenvorgangs.
func NewTaskEvent(task domainaufgabe.Task, kind, actor, source, assigneeName string, at time.Time) Event {
	event := Event{ID: uuid.NewString(), OrganizationID: task.OrganizationID,
		ProjectID: task.ProjectID, TaskID: task.ID, Kind: kind, Actor: actor,
		Source: source, OccurredAt: at.UTC(), ObjectTitle: task.Title,
		DeepLink: "/organisationen/" + task.OrganizationID + "/projekte/" + task.ProjectID + "/aufgaben/" + task.ID}
	if kind == KindAssigned {
		event.AssigneeID, event.AssigneeName = task.AssigneeID, assigneeName
	}

	return event
}

// Valid prüft nur die verifizierbaren Felder eines fachlichen Ereignisses.
func (e Event) Valid() bool {
	if e.ID == "" || e.OrganizationID == "" || e.ProjectID == "" || e.TaskID == "" || e.Actor == "" || e.ObjectTitle == "" || e.DeepLink == "" || e.OccurredAt.IsZero() {
		return false
	}

	if e.Source != SourceAPI && e.Source != SourceBrowser {
		return false
	}

	return e.Kind == KindCreated && e.AssigneeID == "" && e.AssigneeName == "" ||
		e.Kind == KindAssigned && e.AssigneeID != "" && e.AssigneeName != ""
}
