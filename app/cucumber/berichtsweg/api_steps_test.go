package berichtsweg

import (
	"encoding/json"
	"fmt"
	"github.com/cucumber/godog"
	"net/http"
	"strings"
)

func (s *Suite) initialize(sc *godog.ScenarioContext) {
	sc.After(s.after)
	sc.Step(`^ich starte für Story 37 einen lokalen Server mit neuer SQLite-Datenbank als Betreiberin$`, s.setupAPI)
	sc.Step(`^die Organisationen "([^"]+)" und "([^"]+)" bestehen$`, s.ensureOrganizations)
	sc.Step(`^die Eino-Agenten "([^"]+)", "([^"]+)" und "([^"]+)" bestehen in "([^"]+)"$`, s.ensureThreeAgents)
	sc.Step(`^der Eino-Agent "([^"]+)" besteht in "([^"]+)"$`, s.ensureAgent)
	sc.Step(`^ich "([^"]+)" als Vorgesetzten von "([^"]+)" in "([^"]+)" über HTTP speichere$`, s.assign)
	sc.Step(`^"([^"]+)" berichtet an "([^"]+)" in "([^"]+)"$`, s.assignGiven)
	sc.Step(`^die Aufgabe "([^"]+)" ist "([^"]+)" in "([^"]+)" zugewiesen$`, s.ensureTask)
	sc.Step(`^ich "([^"]+)" als neue Vorgesetzte von "([^"]+)" in "([^"]+)" über HTTP speichere$`, s.assign)
	sc.Step(`^zeigt das Organigramm von "([^"]+)" "([^"]+)" unmittelbar unter "([^"]+)"$`, s.chartChild)
	sc.Step(`^das Profil von "([^"]+)" zeigt genau "([^"]+)" als Vorgesetzten$`, s.profileParent)
	sc.Step(`^das Profil von "([^"]+)" zeigt weiterhin genau "([^"]+)" als Vorgesetzten$`, s.profileParent)
	sc.Step(`^beide Agenten behalten die Ausführungsart "Eino"$`, s.einoAgents)
	sc.Step(`^ich den Server beende und mit derselben SQLite-Datenbank erneut starte$`, s.restart)
	sc.Step(`^zeigt das Organigramm von "([^"]+)" denselben Berichtsweg genau einmal$`, s.chartOnce)
	sc.Step(`^zeigt das Organigramm "([^"]+)" unmittelbar unter "([^"]+)" und nicht mehr unter "([^"]+)"$`, s.chartMoved)
	sc.Step(`^das Profil von "([^"]+)" zeigt genau "([^"]+)" als Vorgesetzte$`, s.profileParent)
	sc.Step(`^die Aufgabe "([^"]+)" bleibt "([^"]+)" zugewiesen$`, s.taskAssigned)
	sc.Step(`^bleiben der neue Berichtsweg und die Zuweisung von "([^"]+)" erhalten$`, s.persistedTask)
	sc.Step(`^ich ("[^"]+") als Vorgesetzte von ("[^"]+") in "([^"]+)" über HTTP zu speichern versuche$`, s.tryAssign)
	sc.Step(`^wird die Berichtskante mit dem Grund ("[^"]+") abgewiesen$`, s.rejected)
	sc.Step(`^das Organigramm von "([^"]+)" zeigt weiterhin "([^"]+)" über "([^"]+)" über "([^"]+)"$`, s.chartChain)
	sc.Step(`^das Organigramm von "([^"]+)" bleibt unverändert$`, s.chartUnchanged)
	sc.Step(`^"([^"]+)" ist für die aktuelle serverseitige Identität nicht zugeordnet$`, s.foreignIdentity)
	sc.Step(`^ich das Organigramm und das Agentenprofil von "([^"]+)" über HTTP abrufe$`, s.foreignReads)
	sc.Step(`^erhalte ich datenfreie Antworten wie für eine unbekannte Organisationskennung$`, s.noForeignData)
	sc.Step(`^ich den Berichtsweg von "([^"]+)" über die direkte URL zu ändern versuche$`, s.foreignWrite)
	sc.Step(`^erhalte ich eine datenfreie Ablehnung ohne Änderung des Organigramms von "([^"]+)"$`, s.foreignDenied)
	sc.Step(`^in "([^"]+)" ist kein Delegationsübergang von "requested" nach "approved" freigegeben$`, s.noDelegationGiven)
	sc.Step(`^ich den Berichtsweg von "([^"]+)" und die Arbeitsregeln von "([^"]+)" über HTTP abrufe$`, s.readReportAndRules)
	sc.Step(`^ist "([^"]+)" als Berichtsempfänger erkennbar$`, s.reportRecipient)
	sc.Step(`^der Delegationsübergang von "requested" nach "approved" ist weiterhin nicht freigegeben$`, s.noDelegation)
	s.registerBrowser(sc)
}

func (s *Suite) assign(parent, agent, org string) error {
	path := "/api/organisationen/" + s.orgID(org) + "/agenten/" + s.agentID(org, agent) + "/berichtsweg"
	if err := s.call("PUT", path, map[string]string{"parent_id": s.agentID(org, parent)}); err != nil {
		return err
	}
	return s.status(http.StatusOK)
}

func (s *Suite) assignGiven(agent, parent, org string) error { return s.assign(parent, agent, org) }

func (s *Suite) chart(org string) (Lines, error) {
	if err := s.call("GET", s.chartPath(org), nil); err != nil {
		return Lines{}, err
	}
	if err := s.status(http.StatusOK); err != nil {
		return Lines{}, err
	}
	var chart Lines
	err := json.Unmarshal(s.response.Body, &chart)
	return chart, err
}

func (s *Suite) line(org, name string) (Line, error) {
	chart, err := s.chart(org)
	if err != nil {
		return Line{}, err
	}
	for _, line := range chart.Lines {
		if line.AgentID == s.agentID(org, name) {
			return line, nil
		}
	}
	return Line{}, fmt.Errorf("Agent %s fehlt im Organigramm: %+v", name, chart)
}

func (s *Suite) chartChild(org, child, parent string) error {
	line, err := s.line(org, child)
	if err != nil {
		return err
	}
	if line.ParentID != s.agentID(org, parent) {
		return fmt.Errorf("%s berichtet an %s statt %s", child, line.ParentID, parent)
	}
	return nil
}

func (s *Suite) chartOnce(org string) error {
	chart, err := s.chart(org)
	if err != nil {
		return err
	}
	counts := map[string]int{}
	for _, line := range chart.Lines {
		counts[line.AgentID]++
	}
	for _, count := range counts {
		if count != 1 {
			return fmt.Errorf("Berichtslinie mehrfach: %+v", chart)
		}
	}
	return s.chartChild(org, "Mira", "Kai")
}

func (s *Suite) chartMoved(child, parent, old string) error {
	if err := s.chartChild("Nordstern", child, parent); err != nil {
		return err
	}
	line, err := s.line("Nordstern", child)
	if err != nil {
		return err
	}
	if line.ParentID == s.agentID("Nordstern", old) {
		return fmt.Errorf("alte Berichtslinie blieb")
	}
	return nil
}

func (s *Suite) profileParent(agent, parent string) error {
	org := "Nordstern"
	if err := s.call("GET", s.profilePath(org, agent), nil); err != nil {
		return err
	}
	if err := s.status(http.StatusOK); err != nil {
		return err
	}
	var profile struct {
		ParentID   string `json:"parent_id"`
		ParentName string `json:"parent_name"`
	}
	if err := json.Unmarshal(s.response.Body, &profile); err != nil {
		return err
	}
	if profile.ParentID != s.agentID(org, parent) || profile.ParentName != parent {
		return fmt.Errorf("falscher Vorgesetzter im Profil: %+v", profile)
	}
	return nil
}

func (s *Suite) einoAgents() error {
	for _, name := range []string{"Kai", "Mira"} {
		line, err := s.line("Nordstern", name)
		if err != nil {
			return err
		}
		if line.ExecutionKind != "eino" {
			return fmt.Errorf("%s: %s", name, line.ExecutionKind)
		}
	}
	return nil
}

func (s *Suite) persistedTask(title string) error {
	if err := s.chartMoved("Mira", "Lena", "Kai"); err != nil {
		return err
	}
	return s.taskAssigned(title, "Mira")
}

func (s *Suite) tryAssign(parent, agent, org string) error {
	parentName, agentName := strings.Trim(parent, `"`), strings.Trim(agent, `"`)
	parentID := s.agentID(org, parentName)
	if parentID == "" {
		parentID = s.agentID("Südstern", parentName)
	}
	if parentID == "" {
		return fmt.Errorf("Test-Vorgesetzter %s fehlt", parentName)
	}
	path := "/api/organisationen/" + s.orgID(org) + "/agenten/" + s.agentID(org, agentName) + "/berichtsweg"
	return s.call("PUT", path, map[string]string{"parent_id": parentID})
}

func (s *Suite) rejected(reason string) error {
	reasons := map[string]string{"Selbstbezug": "self_parent", "Nachfahre und Zyklus": "cycle", "Agent nicht in dieser Organisation gefunden": "agent_not_found"}
	expected := reasons[strings.Trim(reason, `"`)]
	if expected == "" || !strings.Contains(string(s.response.Body), expected) {
		return fmt.Errorf("Ablehnung %s: HTTP %d %s", expected, s.response.Status, s.response.Body)
	}
	if s.response.Status < 400 || s.response.Status >= 500 {
		return fmt.Errorf("ungültige Kante akzeptiert: %d", s.response.Status)
	}
	return nil
}

func (s *Suite) chartChain(org, a, b, c string) error {
	if err := s.chartChild(org, b, a); err != nil {
		return err
	}
	return s.chartChild(org, c, b)
}

func (s *Suite) chartUnchanged(org string) error {
	chart, err := s.chart(org)
	if err != nil {
		return err
	}
	for _, line := range chart.Lines {
		if line.ParentID != "" {
			return fmt.Errorf("fremde Org verändert: %+v", chart)
		}
	}
	return nil
}

func (s *Suite) noDelegationGiven(org string) error {
	if err := s.call("GET", "/api/organisationen/"+s.orgID(org)+"/arbeitsregeln", nil); err != nil {
		return err
	}
	return s.noDelegation()
}
func (s *Suite) readReportAndRules(agent, org string) error {
	if _, err := s.chart(org); err != nil {
		return err
	}
	return s.call("GET", "/api/organisationen/"+s.orgID(org)+"/arbeitsregeln", nil)
}
func (s *Suite) reportRecipient(name string) error { return s.chartChild("Nordstern", name, "Kai") }
func (s *Suite) noDelegation() error {
	if err := s.status(http.StatusOK); err != nil {
		return err
	}
	var rules struct {
		Rules []struct {
			Area string `json:"area"`
			From string `json:"from"`
			To   string `json:"to"`
		} `json:"rules"`
	}
	if err := json.Unmarshal(s.response.Body, &rules); err != nil {
		return err
	}
	for _, rule := range rules.Rules {
		if rule.Area == "delegation" && rule.From == "requested" && rule.To == "approved" {
			return fmt.Errorf("Delegation unerwartet erlaubt")
		}
	}
	return nil
}
