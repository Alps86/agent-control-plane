package codexagent

import (
	"encoding/json"
	"fmt"
	"github.com/cucumber/godog"
	"net/http"
	"strings"
)

func (s *Suite) initialize(sc *godog.ScenarioContext) {
	sc.After(s.after)
	sc.Step(`^ich starte die Anwendung mit einer neuen SQLite-Datenbank als Betreiberin$`, s.fresh)
	sc.Step(`^die Organisation "([^"]+)" wurde über die öffentliche Anwendung angelegt$`, s.createOrganization)
	sc.Step(`^ich in "([^"]+)" den Agenten "([^"]+)" mit der Ausführungsart "Codex CLI" über HTTP anlege$`, s.createAgent)
	sc.Step(`^antwortet die Anwendung mit einer neuen Agentenkennung$`, s.agentCreated)
	sc.Step(`^die Agentenübersicht von "([^"]+)" enthält "([^"]+)" genau einmal mit der Ausführungsart "Codex CLI"$`, s.listOnce)
	sc.Step(`^enthält die Agentenübersicht von "([^"]+)" "([^"]+)" genau einmal mit der Ausführungsart "Codex CLI"$`, s.listOnce)
	sc.Step(`^die Agentendetails nennen "([^"]+)" als Organisation$`, s.detailOrganization)
	sc.Step(`^die Agentendetails zeigen die freigegebenen benannten Fachfähigkeiten oder ausdrücklich "Keine Fachfähigkeiten freigegeben"$`, s.detailCapabilities)
	sc.Step(`^ich die Anwendung mit derselben SQLite-Datenbank neu starte$`, s.restart)
	sc.Step(`^die Organisation "([^"]+)" und ihr Codex-CLI-Agent "([^"]+)" bestehen$`, s.organizationAndAgent)
	sc.Step(`^für das gewählte Codex-CLI-Profil ist die Sperre direkter Shell-, Terminal-, Prozess-, Interpreter-, HTTP- und Dateisystemwege nicht nachgewiesen$`, s.noEnforcement)
	sc.Step(`^ich "([^"]+)" über die öffentliche Agenten-API abfrage$`, s.getNamedAgent)
	sc.Step(`^"([^"]+)" ist nicht startbereit$`, s.notReady)
	sc.Step(`^ist "([^"]+)" nicht startbereit$`, s.notReady)
	sc.Step(`^die Antwort nennt den fehlenden Durchsetzungsnachweis als Grund$`, s.enforcementReason)
	sc.Step(`^für "([^"]+)" fehlt eine zum Start erforderliche Codex-CLI-Konfiguration$`, s.noConfiguration)
	sc.Step(`^die Antwort benennt die fehlende Konfiguration$`, s.configurationReason)
	sc.Step(`^die Antwort behauptet weder einen gestarteten Lauf noch einen verbundenen Modellanbieter$`, s.noRunOrModel)
	sc.Step(`^ein HTTP-Aufruf ohne eindeutige Betreiberidentität in "([^"]+)" den Codex-CLI-Agenten "([^"]+)" anlegt$`, s.unassignedCreate)
	sc.Step(`^der Aufruf ohne Agenten-Teilerzeugung verweigert$`, s.deniedCreate)
	sc.Step(`^wird der Aufruf ohne Agenten-Teilerzeugung verweigert$`, s.deniedCreate)
	sc.Step(`^die berechtigte Agentenübersicht von "([^"]+)" enthält "([^"]+)" nicht$`, s.agentAbsent)
	sc.Step(`^"([^"]+)" ist ein Codex-CLI-Agent von "([^"]+)"$`, s.existingAgent)
	sc.Step(`^die Organisation "([^"]+)" wurde über die öffentliche Anwendung angelegt$`, s.createOrganization)
	sc.Step(`^ich "([^"]+)" über die öffentliche Agenten-URL von "([^"]+)" abfrage$`, s.crossOrganization)
	sc.Step(`^ich dieselbe datenfreie Ablehnung wie bei einer unbekannten Agentenkennung$`, s.sameUnknown)
	sc.Step(`^erhalte ich dieselbe datenfreie Ablehnung wie bei einer unbekannten Agentenkennung$`, s.sameUnknown)
	sc.Step(`^der Name "([^"]+)" sowie seine Fachfähigkeiten erscheinen nicht in der Antwort$`, s.noLeak)
	s.registerBrowser(sc)
}

func (s *Suite) createOrganization(name string) error {
	body, _ := json.Marshal(map[string]string{"name": name, "description": "Testorganisation"})
	if err := s.request("POST", "/api/organisationen", string(body), "application/json"); err != nil {
		return err
	}
	if s.response.Status != 201 {
		return fmt.Errorf("Organisation: HTTP %d: %s", s.response.Status, s.response.Body)
	}
	var item Organization
	if err := json.Unmarshal(s.response.Body, &item); err != nil {
		return err
	}
	if item.ID == "" || item.Name != name {
		return fmt.Errorf("Organisation ungültig: %s", s.response.Body)
	}
	s.orgID = item.ID
	return nil
}
func (s *Suite) createAgent(organization, name string) error {
	if err := s.requireOrganization(organization); err != nil {
		return err
	}
	body, _ := json.Marshal(map[string]string{"name": name, "template_id": "codex-cli", "execution_kind": "codex_cli"})
	s.agentName = name
	return s.request("POST", s.agentsPath(), string(body), "application/json")
}
func (s *Suite) agentCreated() error {
	if s.response.Status != 201 {
		return fmt.Errorf("Agent: HTTP %d: %s", s.response.Status, s.response.Body)
	}
	var agent Agent
	if err := json.Unmarshal(s.response.Body, &agent); err != nil {
		return err
	}
	if agent.ID == "" || agent.Name != s.agentName || agent.ExecutionKind != "codex_cli" {
		return fmt.Errorf("Agentenantwort ungültig: %s", s.response.Body)
	}
	s.agentID = agent.ID
	if s.response.Location != s.agentsPath()+"/"+agent.ID {
		return fmt.Errorf("Location %q", s.response.Location)
	}
	return nil
}
func (s *Suite) organizationAndAgent(organization, name string) error {
	if err := s.fresh(); err != nil {
		return err
	}
	if err := s.createOrganization(organization); err != nil {
		return err
	}
	if err := s.createAgent(organization, name); err != nil {
		return err
	}
	return s.agentCreated()
}
func (s *Suite) existingAgent(name, organization string) error {
	return s.organizationAndAgent(organization, name)
}
func (s *Suite) requireOrganization(name string) error {
	if s.orgID == "" {
		return fmt.Errorf("Organisation %q fehlt", name)
	}
	return nil
}
func (s *Suite) agentsPath() string { return "/api/organisationen/" + s.orgID + "/agenten" }
func (s *Suite) getAgent() error    { return s.request("GET", s.agentsPath()+"/"+s.agentID, "", "") }
func (s *Suite) getNamedAgent(name string) error {
	if name != s.agentName {
		return fmt.Errorf("Agent %q statt %q", name, s.agentName)
	}
	return s.getAgent()
}
func (s *Suite) listOnce(organization, name string) error {
	if err := s.requireOrganization(organization); err != nil {
		return err
	}
	if err := s.request("GET", s.agentsPath(), "", ""); err != nil {
		return err
	}
	if s.response.Status != 200 {
		return fmt.Errorf("Liste: HTTP %d", s.response.Status)
	}
	var list AgentList
	if err := json.Unmarshal(s.response.Body, &list); err != nil {
		return err
	}
	count := 0
	for _, item := range list.Agents {
		if item.Name == name && item.ExecutionKind == "codex_cli" {
			count++
		}
	}
	if count != 1 {
		return fmt.Errorf("%q: %d Codex-CLI-Agenten: %s", name, count, s.response.Body)
	}
	return nil
}
func (s *Suite) detailOrganization(organization string) error {
	if err := s.getAgent(); err != nil {
		return err
	}
	var agent Agent
	if err := json.Unmarshal(s.response.Body, &agent); err != nil {
		return err
	}
	if s.response.Status != 200 || agent.OrganizationID != s.orgID {
		return fmt.Errorf("Organisation %q nicht gebunden: %s", organization, s.response.Body)
	}
	return nil
}
func (s *Suite) detailCapabilities() error {
	var agent Agent
	if err := json.Unmarshal(s.response.Body, &agent); err != nil {
		return err
	}
	if agent.Capabilities == nil || len(agent.Capabilities) != 0 {
		return fmt.Errorf("Codex-CLI-Vorlage muss eine explizit leere Fähigkeitsliste haben: %s", s.response.Body)
	}
	path := "/organisationen/" + s.orgID + "/agenten/" + s.agentID
	if err := s.request("GET", path, "", ""); err != nil {
		return err
	}
	if s.response.Status != 200 || !strings.Contains(string(s.response.Body), "Keine Fachfähigkeiten freigegeben") {
		return fmt.Errorf("expliziter UI-Leerstand fehlt: HTTP %d: %s", s.response.Status, s.response.Body)
	}
	return nil
}
func (s *Suite) noEnforcement() error {
	s.cliReference = "codex"
	if err := s.restart(); err != nil {
		return err
	}
	if err := s.request("GET", s.agentsPath()+"/"+s.agentID+"/bereitschaft", "", ""); err != nil {
		return err
	}
	var readiness Readiness
	if err := json.Unmarshal(s.response.Body, &readiness); err != nil {
		return err
	}
	if s.response.Status != 200 || readiness.Ready || readiness.Code != "codex_cli_enforcement_unverified" {
		return fmt.Errorf("Durchsetzung fälschlich belegt: %s", s.response.Body)
	}
	return nil
}
func (s *Suite) noConfiguration(name string) error {
	if err := s.getNamedAgent(name); err != nil {
		return err
	}
	var agent Agent
	if err := json.Unmarshal(s.response.Body, &agent); err != nil {
		return err
	}
	if agent.Readiness.Code != "codex_cli_reference_missing" {
		return fmt.Errorf("CLI-Konfiguration nicht als fehlend markiert: %s", s.response.Body)
	}
	return nil
}
func (s *Suite) notReady(name string) error {
	if name != s.agentName {
		return fmt.Errorf("Agent %q unbekannt", name)
	}
	var agent Agent
	if err := json.Unmarshal(s.response.Body, &agent); err != nil {
		return err
	}
	if s.response.Status != 200 || agent.Readiness.Ready {
		return fmt.Errorf("Agent startbereit oder Fehler: %s", s.response.Body)
	}
	return nil
}
func (s *Suite) enforcementReason() error {
	var agent Agent
	if err := json.Unmarshal(s.response.Body, &agent); err != nil {
		return err
	}
	if !strings.Contains(strings.ToLower(agent.Readiness.Reason), "nicht nachgewiesen") {
		return fmt.Errorf("Durchsetzungsgrund fehlt: %s", s.response.Body)
	}
	return s.startDenied()
}
func (s *Suite) configurationReason() error {
	var agent Agent
	if err := json.Unmarshal(s.response.Body, &agent); err != nil {
		return err
	}
	if !strings.Contains(strings.ToLower(agent.Readiness.Reason), "konfiguration") {
		return fmt.Errorf("Konfigurationsgrund fehlt: %s", s.response.Body)
	}
	return nil
}
func (s *Suite) noRunOrModel() error {
	text := strings.ToLower(string(s.response.Body))
	if strings.Contains(text, "run_id") || strings.Contains(text, "provider") || strings.Contains(text, "model_id") {
		return fmt.Errorf("Lauf/Modell behauptet: %s", text)
	}
	return s.startDenied()
}
func (s *Suite) startDenied() error {
	if err := s.request("POST", s.agentsPath()+"/"+s.agentID+"/start", "", ""); err != nil {
		return err
	}
	if s.response.Status != http.StatusConflict {
		return fmt.Errorf("Start nicht gesperrt: HTTP %d: %s", s.response.Status, s.response.Body)
	}
	return nil
}
func (s *Suite) unassignedCreate(organization, name string) error {
	if err := s.requireOrganization(organization); err != nil {
		return err
	}
	if err := s.startNegative(); err != nil {
		return err
	}
	return s.postUnassigned(name)
}
func (s *Suite) deniedCreate() error {
	if s.response.Status != 403 {
		return fmt.Errorf("Zugriff nicht verweigert: HTTP %d: %s", s.response.Status, s.response.Body)
	}
	if !strings.Contains(string(s.response.Body), "access_denied") || s.response.Location != "" {
		return fmt.Errorf("Ablehnung enthält unerwartete Daten: %s", s.response.Body)
	}
	return nil
}
func (s *Suite) agentAbsent(organization, name string) error {
	if err := s.requireOrganization(organization); err != nil {
		return err
	}
	if err := s.request("GET", s.agentsPath(), "", ""); err != nil {
		return err
	}
	if strings.Contains(string(s.response.Body), `"name":"`+name+`"`) {
		return fmt.Errorf("Agent trotz Ablehnung gespeichert: %s", s.response.Body)
	}
	return nil
}
func (s *Suite) crossOrganization(name, organization string) error {
	if name != s.agentName {
		return fmt.Errorf("Agent %q unbekannt", name)
	}
	if err := s.requireOrganization(organization); err != nil {
		return err
	}
	return s.getAgent()
}
func (s *Suite) sameUnknown() error {
	if s.response.Status != 404 {
		return fmt.Errorf("Fremdkennung HTTP %d", s.response.Status)
	}
	foreign := string(s.response.Body)
	if err := s.request("GET", s.agentsPath()+"/unbekannt", "", ""); err != nil {
		return err
	}
	if s.response.Status != 404 || string(s.response.Body) != foreign {
		return fmt.Errorf("abweichende datenfreie Ablehnung")
	}
	return nil
}
func (s *Suite) noLeak(name string) error {
	if strings.Contains(string(s.response.Body), name) || strings.Contains(string(s.response.Body), "capabilities") {
		return fmt.Errorf("Fremddaten enthalten: %s", s.response.Body)
	}
	return nil
}
