package aktivitaet

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"agentcontrolplane/app/internal/adapter/sqlite"
	webaktivitaet "agentcontrolplane/app/internal/adapter/web/aktivitaet"
	webaufgabe "agentcontrolplane/app/internal/adapter/web/aufgabe"
	appagent "agentcontrolplane/app/internal/app/agent"
	appaktivitaet "agentcontrolplane/app/internal/app/aktivitaet"
	appaufgabe "agentcontrolplane/app/internal/app/aufgabe"
	apporganisation "agentcontrolplane/app/internal/app/organisation"
	appprojekt "agentcontrolplane/app/internal/app/projekt"
	appprojektarchiv "agentcontrolplane/app/internal/app/projektarchiv"
	"github.com/cucumber/godog"
)

func (s *Suite) registerRollback(sc *godog.ScenarioContext) {
	sc.Step(`^der zweite Ereignisschreibvorgang schlägt kontrolliert fehl$`, s.openFailureFixture)
	sc.Step(`^ich "([^"]+)" im Projekt "([^"]+)" dem Agenten "([^"]+)" über die öffentliche API zuweise und der Schreibfehler eintritt$`, s.createWithFailure)
	sc.Step(`^meldet die Aufgaben-API einen Serverfehler$`, s.writeFailed)
	sc.Step(`^die Aufgaben-API findet die betroffene Aufgabenkennung nicht$`, s.taskRolledBack)
	sc.Step(`^der Aktivitätsverlauf von "([^"]+)" enthält kein Ereignis$`, s.noEvents)
}

func (s *Suite) openFailureFixture() error {
	s.stopServer()
	db, err := sqlite.OpenApplication(context.Background(), s.database)
	if err != nil {
		return fmt.Errorf("Rollbackfixture mit kanonischer Migration öffnen: %w", err)
	}

	s.fixtureDB = db
	s.failingStore = &FailSecondStore{Underlying: db}
	mux := http.NewServeMux()
	s.fixtureServer = httptest.NewUnstartedServer(mux)
	s.fixtureRoutes(db, mux, s.fixtureServer.Listener.Addr().String())
	s.fixtureServer.Start()
	s.address = strings.TrimPrefix(s.fixtureServer.URL, "http://")
	return nil
}

func (s *Suite) fixtureRoutes(db *sqlite.Database, mux *http.ServeMux, bindAddress string) {
	identity := apporganisation.NewLocalIdentity()
	organizations := apporganisation.NewService(sqlite.NewOrganizationStore(db), identity)
	projects := appprojekt.NewService(db, organizations)
	agents := appagent.NewService(db, identity, sqlite.NewOrganizationStore(db))
	activity := appaktivitaet.NewService(s.failingStore, organizations, identity)
	tasks := appaufgabe.NewService(db, projects, agents, activity, db)
	s.fixtureMux(db, mux, bindAddress, tasks, activity, organizations, projects, agents)
}

func (s *Suite) fixtureMux(db *sqlite.Database, mux *http.ServeMux, bindAddress string, tasks *appaufgabe.Service, activity *appaktivitaet.Service,
	organizations *apporganisation.Service, projects *appprojekt.Service, agents *appagent.Service) {
	taskHandler := webaufgabe.NewHandler(tasks, projects, appprojektarchiv.NewService(db, organizations), agents, organizations, nil, bindAddress)
	activityHandler := webaktivitaet.NewHandler(activity, organizations, nil)
	mux.Handle("POST /api/organisationen/{id}/projekte/{projektID}/aufgaben", taskHandler)
	mux.Handle("GET /api/organisationen/{id}/projekte/{projektID}/aufgaben/{aufgabeID}", taskHandler)
	mux.Handle("GET /api/organisationen/{id}/aktivitaet", activityHandler)
}

func (s *Suite) createWithFailure(title, project, agent string) error {
	return s.postTask(title, project, agent, nil)
}

func (s *Suite) writeFailed() error {
	if s.status != http.StatusInternalServerError || s.failingStore.Calls != 2 {
		return fmt.Errorf("Schreibfehler nicht erreicht: HTTP %d, Calls %d: %s", s.status, s.failingStore.Calls, s.body)
	}

	return nil
}

func (s *Suite) taskRolledBack() error {
	if s.failingStore.TaskID == "" {
		return fmt.Errorf("Fehlerszenario hat keine Aufgabenkennung erfasst")
	}

	path := s.taskPath("Nordstern", "Website") + "/" + s.failingStore.TaskID
	if err := s.request("GET", path, ""); err != nil {
		return err
	}

	if s.status != http.StatusNotFound {
		return fmt.Errorf("Aufgabe nach Rollback HTTP %d: %s", s.status, s.body)
	}

	return nil
}
