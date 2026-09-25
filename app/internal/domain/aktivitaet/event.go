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

// Valid prüft Seitengröße und RFC3339-Zeitfenster vor dem Datenzugriff.
func (f Filter) Valid() bool {
	if f.Limit < 0 || f.Limit > 200 || f.Offset < 0 {
		return false
	}
	from, fromOK := f.parseTime(f.From)
	to, toOK := f.parseTime(f.To)
	return fromOK && toOK && (f.From == "" || f.To == "" || !from.After(to))
}

func (f Filter) parseTime(value string) (time.Time, bool) {
	if value == "" {
		return time.Time{}, true
	}
	at, err := time.Parse(time.RFC3339, value)
	return at, err == nil
}
