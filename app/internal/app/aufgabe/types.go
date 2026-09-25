package aufgabe

import (
	"errors"

	appagent "agentcontrolplane/app/internal/app/agent"
	apporganisation "agentcontrolplane/app/internal/app/organisation"
	appprojekt "agentcontrolplane/app/internal/app/projekt"
	domainaufgabe "agentcontrolplane/app/internal/domain/aufgabe"
	portaktivitaet "agentcontrolplane/app/internal/port/aktivitaet"
	portaufgabe "agentcontrolplane/app/internal/port/aufgabe"
)

var ErrTitleRequired = domainaufgabe.ErrTitleRequired
var ErrAssigneeRequired = domainaufgabe.ErrAssigneeRequired
var ErrInvalidPriority = domainaufgabe.ErrInvalidPriority
var ErrInvalidAssignee = errors.New("Zuständiger Agent ist ungültig")
var ErrAssigneePaused = errors.New("Zuständiger Agent ist pausiert")
var ErrInvalidProject = errors.New("Projekt ist ungültig")
var ErrNotFound = portaufgabe.ErrNotFound
var ErrAccessDenied = apporganisation.ErrAccessDenied

// CreateInput enthält die vom Betreiber wählbaren Aufgabenfelder.
type CreateInput struct {
	ProjectID   string `json:"project_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Priority    string `json:"priority"`
	AssigneeID  string `json:"assignee_id"`
	Source      string `json:"-"`
}

// Service setzt die organisationsgebundene Aufgabenanlage durch.
type Service struct {
	store      portaufgabe.Store
	projects   *appprojekt.Service
	agents     *appagent.Service
	recorder   portaktivitaet.Recorder
	transactor portaktivitaet.Transactor
}
