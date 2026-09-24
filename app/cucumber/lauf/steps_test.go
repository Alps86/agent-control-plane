package lauf

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"agentcontrolplane/app/internal/adapter/sqlite"
	applauf "agentcontrolplane/app/internal/app/lauf"
	apprechte "agentcontrolplane/app/internal/app/rechte"
	domainlauf "agentcontrolplane/app/internal/domain/lauf"
	domainrechte "agentcontrolplane/app/internal/domain/rechte"
	"github.com/cucumber/godog"
)

// NewSuite erzeugt einen isolierten fachlichen Szenariokontext.
func NewSuite() *Suite {
	return &Suite{}
}

func (s *Suite) InitializeScenario(sc *godog.ScenarioContext) {
	sc.Before(s.before)
	sc.After(s.after)
	s.registerSetup(sc)
	s.registerActions(sc)
	s.registerAssertions(sc)
}

func (s *Suite) registerSetup(sc *godog.ScenarioContext) {
	sc.Step(`^ein neuer lokaler Datenbestand ist für die Lauf-Anwendung konfiguriert$`, s.configure)
	sc.Step(`^die Lauf-Anwendung kennt die Aufgabe "([^"]*)" in Organisation "([^"]*)"$`, s.task)
	sc.Step(`^die Rechtegrenze erlaubt dem Betreiber "([^"]*)" den Zugriff auf "([^"]*)"$`, s.allow)
	sc.Step(`^die Rechtegrenze verweigert dem Betreiber "([^"]*)" den Zugriff auf "([^"]*)"$`, s.deny)
	sc.Step(`^der Betreiber "([^"]*)" hat einen Lauf für "([^"]*)" reserviert$`, s.preReserved)
}

func (s *Suite) registerActions(sc *godog.ScenarioContext) {
	sc.Step(`^der Betreiber "([^"]*)" über die öffentliche Lauf-Anwendungsgrenze gleichzeitig zweimal einen Start für "([^"]*)" anfordert$`, s.concurrent)
	sc.Step(`^der Betreiber "([^"]*)" über die öffentliche Lauf-Anwendungsgrenze einen Start für "([^"]*)" anfordert$`, s.reserve)
	sc.Step(`^der Betreiber "([^"]*)" erneut einen Start für "([^"]*)" anfordert$`, s.reserve)
	sc.Step(`^ich die Lauf-Anwendung mit demselben lokalen Datenbestand neu starte$`, s.restart)
	sc.Step(`^die Lauf-Anwendung vor einem Adapterstart einen Vorbereitungsfehler für diesen Lauf meldet$`, s.failPreparation)
	sc.Step(`^die Lauf-Anwendung diesen Lauf ohne tatsächlich gestarteten Adapter als "Erfolgreich" markieren soll$`, s.completeWithoutEvidence)
}

func (s *Suite) registerAssertions(sc *godog.ScenarioContext) {
	sc.Step(`^wird genau eine der beiden Startanforderungen als Reservierung angenommen$`, s.oneAccepted)
	sc.Step(`^die andere Startanforderung meldet einen bereits reservierten Lauf$`, s.otherExisting)
	sc.Step(`^die öffentliche Laufansicht zeigt genau einen aktiven Lauf für "([^"]*)"$`, s.oneActive)
	sc.Step(`^die öffentliche Laufansicht zeigt weiterhin genau einen aktiven Lauf für "([^"]*)"$`, s.oneActive)
	sc.Step(`^kein Lauf ist als erfolgreich abgeschlossen gekennzeichnet$`, s.noSuccess)
	sc.Step(`^erhält er eine Reservierung mit einer Laufkennung und dem Zustand "Reserviert"$`, s.reservationReceived)
	sc.Step(`^sieht der Betreiber "([^"]*)" dieselbe Laufkennung im Zustand "Reserviert"$`, s.sameReserved)
	sc.Step(`^meldet die Lauf-Anwendung den bereits reservierten Lauf$`, s.existingRun)
	sc.Step(`^zeigt die öffentliche Laufansicht denselben Lauf im Zustand "Fehlgeschlagen" mit dem Vorbereitungsfehler$`, s.failedVisible)
	sc.Step(`^weist die Lauf-Anwendung den Zustandswechsel zurück$`, s.transitionRejected)
	sc.Step(`^die öffentliche Laufansicht zeigt den Lauf weiterhin im Zustand "Reserviert"$`, s.stillReserved)
	sc.Step(`^zeigt die öffentliche Laufansicht den Lauf weiterhin im Zustand "Reserviert"$`, s.stillReserved)
	sc.Step(`^verweigert die Lauf-Anwendung den Start ohne Aufgabendaten preiszugeben$`, s.deniedStart)
	sc.Step(`^die öffentliche Laufansicht für "([^"]*)" zeigt keinen Lauf für "([^"]*)"$`, s.noRun)
}

func (s *Suite) before(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
	s.directory = &Directory{organizations: map[string][]string{}, resources: map[string]domainrechte.Resource{}}
	s.database, s.service, s.path = nil, nil, ""
	s.run, s.lastRun, s.parallel = domainlauf.Run{}, domainlauf.Run{}, nil
	s.lastError, s.accepted, s.actorID, s.taskID = nil, false, "", ""
	return ctx, nil
}

func (s *Suite) after(ctx context.Context, _ *godog.Scenario, _ error) (context.Context, error) {
	if s.database != nil {
		if err := s.database.Close(); err != nil {
			return ctx, err
		}
	}

	if s.path != "" {
		return ctx, os.RemoveAll(filepath.Dir(s.path))
	}

	return ctx, nil
}

func (d *Directory) Actors() []domainrechte.Actor {
	return d.actors
}

func (d *Directory) OrganizationIDs(id string) []string {
	return d.organizations[id]
}

func (d *Directory) Resource(id string) (domainrechte.Resource, bool) {
	resource, found := d.resources[id]
	return resource, found
}

func (d *Directory) Assigned(_, _ string) bool {
	return false
}

func (s *Suite) configure() error {
	directory, err := os.MkdirTemp("", "lauf-godog-*")
	if err != nil {
		return err
	}

	s.path = filepath.Join(directory, "lauf.sqlite")
	return s.open()
}

func (s *Suite) open() error {
	database, err := sqlite.OpenWithMigrations(context.Background(), s.path, sqlite.RunMigration(2))
	if err != nil {
		return err
	}

	s.database = database
	s.service = applauf.NewService(database, apprechte.NewReader(s.directory))
	return nil
}

func (s *Suite) task(id, organization string) error {
	s.taskID = id
	s.directory.resources[id] = domainrechte.NewResource(id, id, organization)
	return nil
}

func (s *Suite) allow(actorID, taskID string) error {
	resource, found := s.directory.Resource(taskID)
	if !found {
		return fmt.Errorf("Aufgabe %q fehlt", taskID)
	}

	s.actorID = actorID
	s.directory.actors = []domainrechte.Actor{domainrechte.NewActor(actorID, domainrechte.Operator)}
	s.directory.organizations[actorID] = []string{resource.OrganizationID}
	return nil
}

func (s *Suite) deny(actorID, taskID string) error {
	if _, found := s.directory.Resource(taskID); !found {
		return fmt.Errorf("Aufgabe %q fehlt", taskID)
	}

	s.actorID = actorID
	s.directory.actors = []domainrechte.Actor{domainrechte.NewActor(actorID, domainrechte.Operator)}
	s.directory.organizations[actorID] = []string{"org-fremd"}
	return nil
}

func (s *Suite) reserve(actorID, taskID string) error {
	if actorID != s.actorID {
		return fmt.Errorf("unbekannter Betreiber %q", actorID)
	}

	s.lastRun, s.accepted, s.lastError = s.service.Reserve(context.Background(), taskID)
	if s.accepted {
		s.run = s.lastRun
	}

	return nil
}

func (s *Suite) preReserved(actorID, taskID string) error {
	if err := s.reserve(actorID, taskID); err != nil {
		return err
	}

	return s.reservationReceived()
}

func (s *Suite) concurrent(actorID, taskID string) error {
	if actorID != s.actorID {
		return fmt.Errorf("unbekannter Betreiber %q", actorID)
	}

	results := make(chan Reservation, 2)
	for range 2 {
		go s.callReserve(taskID, results)
	}

	s.parallel = []Reservation{<-results, <-results}
	return nil
}

func (s *Suite) callReserve(taskID string, results chan<- Reservation) {
	run, accepted, err := s.service.Reserve(context.Background(), taskID)
	results <- Reservation{run: run, accepted: accepted, err: err}
}

func (s *Suite) restart() error {
	if err := s.database.Close(); err != nil {
		return err
	}

	s.database = nil
	return s.open()
}

func (s *Suite) failPreparation() error {
	s.lastRun, s.lastError = s.service.FailPreparation(context.Background(), s.taskID, s.run.ID, "Vorbereitung fehlgeschlagen")
	return nil
}

func (s *Suite) completeWithoutEvidence() error {
	s.lastRun, s.lastError = s.service.Complete(context.Background(), s.taskID, s.run.ID)
	return nil
}

func (s *Suite) oneAccepted() error {
	if len(s.parallel) != 2 {
		return fmt.Errorf("erwartet zwei Startantworten, erhalten %d", len(s.parallel))
	}

	count, err := s.acceptedCount()
	if err != nil {
		return err
	}

	if count != 1 {
		return fmt.Errorf("angenommene Reservierungen: %d statt 1", count)
	}

	return nil
}

func (s *Suite) acceptedCount() (int, error) {
	count := 0
	for _, result := range s.parallel {
		if result.err != nil {
			return 0, result.err
		}

		if result.accepted {
			count++
			s.run = result.run
		}
	}

	return count, nil
}

func (s *Suite) otherExisting() error {
	for _, result := range s.parallel {
		if !result.accepted && result.err == nil && result.run.ID == s.run.ID {
			return nil
		}
	}

	return fmt.Errorf("zweite Anfrage meldet nicht denselben Lauf")
}

func (s *Suite) oneActive(taskID string) error {
	runs, err := s.service.List(context.Background(), taskID)
	if err != nil {
		return err
	}

	if len(runs) != 1 || runs[0].ID != s.run.ID || runs[0].Status != domainlauf.Reserviert {
		return fmt.Errorf("erwartet genau einen aktiven Lauf, erhalten %+v", runs)
	}

	return nil
}

func (s *Suite) noSuccess() error {
	runs, err := s.service.List(context.Background(), s.taskID)
	if err != nil {
		return err
	}

	for _, run := range runs {
		if run.Status != domainlauf.Reserviert && run.Status != domainlauf.Fehlgeschlagen {
			return fmt.Errorf("unbelegter Erfolgsstatus: %+v", run)
		}
	}

	return nil
}

func (s *Suite) reservationReceived() error {
	if s.lastError != nil {
		return s.lastError
	}

	if !s.accepted || s.run.ID == "" || s.run.Status != domainlauf.Reserviert {
		return fmt.Errorf("ungültige Reservierung: %+v, angenommen=%t", s.run, s.accepted)
	}

	return nil
}

func (s *Suite) sameReserved(actorID string) error {
	if actorID != s.actorID {
		return fmt.Errorf("unbekannter Betreiber %q", actorID)
	}

	run, err := s.service.Lookup(context.Background(), s.taskID, s.run.ID)
	if err != nil {
		return err
	}

	if run.ID != s.run.ID || run.Status != domainlauf.Reserviert {
		return fmt.Errorf("Lauf nach Neustart: %+v", run)
	}

	return nil
}

func (s *Suite) existingRun() error {
	if s.lastError != nil {
		return s.lastError
	}

	if s.accepted || s.lastRun.ID != s.run.ID {
		return fmt.Errorf("erwartet bestehenden Lauf, erhalten %+v, angenommen=%t", s.lastRun, s.accepted)
	}

	return nil
}

func (s *Suite) failedVisible() error {
	run, err := s.service.Lookup(context.Background(), s.taskID, s.run.ID)
	if err != nil {
		return err
	}

	if run.ID != s.run.ID || run.Status != domainlauf.Fehlgeschlagen || run.FailureReason != "Vorbereitung fehlgeschlagen" {
		return fmt.Errorf("Vorbereitungsfehler nicht sichtbar: %+v", run)
	}

	return nil
}

func (s *Suite) transitionRejected() error {
	if !errors.Is(s.lastError, applauf.ErrAdapterEvidenceRequired) {
		return fmt.Errorf("erwartet Ablehnung ohne Adapterbeleg, erhalten %v", s.lastError)
	}

	return nil
}

func (s *Suite) stillReserved() error {
	run, err := s.service.Lookup(context.Background(), s.taskID, s.run.ID)
	if err != nil {
		return err
	}

	if run.ID != s.run.ID || run.Status != domainlauf.Reserviert {
		return fmt.Errorf("Reservierung nach Ablehnung geändert: %+v", run)
	}

	return nil
}

func (s *Suite) deniedStart() error {
	if !errors.Is(s.lastError, applauf.ErrUnavailable) || s.accepted || s.lastRun.ID != "" {
		return fmt.Errorf("unzulässige Auskunft: %+v, angenommen=%t, Fehler=%v", s.lastRun, s.accepted, s.lastError)
	}

	return nil
}

func (s *Suite) noRun(actorID, taskID string) error {
	if err := s.allow(actorID, taskID); err != nil {
		return err
	}

	runs, err := s.service.List(context.Background(), taskID)
	if err != nil {
		return err
	}

	if len(runs) != 0 {
		return fmt.Errorf("unerwartete Läufe: %+v", runs)
	}

	return nil
}
