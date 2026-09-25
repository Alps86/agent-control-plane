package projektarchiv

import domainprojekt "agentcontrolplane/app/internal/domain/projekt"

// Status beschreibt, ob ein Projekt neue Arbeit annehmen darf.
type Status string

const (
	StatusActive   Status = "active"
	StatusArchived Status = "archived"
)

// ArchivedProject erhält die fachlichen Projektdaten unverändert.
type ArchivedProject struct {
	domainprojekt.Project
	ArchivedAt string `json:"archived_at"`
}
