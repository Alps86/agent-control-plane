package projekte

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"

	"agentcontrolplane/app/internal/adapter/sqlite"
	webprojekt "agentcontrolplane/app/internal/adapter/web/projekt"
	apporganisation "agentcontrolplane/app/internal/app/organisation"
	appprojekt "agentcontrolplane/app/internal/app/projekt"
	appziel "agentcontrolplane/app/internal/app/ziel"
	"agentcontrolplane/app/internal/domain/rechte"
	portorganisation "agentcontrolplane/app/internal/port/organisation"
	"github.com/cucumber/godog"
)

func (s *Suite) registerSecuritySteps(sc *godog.ScenarioContext) {
	sc.Step(`^antwortet der Server mit einem datenfreien HTTP-Status 404$`, s.notFound)
	sc.Step(`^antwortet der Server mit demselben datenfreien HTTP-Status 404$`, s.sameNotFound)
	sc.Step(`^der Server für einen HTTP-Aufruf keine eindeutige Betreiberidentität ermittelt$`, s.noIdentity)
	sc.Step(`^verweigert er das Lesen der Projektübersicht von "([^"]*)" ohne Projektdaten$`, s.deniedProjectList)
	sc.Step(`^er verweigert das Anlegen eines Projekts in "([^"]*)" ohne Teilerzeugung$`, s.deniedProjectCreate)
	sc.Step(`^ich als nicht zugeordnete Betreiberin die Projektübersicht von "([^"]*)" über HTTP abrufe$`, s.unassignedProjectList)
	sc.Step(`^ich als nicht zugeordnete Betreiberin das Projektdetail von "([^"]*)" über HTTP abrufe$`, s.unassignedProjectDetail)
	sc.Step(`^als nicht zugeordnete Betreiberin kann ich in "([^"]*)" kein Projekt anlegen$`, s.unassignedCannotCreate)
	sc.Step(`^ich von Origin "([^"]*)" das Projekt "([^"]*)" mit dem Ziel "([^"]*)" in "([^"]*)" über HTTP anlege$`, s.foreignOrigin)
	sc.Step(`^ich über Host "([^"]*)" das Projekt "([^"]*)" mit dem Ziel "([^"]*)" in "([^"]*)" über HTTP anlege$`, s.foreignHost)
	sc.Step(`^antwortet der Server mit dem HTTP-Status (\d+)$`, s.statusCode)
}

func (s *Suite) notFound() error {
	if err := s.assertNotFound(); err != nil {
		return err
	}

	s.unknownBody = append([]byte(nil), s.response.Body...)
	return nil
}

func (s *Suite) sameNotFound() error {
	if err := s.assertNotFound(); err != nil {
		return err
	}

	if len(s.unknownBody) == 0 || string(s.unknownBody) != string(s.response.Body) {
		return fmt.Errorf("404-Antworten verschieden: %q / %q", s.unknownBody, s.response.Body)
	}

	return nil
}

func (s *Suite) assertNotFound() error {
	if err := s.statusCode(http.StatusNotFound); err != nil {
		return err
	}

	var failure FieldError
	if err := json.Unmarshal(s.response.Body, &failure); err != nil {
		return err
	}

	if failure.Error != "not_found" || s.active != "" && (strings.Contains(string(s.response.Body), s.active) || strings.Contains(string(s.response.Body), s.organizations[s.active])) {
		return fmt.Errorf("404 mit Daten: %s", s.response.Body)
	}

	return nil
}

func (i *FixedIdentity) Actors() []rechte.Actor { return []rechte.Actor{i.Actor} }

func (s *Suite) noIdentity() error { return s.startNegative(nil) }

func (s *Suite) startNegative(identity *FixedIdentity) error {
	s.stopNegative()
	db, err := s.openNegativeDatabase()
	if err != nil {
		return err
	}

	organizations := apporganisation.NewService(sqlite.NewOrganizationStore(db), s.negativeIdentity(identity))
	goals := appziel.NewService(db, organizations)
	projects := appprojekt.NewService(db, organizations)
	server := httptest.NewUnstartedServer(nil)
	server.Config.Handler = webprojekt.NewHandler(projects, organizations, goals, nil, server.Listener.Addr().String())
	server.Start()
	s.negative = &NegativeServer{DB: db, Server: server}
	return nil
}

func (s *Suite) negativeIdentity(identity *FixedIdentity) portorganisation.Identity {
	if identity == nil {
		return nil
	}

	return identity
}

func (s *Suite) openNegativeDatabase() (*sqlite.Database, error) {
	return sqlite.OpenWithMigrations(context.Background(), s.database,
		sqlite.RunMigration(2), sqlite.OrganizationMigration(), sqlite.GoalMigration(),
		sqlite.AgentMigration(), sqlite.ProjectMigration())
}

func (s *Suite) stopNegative() {
	if s.negative == nil {
		return
	}

	s.negative.Server.Close()
	_ = s.negative.DB.Close()
	s.negative = nil
}

func (s *Suite) negativeRequest(method, path, body string) error {
	request, err := http.NewRequest(method, s.negative.Server.URL+path, strings.NewReader(body))
	if err != nil {
		return err
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Actor-ID", "local-operator")
	request.Header.Set("X-Operator-ID", "local-operator")
	response, err := s.client.Do(request)
	if err != nil {
		return err
	}

	defer response.Body.Close()
	data, err := io.ReadAll(response.Body)
	s.response = &HTTPResponse{Status: response.StatusCode, Body: data}
	return err
}

func (s *Suite) assertDenied() error {
	if err := s.statusCode(http.StatusForbidden); err != nil {
		return err
	}

	var failure FieldError
	if err := json.Unmarshal(s.response.Body, &failure); err != nil {
		return err
	}

	if failure.Error != "access_denied" || strings.Contains(string(s.response.Body), s.active) {
		return fmt.Errorf("403 mit Daten: %s", s.response.Body)
	}

	return nil
}

func (s *Suite) deniedProjectList(name string) error {
	if err := s.negativeRequest("GET", s.projectPath(name)+"?actor=local-operator", ""); err != nil {
		return err
	}

	return s.assertDenied()
}

func (s *Suite) deniedProjectCreate(name string) error {
	body := s.projectBody("Verbotenes Projekt", "Unerlaubt", s.firstGoalID(name))
	if err := s.negativeRequest("POST", s.projectPath(name)+"?actor=local-operator", body); err != nil {
		return err
	}

	if err := s.assertDenied(); err != nil {
		return err
	}

	return s.assertProjectAbsent(name, "Verbotenes Projekt")
}

func (s *Suite) unassignedProjectList(name string) error {
	if err := s.startNegative(&FixedIdentity{Actor: rechte.NewActor("other-operator", rechte.Operator)}); err != nil {
		return err
	}

	if err := s.negativeRequest("GET", "/api/organisationen/unbekannt/projekte", ""); err != nil {
		return err
	}

	s.active = name
	return s.negativeRequest("GET", s.projectPath(name), "")
}

func (s *Suite) unassignedProjectDetail(name string) error {
	return s.negativeRequest("GET", s.detailPath(s.active, name), "")
}

func (s *Suite) unassignedCannotCreate(name string) error {
	body := s.projectBody("Verbotenes Projekt", "Unerlaubt", s.firstGoalID(name))
	if err := s.negativeRequest("POST", s.projectPath(name), body); err != nil {
		return err
	}

	if err := s.assertNotFound(); err != nil {
		return err
	}

	return s.assertProjectAbsent(name, "Verbotenes Projekt")
}

func (s *Suite) assertProjectAbsent(organization, name string) error {
	if err := s.getProjects(organization); err != nil {
		return err
	}

	list, err := s.parseProjects()
	if err != nil {
		return err
	}

	for _, project := range list.Projects {
		if project.Name == name {
			return fmt.Errorf("verweigertes Projekt angelegt: %+v", project)
		}
	}

	return nil
}

func (s *Suite) firstGoalID(organization string) string {
	for key, id := range s.goals {
		if strings.HasPrefix(key, organization+"\x00") {
			return id
		}
	}

	return ""
}

func (s *Suite) projectBody(name, description, goalID string) string {
	data, _ := json.Marshal(map[string]string{"name": name, "description": description, "goal_id": goalID})
	return string(data)
}

func (s *Suite) foreignOrigin(origin, project, goal, organization string) error {
	return s.foreignWrite(project, goal, organization, origin, "")
}

func (s *Suite) foreignHost(host, project, goal, organization string) error {
	return s.foreignWrite(project, goal, organization, "", host)
}

func (s *Suite) foreignWrite(project, goal, organization, origin, host string) error {
	headers := http.Header{"Content-Type": []string{"application/json"}}
	if origin != "" {
		headers.Set("Origin", origin)
	}

	return s.request("POST", s.projectPath(organization), s.projectBody(project, "Unerlaubt", s.goals[s.goalKey(organization, goal)]), headers, host)
}
