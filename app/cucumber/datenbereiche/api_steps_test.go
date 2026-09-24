package datenbereiche

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"agentcontrolplane/app/internal/domain/rechte"

	"github.com/cucumber/godog"
)

func (s *Suite) initializeScenario(sc *godog.ScenarioContext) {
	sc.After(s.afterScenario)
	s.registerSetupSteps(sc)
	s.registerScopeSteps(sc)
	s.registerActionSteps(sc)
	s.registerBrowserSteps(sc)
}

func (s *Suite) registerSetupSteps(sc *godog.ScenarioContext) {
	sc.Step(`^ich starte den lokalen Server mit einer neuen SQLite-Datenbank als Betreiberin$`, s.freshServer)
	sc.Step(`^die Organisationen "([^"]+)" und "([^"]+)" bestehen$`, s.ensureOrganizations)
	sc.Step(`^die Organisation "([^"]+)" mit dem Projekt "([^"]+)" besteht$`, s.organizationWithProject)
	sc.Step(`^die Projekte "([^"]+)" und "([^"]+)" gehören zu "([^"]+)"$`, s.ensureTwoProjects)
	sc.Step(`^das Projekt "([^"]+)" gehört zu "([^"]+)"$`, s.ensureProject)
	sc.Step(`^"([^"]+)" ist ein Projekt der anderen Organisation "([^"]+)"$`, s.ensureProject)
	sc.Step(`^"([^"]+)" ist ein Eino-Agent von "([^"]+)"$`, s.ensureAgent)
	sc.Step(`^"([^"]+)" ist ein Eino-Agent von "([^"]+)" ohne Datenbereich$`, s.ensureAgent)
	sc.Step(`^"([^"]+)" ist ein Codex-CLI-Agent von "([^"]+)"$`, s.ensureCodexAgent)
	sc.Step(`^der öffentliche Projekt-Datenpfad ermittelt "([^"]+)" über den serverseitigen Identitätsport$`, s.agentIdentity)
}

func (s *Suite) registerScopeSteps(sc *godog.ScenarioContext) {
	sc.Step(`^ich "([^"]+)" über die öffentliche Betreiber-API "([^"]+)" zum Lesen freigebe$`, s.grantRead)
	sc.Step(`^ich "([^"]+)" über die öffentliche Betreiber-API "([^"]+)" zum Lesen und Schreiben freigebe$`, s.grantReadWrite)
	sc.Step(`^ich habe "([^"]+)" "([^"]+)" zum Lesen freigegeben$`, s.grantRead)
	sc.Step(`^ich habe "([^"]+)" "([^"]+)" zum Lesen und Schreiben freigegeben$`, s.grantReadWrite)
	sc.Step(`^ich über die öffentliche Betreiber-API "([^"]+)" und "([^"]+)" für "([^"]+)" zum Lesen freigebe$`, s.grantTwoRead)
	sc.Step(`^ich "([^"]+)" zum Lesen und "([^"]+)" ohne Lesen und Schreiben über die öffentliche Betreiber-API sende$`, s.grantWithDisabledForeign)
	sc.Step(`^ich "([^"]+)" zum Lesen und ein unbekanntes Projekt ohne Lesen und Schreiben über die öffentliche Betreiber-API sende$`, s.grantWithDisabledUnknown)
	sc.Step(`^ich "([^"]+)" über die öffentliche Betreiber-API das Schreiben in "([^"]+)" entziehe$`, s.revokeWrite)
	sc.Step(`^die Datenbereichs-API zeigt weiterhin nur "([^"]+)" zum Lesen$`, s.onlyReadScope)
	sc.Step(`^die Datenbereichs-API zeigt "([^"]+)" mit Lesen und ohne Schreiben$`, s.onlyReadScope)
	sc.Step(`^zeigt die Datenbereichs-API "([^"]+)" mit Lesen und ohne Schreiben$`, s.onlyReadScope)
	sc.Step(`^zeigt die Datenbereichs-API für "([^"]+)" "([^"]+)" mit Lesen und Schreiben$`, s.readWriteScope)
	sc.Step(`^wird die gesamte Änderung ohne fremde Projektdaten abgewiesen$`, s.atomicForeignDenied)
	sc.Step(`^ich "([^"]+)" unter "([^"]+)" über die öffentliche Datenbereichs-API abrufe$`, s.getForeignAgentScope)
	sc.Step(`^erhalte ich eine datenfreie Antwort mit HTTP-Status 404$`, s.notFoundWithoutData)
}

func (s *Suite) ensureOrganizations(first, second string) error {
	if err := s.ensureOrganization(first); err != nil {
		return err
	}

	return s.ensureOrganization(second)
}

func (s *Suite) ensureOrganization(name string) error {
	if s.organizations[name] != "" {
		return nil
	}

	body, _ := json.Marshal(map[string]string{"name": name, "description": "Testorganisation " + name})
	if err := s.request(http.MethodPost, "/api/organisationen", string(body), "application/json"); err != nil {
		return err
	}

	return s.rememberOrganization(name)
}

func (s *Suite) rememberOrganization(name string) error {
	if s.response.Status != http.StatusCreated {
		return fmt.Errorf("Organisation %s: HTTP %d: %s", name, s.response.Status, s.response.Body)
	}

	var organization Organization
	if err := json.Unmarshal(s.response.Body, &organization); err != nil {
		return err
	}

	s.organizations[name] = organization.ID
	return nil
}

func (s *Suite) ensureAgent(name, org string) error {
	return s.createAgent(name, org, "recherche", "eino")
}

func (s *Suite) ensureCodexAgent(name, org string) error {
	return s.createAgent(name, org, "codex-cli", "codex_cli")
}

func (s *Suite) createAgent(name, org, template, kind string) error {
	if s.agents[name] != "" {
		return nil
	}

	if err := s.ensureOrganization(org); err != nil {
		return err
	}

	body, _ := json.Marshal(map[string]string{"name": name, "template_id": template, "execution_kind": kind})
	if err := s.request(http.MethodPost, s.agentListPath(org), string(body), "application/json"); err != nil {
		return err
	}

	return s.rememberAgent(name, org, kind)
}

func (s *Suite) rememberAgent(name, org, kind string) error {
	if s.response.Status != http.StatusCreated {
		return fmt.Errorf("Agent %s: HTTP %d: %s", name, s.response.Status, s.response.Body)
	}

	var agent Agent
	if err := json.Unmarshal(s.response.Body, &agent); err != nil {
		return err
	}
	if agent.Name != name || agent.OrganizationID != s.organizations[org] || agent.ExecutionKind != kind {
		return fmt.Errorf("falsches Agentenprofil: %+v", agent)
	}

	s.agents[name] = agent.ID
	s.agentActor = rechte.NewActor(agent.ID, rechte.Agent)
	return nil
}

func (s *Suite) agentListPath(org string) string {
	return "/api/organisationen/" + url.PathEscape(s.organizations[org]) + "/agenten"
}

func (s *Suite) scopePath(org, agent string) string {
	return s.agentListPath(org) + "/" + url.PathEscape(s.agents[agent]) + "/datenbereich"
}

func (s *Suite) grantRead(agent, project string) error {
	return s.replaceScopes(agent, []map[string]any{{"project_id": s.projects[project], "can_read": true, "can_write": false}})
}

func (s *Suite) grantReadWrite(agent, project string) error {
	return s.replaceScopes(agent, []map[string]any{{"project_id": s.projects[project], "can_read": true, "can_write": true}})
}

func (s *Suite) grantTwoRead(first, second, agent string) error {
	return s.replaceScopes(agent, []map[string]any{{"project_id": s.projects[first], "can_read": true}, {"project_id": s.projects[second], "can_read": true}})
}

func (s *Suite) grantWithDisabledForeign(first, second string) error {
	return s.grantWithDisabledTarget(first, s.projects[second])
}

func (s *Suite) grantWithDisabledUnknown(first string) error {
	return s.grantWithDisabledTarget(first, "unbekanntes-projekt")
}

func (s *Suite) grantWithDisabledTarget(first, targetID string) error {
	return s.replaceScopes("Mira", []map[string]any{{"project_id": s.projects[first], "can_read": true},
		{"project_id": targetID, "can_read": false, "can_write": false}})
}

func (s *Suite) revokeWrite(agent, project string) error {
	return s.grantRead(agent, project)
}

func (s *Suite) replaceScopes(agent string, scopes []map[string]any) error {
	body, _ := json.Marshal(map[string]any{"scopes": scopes})
	return s.requestScope(http.MethodPut, s.scopePath("Nordstern", agent), string(body), "application/json")
}

func (s *Suite) onlyReadScope(project string) error {
	if err := s.requestScope(http.MethodGet, s.scopePath("Nordstern", s.activeAgent), "", ""); err != nil {
		return err
	}

	var response ScopeResponse
	if err := json.Unmarshal(s.response.Body, &response); err != nil {
		return err
	}
	if s.response.Status != http.StatusOK || len(response.Scopes) != 1 || response.Scopes[0].ProjectID != s.projects[project] || !response.Scopes[0].CanRead || response.Scopes[0].CanWrite {
		return fmt.Errorf("erwartete Lesefreigabe fehlt: HTTP %d %s", s.response.Status, s.response.Body)
	}

	return nil
}

func (s *Suite) readWriteScope(agent, project string) error {
	if err := s.requestScope(http.MethodGet, s.scopePath("Nordstern", agent), "", ""); err != nil {
		return err
	}

	var response ScopeResponse
	if err := json.Unmarshal(s.response.Body, &response); err != nil {
		return err
	}
	if s.response.Status != http.StatusOK || len(response.Scopes) != 1 || response.Scopes[0].ProjectID != s.projects[project] || !response.Scopes[0].CanRead || !response.Scopes[0].CanWrite {
		return fmt.Errorf("Lese-/Schreibfreigabe fehlt: HTTP %d %s", s.response.Status, s.response.Body)
	}

	return nil
}

func (s *Suite) atomicForeignDenied() error {
	if s.response == nil || s.response.Status != http.StatusNotFound || strings.Contains(string(s.response.Body), s.projects["Fremd"]) {
		return fmt.Errorf("fremde Freigabe nicht datenfrei verweigert: %+v", s.response)
	}

	return nil
}

func (s *Suite) getForeignAgentScope(agent, org string) error {
	return s.requestScope(http.MethodGet, s.scopePath(org, agent), "", "")
}

func (s *Suite) notFoundWithoutData() error {
	if s.response == nil || s.response.Status != http.StatusNotFound || strings.Contains(string(s.response.Body), "Mira") {
		return fmt.Errorf("fremde Agentendaten in Antwort: %+v", s.response)
	}

	return nil
}
