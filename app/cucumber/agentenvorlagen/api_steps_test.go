package agentenvorlagen

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/cucumber/godog"
)

func (s *Suite) initializeScenario(sc *godog.ScenarioContext) {
	sc.After(s.afterScenario)
	s.registerAPISteps(sc)
	s.registerBrowserSteps(sc)
}

func (s *Suite) registerAPISteps(sc *godog.ScenarioContext) {
	sc.Step(`^ich starte den lokalen Server mit einer neuen SQLite-Datenbank als Betreiberin$`, s.freshServer)
	sc.Step(`^die Organisation "([^"]+)" besteht$`, s.ensureOrganization)
	sc.Step(`^die Organisationen "([^"]+)" und "([^"]+)" bestehen$`, s.ensureOrganizations)
	sc.Step(`^ich die Agentenvorlage "([^"]+)" für "([^"]+)" über HTTP abrufe$`, s.getTemplate)
	sc.Step(`^zeigt die Vorlage eine Rolle, eine Anweisung und benannte Fachfähigkeiten$`, s.templateHasProfile)
	sc.Step(`^ich den Agenten "([^"]+)" aus der Vorlage "([^"]+)" für "([^"]+)" über HTTP anlege$`, s.createAgent)
	sc.Step(`^der Agent "([^"]+)" wurde in "([^"]+)" aus der Vorlage "([^"]+)" angelegt$`, s.agentExists)
	sc.Step(`^antwortet der Server mit einer neuen Agentenkennung und dem HTTP-Status 201$`, s.createdAgent)
	sc.Step(`^der Agent "([^"]+)" gehört zu "([^"]+)" und verwendet die Ausführungsart "Eino"$`, s.agentBelongsToOrg)
	sc.Step(`^die Agentendetails zeigen seine Rolle, Anweisung und benannten Fachfähigkeiten$`, s.agentDetailProfile)
	sc.Step(`^die Agentendetails zeigen den Modellstatus "nicht verbunden" und die Bereitschaft "nicht startbereit"$`, s.agentNotReady)
	sc.Step(`^ich den Server beende und mit derselben SQLite-Datenbank erneut starte$`, s.restartServer)
	sc.Step(`^finde ich "([^"]+)" unter derselben Agentenkennung in "([^"]+)" genau einmal$`, s.agentSameAfterRestart)
	sc.Step(`^Rolle, Anweisung, Fachfähigkeiten und Bereitschaft sind unverändert sichtbar$`, s.agentUnchanged)
	s.registerAPISecuritySteps(sc)
	s.registerAPITemplateSteps(sc)
}

func (s *Suite) registerAPISecuritySteps(sc *godog.ScenarioContext) {
	sc.Step(`^ich "([^"]+)" über seine Agentenkennung in "([^"]+)" per HTTP abrufe$`, s.getAgentFromOrg)
	sc.Step(`^antwortet der Server mit einem datenfreien HTTP-Status 404$`, s.notFoundWithoutData)
	sc.Step(`^die Agentenliste von "([^"]+)" enthält "([^"]+)" nicht$`, s.agentAbsent)
	sc.Step(`^ich dieselbe Agentenkennung in "([^"]+)" per HTTP abrufe$`, s.getSameAgent)
	sc.Step(`^erhalte ich "([^"]+)" mit seiner ursprünglichen Agentenkennung$`, s.sameAgentID)
	sc.Step(`^für "([^"]+)" ist keine Modellverbindung zugeordnet$`, s.noModelConnection)
	sc.Step(`^ich einen Start für "([^"]+)" über die öffentliche Anwendungsgrenze anfordere$`, s.requestStart)
	sc.Step(`^wird der Start mit dem Grund "Modellverbindung fehlt" abgewiesen$`, s.startDenied)
	sc.Step(`^die Startantwort enthält keine Laufkennung und keine Startbestätigung$`, s.noRunConfirmation)
	sc.Step(`^"([^"]+)" bleibt als "nicht startbereit" sichtbar$`, s.stillNotReady)
	sc.Step(`^ich als nicht zugeordnete Betreiberin einen Agenten für "([^"]+)" über HTTP anlege$`, s.unassignedCreate)
	sc.Step(`^verweigert der Server die Anlage ohne Teilerzeugung$`, s.unassignedDenied)
	sc.Step(`^ich als nicht zugeordnete Betreiberin die Agentenliste von "([^"]+)" über HTTP abrufe$`, s.unassignedList)
	sc.Step(`^erhalte ich keine Agentendaten$`, s.noAgentData)
	s.registerAPIOriginSteps(sc)
}

func (s *Suite) registerAPIOriginSteps(sc *godog.ScenarioContext) {
	sc.Step(`^ich "([^"]+)" mit dem Agentennamen "([^"]+)" und der Vorlage "([^"]+)" von Origin "([^"]+)" per HTTP POST aufrufe$`, s.foreignOriginPost)
	sc.Step(`^antwortet der Server mit dem HTTP-Status (\d+)$`, s.statusCode)
	sc.Step(`^ich denselben Server an einer öffentlichen Listener-Adresse neu starte$`, s.startWildcardServer)
	sc.Step(`^wird der Serverstart wegen der nicht lokalen APP_ADDR-Konfiguration abgewiesen$`, s.wildcardConfigurationRejected)
	sc.Step(`^auf dem Wildcard-Port ist kein Server erreichbar$`, s.noWildcardListener)
	sc.Step(`^ich denselben Server mit derselben SQLite-Datenbank auf Loopback neu starte$`, s.restartServer)
	sc.Step(`^bleibt die Agentenliste von "([^"]+)" leer$`, s.listEmpty)
	sc.Step(`^ein entfernter Peer am öffentlichen Agenten-Handler einen POST mit dem Host "localhost" sendet$`, s.remotePeerSpoofedPost)
	sc.Step(`^ich starte denselben Server an der Loopback-Adresse "127\.0\.0\.2" neu$`, s.startAlternateLoopbackServer)
}

func (s *Suite) registerAPITemplateSteps(sc *godog.ScenarioContext) {
	sc.Step(`^ich die Teamvorlage "([^"]+)" für "([^"]+)" über HTTP als Vorschau abrufe$`, s.getTeamTemplate)
	sc.Step(`^sehe ich die auswählbare Agentenvorlage "([^"]+)" mit Rolle und Auftrag$`, s.teamIncludesTemplate)
	sc.Step(`^ich daraus nur den Agenten "([^"]+)" für "([^"]+)" über HTTP anlege$`, s.createFromTeam)
	sc.Step(`^enthält die Agentenliste von "([^"]+)" genau "([^"]+)"$`, s.listExactlyAgent)
	sc.Step(`^es entstehen keine weiteren Agenten der Teamvorlage$`, s.noTeamImport)
	sc.Step(`^ich aus der Vorlage "([^"]+)" einen Eino-Agenten mit der zusätzlichen Fähigkeit "([^"]+)" über HTTP anlege$`, s.createWithCapability)
	sc.Step(`^wird die Fähigkeit als unzulässig abgewiesen$`, s.capabilityDenied)
	sc.Step(`^in "([^"]+)" entsteht kein Agent aus dieser Anfrage$`, s.noAgentFromRequest)
	sc.Step(`^ich einen Agenten mit dem Namen "([^"]*)" aus der Vorlage "([^"]+)" für "([^"]+)" über HTTP anlege$`, s.createAgent)
	sc.Step(`^antwortet der Server mit einem Hinweis am Feld "Name"$`, s.nameError)
	sc.Step(`^die Agentenliste von "([^"]+)" bleibt leer$`, s.listEmpty)
	sc.Step(`^ich den Agenten "([^"]+)" aus der unbekannten Vorlage "([^"]+)" für "([^"]+)" über HTTP anlege$`, s.createAgent)
	sc.Step(`^antwortet der Server mit einem Hinweis am Feld "Vorlage"$`, s.templateError)
	sc.Step(`^ich den Agenten "([^"]+)" erneut aus der Vorlage "([^"]+)" für "([^"]+)" über HTTP anlege$`, s.createAgent)
	sc.Step(`^antwortet der Server mit einem Konflikthinweis am Feld "Name"$`, s.nameConflict)
	sc.Step(`^die Agentenliste von "([^"]+)" enthält "([^"]+)" genau einmal$`, s.listAgentOnce)
}

func (s *Suite) ensureOrganization(name string) error {
	if s.address == "" {
		if err := s.freshServer(); err != nil {
			return err
		}
	}
	if s.orgIDs[name] != "" {
		return nil
	}
	return s.createOrganization(name)
}

func (s *Suite) createOrganization(name string) error {
	body, _ := json.Marshal(map[string]string{"name": name, "description": "Testorganisation " + name})
	if err := s.request("POST", "/api/organisationen", string(body), "application/json"); err != nil {
		return err
	}
	if s.response.Status != http.StatusCreated {
		return fmt.Errorf("Organisation %s: HTTP %d: %s", name, s.response.Status, s.response.Body)
	}
	var item struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(s.response.Body, &item); err != nil {
		return err
	}
	s.orgIDs[name] = item.ID
	return nil
}

func (s *Suite) ensureOrganizations(first, second string) error {
	if err := s.ensureOrganization(first); err != nil {
		return err
	}
	return s.ensureOrganization(second)
}

func (s *Suite) orgPath(name string) string {
	return "/api/organisationen/" + url.PathEscape(s.orgIDs[name]) + "/agenten"
}

func (s *Suite) agentPath(org, agent string) string {
	return s.orgPath(org) + "/" + url.PathEscape(s.agentIDs[agent])
}

func (s *Suite) createAgent(name, template, org string) error {
	if err := s.ensureOrganization(org); err != nil {
		return err
	}
	body, _ := json.Marshal(map[string]any{"name": name, "template_id": strings.ToLower(template), "execution_kind": "eino"})
	return s.request("POST", s.orgPath(org), string(body), "application/json")
}

func (s *Suite) agentExists(name, org, template string) error {
	if err := s.createAgent(name, template, org); err != nil {
		return err
	}
	return s.createdAgent()
}

func (s *Suite) createdAgent() error {
	if err := s.statusCode(http.StatusCreated); err != nil {
		return err
	}
	profile, err := s.decodeProfile()
	if err != nil {
		return err
	}
	if profile.ID == "" || profile.Name == "" || profile.OrganizationID == "" {
		return fmt.Errorf("Agentenkennung oder Kontext fehlt: %s", s.response.Body)
	}
	if s.response.Location != s.orgPath(s.orgName(profile.OrganizationID))+"/"+profile.ID {
		return fmt.Errorf("falscher Location-Header: %q", s.response.Location)
	}
	s.agentIDs[profile.Name] = profile.ID
	s.initialID = profile.ID
	s.initialProfile = profile
	return nil
}

func (s *Suite) orgName(id string) string {
	for name, candidate := range s.orgIDs {
		if candidate == id {
			return name
		}
	}
	return ""
}

func (s *Suite) statusCode(status int) error {
	if s.response == nil || s.response.Status != status {
		return fmt.Errorf("HTTP-Status %d erwartet, erhalten: %+v", status, s.response)
	}
	return nil
}
