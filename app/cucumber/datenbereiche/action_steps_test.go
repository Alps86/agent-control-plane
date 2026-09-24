package datenbereiche

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	appdatenbereich "agentcontrolplane/app/internal/app/datenbereich"
	"agentcontrolplane/app/internal/domain/rechte"
	"github.com/cucumber/godog"
)

func (s *Suite) registerActionSteps(sc *godog.ScenarioContext) {
	sc.Step(`^"([^"]+)" über die öffentliche App-Fassade die Daten von "([^"]+)" liest$`, s.readProject)
	sc.Step(`^"([^"]+)" über die öffentliche App-Fassade Daten in "([^"]+)" schreibt$`, s.writeProject)
	sc.Step(`^"([^"]+)" über die öffentliche App-Fassade die Daten eines unbekannten Projekts liest$`, s.readUnknown)
	sc.Step(`^wird der Lesezugriff ohne Projektdaten verweigert$`, s.readDenied)
	sc.Step(`^wird der Schreibzugriff ohne Datenänderung verweigert$`, s.writeDenied)
	sc.Step(`^erhält "([^"]+)" ausschließlich die Daten von "([^"]+)"$`, s.readAllowed)
	sc.Step(`^ist die Änderung beim Lesen von "([^"]+)" sichtbar$`, s.writeVisible)
	sc.Step(`^ist die Änderung beim Lesen von "([^"]+)" durch "([^"]+)" sichtbar$`, s.writeVisibleForAgent)
	sc.Step(`^die Antwort enthält weder Kennung noch Inhalt von "([^"]+)"$`, s.noForeignContent)
	sc.Step(`^sind Fehlerklasse und Projektausgabe gleich der vorherigen Verweigerung$`, s.sameOpaqueDenial)
	sc.Step(`^"([^"]+)" kann die Daten von "([^"]+)" weder lesen noch schreiben$`, s.neitherActionAllowed)
	sc.Step(`^ein Betreiberakteur mit "Miras" Kennung über die öffentliche App-Fassade "([^"]+)" lesen möchte$`, s.operatorWithAgentID)
	sc.Step(`^bleibt die serverseitig ermittelte Identität für den Projekt-Datenpfad maßgeblich$`, s.operatorCannotActAsAgent)
	sc.Step(`^der Aufruf erhält keine Daten von "([^"]+)"$`, s.noProjectData)
	sc.Step(`^der nächste Lesezugriff von "([^"]+)" auf "([^"]+)" wird ohne Projektdaten verweigert$`, s.nextReadDenied)
	s.registerAuditSteps(sc)
}

func (s *Suite) registerAuditSteps(sc *godog.ScenarioContext) {
	sc.Step(`^die Verweigerung ist mit Agent, Projekt, Aktion und Zeitpunkt auditiert$`, s.auditLastDenial)
	sc.Step(`^beide Verweigerungen sind getrennt auditiert$`, s.twoDenials)
	sc.Step(`^beide Verweigerungen sind ohne fremde Projektkennung auditiert$`, s.twoForeignDenials)
	sc.Step(`^die Verweigerung bleibt nach einem weiteren Neustart im Audit sichtbar$`, s.auditAfterRestart)
	sc.Step(`^"([^"]+)" bleibt ohne Codex-CLI-Durchsetzungsnachweis nicht startbereit$`, s.codexNotReady)
	sc.Step(`^ich den Server mit derselben SQLite-Datenbank neu starte$`, s.restartServer)
}

func (s *Suite) agentIdentity(name string) error {
	if s.agents[name] == "" {
		return fmt.Errorf("Agent %s fehlt", name)
	}

	s.agentActor = rechte.NewActor(s.agents[name], rechte.Agent)
	s.activeAgent = name
	return s.openDataFacade()
}

func (s *Suite) readProject(agent, project string) error {
	if err := s.checkAgent(agent); err != nil {
		return err
	}

	if err := s.beforeProjectAction(project, "read_project"); err != nil {
		return err
	}

	item, err := s.dataService.ReadProject(context.Background(), s.organizations["Nordstern"], s.projects[project])
	s.captureProject(item.ID, item.OrganizationID, item.Name, item.Description, err)
	return nil
}

func (s *Suite) writeProject(agent, project string) error {
	if err := s.checkAgent(agent); err != nil {
		return err
	}
	if err := s.beforeProjectAction(project, "write_project_description"); err != nil {
		return err
	}

	change := "Geändert durch " + agent + " in " + project
	item, err := s.dataService.UpdateProjectDescription(context.Background(), s.organizations["Nordstern"], s.projects[project], change)
	s.captureProject(item.ID, item.OrganizationID, item.Name, item.Description, err)
	return nil
}

func (s *Suite) checkAgent(name string) error {
	if s.agentActor.ID != s.agents[name] || s.agentActor.Kind != rechte.Agent {
		return fmt.Errorf("serverseitige Agentenidentität stimmt nicht mit %s überein", name)
	}

	if s.dataService == nil {
		return s.openDataFacade()
	}

	return nil
}

func (s *Suite) beforeProjectAction(project, action string) error {
	s.previousError = s.lastError
	s.lastProject, s.lastAction = project, action
	denials, err := s.dataService.ListDenials(context.Background(), s.organizations["Nordstern"], s.agents[s.activeAgent])
	if err != nil {
		return err
	}

	s.denialCount = len(denials)
	s.beforeDetail, _ = s.operatorProjectDescription(project)
	return nil
}

func (s *Suite) captureProject(id, orgID, name, detail string, err error) {
	s.lastID, s.lastOrgID, s.lastName, s.lastDetail = id, orgID, name, detail
	s.lastError = err
}

func (s *Suite) operatorProjectDescription(project string) (string, error) {
	projects, err := s.dataService.Projects(context.Background(), s.organizations["Nordstern"], s.agents[s.activeAgent])
	if err != nil {
		return "", err
	}

	for _, item := range projects {
		if item.ID == s.projects[project] {
			return item.Description, nil
		}
	}

	return "", nil
}

func (s *Suite) readUnknown(agent string) error {
	s.projects["<unbekannt>"] = "unbekanntes-projekt"
	return s.readProject(agent, "<unbekannt>")
}

func (s *Suite) readDenied() error {
	if !errors.Is(s.lastError, appdatenbereich.ErrNotFound) || s.lastID != "" || s.lastOrgID != "" || s.lastName != "" || s.lastDetail != "" {
		return fmt.Errorf("Leseverweigerung enthält Projektdaten: %v %q %q", s.lastError, s.lastName, s.lastDetail)
	}

	return nil
}

func (s *Suite) writeDenied() error {
	if err := s.readDenied(); err != nil {
		return err
	}

	detail, err := s.operatorProjectDescription(s.lastProject)
	if err != nil || detail != s.beforeDetail {
		return fmt.Errorf("verweigerter Schreibzugriff änderte Projektdaten: %q -> %q: %v", s.beforeDetail, detail, err)
	}

	return nil
}

func (s *Suite) readAllowed(agent, project string) error {
	if err := s.checkAgent(agent); err != nil {
		return err
	}
	if s.lastError != nil || s.lastID != s.projects[project] || s.lastOrgID != s.organizations["Nordstern"] || s.lastName != project {
		return fmt.Errorf("freigegebenes Projekt fehlt: %v %s %s", s.lastError, s.lastID, s.lastName)
	}

	return nil
}

func (s *Suite) writeVisible(project string) error {
	return s.writeVisibleForAgent(project, s.activeAgent)
}

func (s *Suite) writeVisibleForAgent(project, agent string) error {
	if s.lastError != nil {
		return s.lastError
	}

	if err := s.readProject(agent, project); err != nil {
		return err
	}
	if s.lastDetail != "Geändert durch "+agent+" in "+project {
		return fmt.Errorf("Schreibänderung nicht lesbar: %q", s.lastDetail)
	}

	return nil
}

func (s *Suite) noForeignContent(project string) error {
	if s.lastID == s.projects[project] || strings.Contains(s.lastDetail, project) || strings.Contains(s.lastName, project) {
		return fmt.Errorf("fremdes Projekt offengelegt: %s %s %s", s.lastID, s.lastName, s.lastDetail)
	}

	return nil
}

func (s *Suite) sameOpaqueDenial() error {
	if !errors.Is(s.previousError, appdatenbereich.ErrNotFound) {
		return fmt.Errorf("vorherige Fehlerklasse: %v", s.previousError)
	}

	return s.readDenied()
}

func (s *Suite) neitherActionAllowed(agent, project string) error {
	if err := s.readProject(agent, project); err != nil {
		return err
	}
	if err := s.readDenied(); err != nil {
		return err
	}
	if err := s.writeProject(agent, project); err != nil {
		return err
	}

	return s.writeDenied()
}

func (s *Suite) operatorWithAgentID(project string) error {
	identity := FixedIdentity{Actor: rechte.NewActor(s.agents["Mira"], rechte.Operator)}
	service := appdatenbereich.NewService(s.dataStore, FixedIdentity{Actor: rechte.NewActor("local-operator", rechte.Operator)}, identity)
	item, err := service.ReadProject(context.Background(), s.organizations["Nordstern"], s.projects[project])
	s.captureProject(item.ID, item.OrganizationID, item.Name, item.Description, err)
	return nil
}

func (s *Suite) operatorCannotActAsAgent() error {
	if !errors.Is(s.lastError, appdatenbereich.ErrAccessDenied) {
		return fmt.Errorf("Betreiberakteur als Agent akzeptiert: %v", s.lastError)
	}

	return nil
}

func (s *Suite) noProjectData(project string) error {
	if s.lastID != "" || s.lastOrgID != "" || s.lastName != "" || s.lastDetail != "" {
		return fmt.Errorf("Projektdaten von %s offengelegt", project)
	}

	return nil
}

func (s *Suite) nextReadDenied(agent, project string) error {
	if err := s.readProject(agent, project); err != nil {
		return err
	}

	return s.readDenied()
}

func (s *Suite) auditLastDenial() error {
	denials, err := s.dataService.ListDenials(context.Background(), s.organizations["Nordstern"], s.agents[s.activeAgent])
	if err != nil {
		return err
	}
	if len(denials) != s.denialCount+1 {
		return fmt.Errorf("Audit: %d statt %d Verweigerungen", len(denials), s.denialCount+1)
	}

	last := denials[len(denials)-1]
	if last.AgentID != s.agents[s.activeAgent] || string(last.Action) != s.lastAction || last.ProjectID != s.projects[s.lastProject] || last.OccurredAt == "" {
		return fmt.Errorf("unvollständige Auditspur: %+v", last)
	}

	return nil
}

func (s *Suite) twoDenials() error {
	denials, err := s.dataService.ListDenials(context.Background(), s.organizations["Nordstern"], s.agents[s.activeAgent])
	if err != nil {
		return err
	}

	if len(denials) != 2 || denials[0].ID == denials[1].ID || denials[1].ProjectID != "" {
		return fmt.Errorf("zwei redigierte Auditverweigerungen fehlen: %+v", denials)
	}

	return nil
}

func (s *Suite) twoForeignDenials() error {
	denials, err := s.dataService.ListDenials(context.Background(), s.organizations["Nordstern"], s.agents[s.activeAgent])
	if err != nil {
		return err
	}

	if len(denials) != 2 || denials[0].ProjectID != "" || denials[1].ProjectID != "" || denials[0].ID == denials[1].ID {
		return fmt.Errorf("fremde Kennung in Auditspur oder fehlende Verweigerung: %+v", denials)
	}

	return nil
}

func (s *Suite) auditAfterRestart() error {
	if err := s.restartServer(); err != nil {
		return err
	}

	return s.auditLastDenial()
}

func (s *Suite) codexNotReady(agent string) error {
	if err := s.request("GET", s.agentListPath("Nordstern")+"/"+s.agents[agent], "", ""); err != nil {
		return err
	}

	var profile Agent
	if err := json.Unmarshal(s.response.Body, &profile); err != nil {
		return err
	}
	if profile.Readiness.Ready || profile.Readiness.Code != "codex_cli_reference_missing" && profile.Readiness.Code != "codex_cli_enforcement_unverified" {
		return fmt.Errorf("Codex CLI ohne Durchsetzung fälschlich bereit: %s", s.response.Body)
	}

	return nil
}
