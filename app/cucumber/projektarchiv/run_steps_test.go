package projektarchiv

import (
	"agentcontrolplane/app/internal/adapter/sqlite"
	applauf "agentcontrolplane/app/internal/app/lauf"
	apprechte "agentcontrolplane/app/internal/app/rechte"
	"agentcontrolplane/app/internal/domain/rechte"
	"context"
	"fmt"
	"strings"
)

func (d RunDirectory) Actors() []rechte.Actor { return []rechte.Actor{d.Actor} }
func (d RunDirectory) OrganizationIDs(id string) []string {
	if id != d.Actor.ID {
		return nil
	}
	return []string{d.OrganizationID}
}
func (d RunDirectory) Resource(id string) (rechte.Resource, bool) { return d.Task, id == d.Task.ID }
func (d RunDirectory) Assigned(string, string) bool               { return false }

func (s *Suite) reserveRun(title string) error {
	taskID := s.tasks[s.key(s.activeOrg, title)]
	if taskID == "" {
		return fmt.Errorf("Aufgabe %s fehlt", title)
	}
	service, database, err := s.openRunService(taskID)
	if err != nil {
		return err
	}
	defer database.Close()
	run, accepted, err := service.Reserve(context.Background(), taskID)
	if err != nil || !accepted || run.ID == "" {
		return fmt.Errorf("Lauf nicht reserviert: %+v %t %v", run, accepted, err)
	}
	s.runID = run.ID
	return nil
}

func (s *Suite) openRunService(taskID string) (*applauf.Service, *sqlite.Database, error) {
	database, err := sqlite.OpenApplication(context.Background(), s.database)
	if err != nil {
		return nil, nil, err
	}
	directory := RunDirectory{Actor: rechte.NewActor("operator", rechte.Operator), OrganizationID: s.organizations[s.activeOrg], Task: rechte.NewResource(taskID, taskID, s.organizations[s.activeOrg])}
	service := applauf.NewService(database, apprechte.NewReader(directory))
	return service, database, nil
}

func (s *Suite) failedRun(title string) error {
	if err := s.reserveRun(title); err != nil {
		return err
	}
	taskID := s.tasks[s.key(s.activeOrg, title)]
	service, database, err := s.openRunService(taskID)
	if err != nil {
		return err
	}
	defer database.Close()
	_, err = service.FailPreparation(context.Background(), taskID, s.runID, "Vorbereitung fehlgeschlagen")
	return err
}

func (s *Suite) runPreserved(title string) error {
	taskID := s.tasks[s.key(s.activeOrg, title)]
	service, database, err := s.openRunService(taskID)
	if err != nil {
		return err
	}
	defer database.Close()
	runs, err := service.List(context.Background(), taskID)
	if err != nil {
		return err
	}
	for _, run := range runs {
		if run.ID == s.runID && string(run.Status) == "Fehlgeschlagen" {
			return nil
		}
	}
	return fmt.Errorf("Lauf %s fehlt im Verlauf: %+v", s.runID, runs)
}

func (s *Suite) attemptArchivedRun(title string) error {
	taskID := s.tasks[s.key(s.activeOrg, title)]
	service, database, err := s.openRunService(taskID)
	if err != nil {
		return err
	}
	defer database.Close()
	run, accepted, err := service.Reserve(context.Background(), taskID)
	if err != nil && !strings.Contains(err.Error(), "project_archived") {
		return fmt.Errorf("unerwarteter Lauf-Fehler: %w", err)
	}

	s.runDenied = !accepted && err != nil && run.ID == ""
	return nil
}

func (s *Suite) runReservationDenied() error {
	if !s.runDenied {
		return fmt.Errorf("Laufreservierung in archiviertem Projekt angenommen")
	}

	return nil
}

func (s *Suite) noNewRun(title string) error {
	taskID := s.tasks[s.key(s.activeOrg, title)]
	service, database, err := s.openRunService(taskID)
	if err != nil {
		return err
	}
	defer database.Close()
	runs, err := service.List(context.Background(), taskID)
	if err != nil {
		return err
	}
	if len(runs) != 0 {
		return fmt.Errorf("Neuer Lauf trotz Archiv: %+v", runs)
	}

	return nil
}
