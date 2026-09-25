package berichtsweg

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/cucumber/godog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func (s *Suite) registerBrowser(sc *godog.ScenarioContext) {
	sc.Step(`^ich öffne Story 37 mit "([^"]+)", den Eino-Agenten "([^"]+)", "([^"]+)" und "([^"]+)" im Browser$`, s.browserSetup)
	sc.Step(`^ich das Detail von "([^"]+)" vom regulären Einstieg aus öffne$`, s.browserOrgDetail)
	sc.Step(`^sehe ich die Aktion "Organigramm"$`, s.browserChartAction)
	sc.Step(`^ich das Organigramm darüber öffne$`, s.browserChartActionOpen)
	sc.Step(`^sehe ich "([^"]+)", "([^"]+)" und "([^"]+)" mit ihren Berichtslinien oder als Agenten ohne Vorgesetzten$`, s.browserAllAgents)
	sc.Step(`^ich "([^"]+)" auswähle$`, s.browserSelectAgent)
	sc.Step(`^erreiche ich ihr Agentenprofil mit dem sichtbaren Vorgesetzten$`, s.browserProfileVisible)
	sc.Step(`^die Aufgabe "([^"]+)" ist "([^"]+)" zugewiesen$`, s.browserTaskGiven)
	sc.Step(`^ich im Profil von "([^"]+)" den Berichtsweg bearbeite$`, s.browserEditProfile)
	sc.Step(`^ich "([^"]+)" als neue Vorgesetzte speichere$`, s.browserAssign)
	sc.Step(`^sehe ich "([^"]+)" als einzige Vorgesetzte von "([^"]+)"$`, s.browserParent)
	sc.Step(`^das Organigramm zeigt "([^"]+)" unmittelbar unter "([^"]+)" und nicht mehr unter "([^"]+)"$`, s.browserMoved)
	sc.Step(`^die Aufgabe "([^"]+)" zeigt weiterhin "([^"]+)" als zuständigen Agenten$`, s.browserTaskAssigned)
	sc.Step(`^ich die Seite neu lade$`, s.browserReload)
	sc.Step(`^bleiben der neue Berichtsweg und die Aufgabenzuweisung sichtbar$`, s.browserPersisted)
	sc.Step(`^ich für ("[^"]+") ("[^"]+") als Vorgesetzte zu speichern versuche$`, s.browserTryAssign)
	sc.Step(`^sehe ich den Ablehnungsgrund ("[^"]+") im Berichtswegformular$`, s.browserRejected)
	sc.Step(`^das Organigramm zeigt weiterhin "([^"]+)" über "([^"]+)" über "([^"]+)"$`, s.browserChain)
	sc.Step(`^ich den Berichtsweg von "([^"]+)" im Organigramm öffne$`, s.browserReport)
	sc.Step(`^sehe ich "([^"]+)" als Vorgesetzten von "([^"]+)"$`, s.browserParent)
	sc.Step(`^ich die Arbeitsregeln von "([^"]+)" öffne$`, s.browserRules)
	sc.Step(`^sehe ich den Delegationsübergang von "requested" nach "approved" weiterhin nicht als freigegeben$`, s.browserNoDelegation)
}

func (s *Suite) browserSetup(org, a, b, c string) error {
	s.rejectedReason = ""
	if err := s.fresh(); err != nil {
		return err
	}
	if err := s.ensureOrganization(org); err != nil {
		return err
	}
	if err := s.ensureThreeAgents(a, b, c, org); err != nil {
		return err
	}
	script, err := filepath.Abs("browser.mjs")
	if err != nil {
		return err
	}
	s.browser = exec.Command("node", script)
	s.browser.Stderr = os.Stderr
	input, err := s.browser.StdinPipe()
	if err != nil {
		return err
	}
	output, err := s.browser.StdoutPipe()
	if err != nil {
		return err
	}
	s.browserInput = bufio.NewWriter(input)
	s.browserOutput = bufio.NewScanner(output)
	if err := s.browser.Start(); err != nil {
		return err
	}
	return s.browserNavigate("/organisationen/"+s.orgID(org)+"/berichtswege", "Organigramm")
}

func (s *Suite) stopBrowser() {
	if s.browser == nil || s.browser.Process == nil {
		return
	}
	_ = s.browser.Process.Kill()
	_ = s.browser.Wait()
	s.browser = nil
}

func (s *Suite) browserCommand(input map[string]any) error {
	data, err := json.Marshal(input)
	if err != nil {
		return err
	}
	if _, err := s.browserInput.Write(append(data, '\n')); err != nil {
		return err
	}
	if err := s.browserInput.Flush(); err != nil {
		return err
	}
	if !s.browserOutput.Scan() {
		return fmt.Errorf("Chrome ohne Antwort: %v", s.browserOutput.Err())
	}
	var reply struct {
		OK    bool        `json:"ok"`
		Error string      `json:"error"`
		Page  BrowserPage `json:"page"`
	}
	if err := json.Unmarshal(s.browserOutput.Bytes(), &reply); err != nil {
		return err
	}
	if !reply.OK {
		return fmt.Errorf("Chrome: %s", reply.Error)
	}
	s.page = reply.Page
	return nil
}

func (s *Suite) browserNavigate(path, heading string) error {
	return s.browserCommand(map[string]any{"url": "http://" + s.address + path, "mode": heading})
}

func (s *Suite) browserOrgDetail(org string) error {
	if err := s.browserNavigate("/organisationen", "Organisationsübersicht"); err != nil {
		return err
	}
	return s.browserCommand(map[string]any{"click": "main a[href=\"/organisationen/" + s.orgID(org) + "\"]", "mode": org})
}

func (s *Suite) browserChartAction() error {
	if !strings.Contains(s.page.Text, "Organigramm") {
		return fmt.Errorf("Organigramm-Aktion fehlt: %s", s.page.Text)
	}
	return nil
}

func (s *Suite) browserChartActionOpen() error {
	selector := "main a[href=\"/organisationen/" + s.orgID("Nordstern") + "/berichtswege\"]"
	return s.browserCommand(map[string]any{"click": selector, "mode": "Organigramm"})
}

func (s *Suite) browserAllAgents(a, b, c string) error {
	for _, name := range []string{a, b, c} {
		if !strings.Contains(s.page.Text, name) {
			return fmt.Errorf("%s fehlt im Organigramm: %s", name, s.page.Text)
		}
	}
	return nil
}

func (s *Suite) browserSelectAgent(name string) error {
	s.selectedAgent = name
	selector := "main a[href=\"/organisationen/" + s.orgID("Nordstern") + "/agenten/" + s.agentID("Nordstern", name) + "\"]"
	return s.browserCommand(map[string]any{"click": selector, "mode": name})
}

func (s *Suite) browserProfileVisible() error {
	if !strings.Contains(s.page.Text, "Vorgesetz") {
		return fmt.Errorf("Vorgesetztenfeld fehlt im Profil: %s", s.page.Text)
	}
	return nil
}

func (s *Suite) browserTaskGiven(title, agent string) error {
	return s.ensureTask(title, agent, "Nordstern")
}

func (s *Suite) browserEditProfile(name string) error {
	if err := s.browserNavigate("/organisationen/"+s.orgID("Nordstern")+"/agenten/"+s.agentID("Nordstern", name), name); err != nil {
		return err
	}
	selector := "main a[href=\"/organisationen/" + s.orgID("Nordstern") + "/berichtswege#agent-" + s.agentID("Nordstern", name) + "\"]"
	return s.browserCommand(map[string]any{"click": selector, "mode": "Organigramm"})
}

func (s *Suite) browserAssign(parent string) error {
	agent := s.selectedAgent
	if agent == "" {
		agent = "Mira"
	}
	id := s.agentID("Nordstern", agent)
	if err := s.browserCommand(map[string]any{"select": map[string]string{"selector": "#parent-" + id, "value": s.agentID("Nordstern", parent)}}); err != nil {
		return err
	}
	selector := "main form[action=\"/organisationen/" + s.orgID("Nordstern") + "/agenten/" + id + "/berichtsweg\"]"
	return s.browserCommand(s.reportSubmit(selector, id))
}

func (s *Suite) reportSubmit(selector, agentID string) map[string]any {
	input := map[string]any{"submit": selector, "mode": "Organigramm"}
	if s.rejectedReason != "" {
		input["alertContains"] = s.rejectedReason
		return input
	}
	input["urlContains"] = "#agent-" + agentID
	return input
}

func (s *Suite) browserParent(parent, agent string) error {
	if err := s.chartChild("Nordstern", agent, parent); err != nil {
		return err
	}
	if !strings.Contains(s.page.Text, parent) {
		return fmt.Errorf("Vorgesetzter im Browser fehlt: %s", s.page.Text)
	}
	return nil
}

func (s *Suite) browserMoved(child, parent, old string) error {
	if err := s.chartMoved(child, parent, old); err != nil {
		return err
	}
	return s.browserNavigate("/organisationen/"+s.orgID("Nordstern")+"/berichtswege", "Organigramm")
}

func (s *Suite) browserTaskAssigned(title, agent string) error {
	if err := s.taskAssigned(title, agent); err != nil {
		return err
	}
	path := "/organisationen/" + s.orgID("Nordstern") + "/projekte/" + s.projects["Nordstern"] + "/aufgaben/" + s.tasks[s.key("Nordstern", title)]
	if err := s.browserNavigate(path, title); err != nil {
		return err
	}
	if !strings.Contains(s.page.Text, agent) {
		return fmt.Errorf("Zuständiger im Taskdetail fehlt: %s", s.page.Text)
	}
	return nil
}

func (s *Suite) browserReload() error {
	path := strings.TrimPrefix(s.page.URL, "http://"+s.address)
	return s.browserNavigate(path, s.page.Heading)
}

func (s *Suite) browserPersisted() error {
	if !strings.Contains(s.page.Text, "Zuständig:") || !strings.Contains(s.page.Text, "Mira") {
		return fmt.Errorf("Task nach Neuladen nicht sichtbar: %s", s.page.Text)
	}
	if err := s.taskAssigned("Startseite prüfen", "Mira"); err != nil {
		return err
	}
	if err := s.browserNavigate("/organisationen/"+s.orgID("Nordstern")+"/berichtswege", "Organigramm"); err != nil {
		return err
	}
	line := s.page.Reports["agent-"+s.agentID("Nordstern", "Mira")]
	if !strings.Contains(line, "Vorgesetzte: Lena") || strings.Contains(line, "Kai") {
		return fmt.Errorf("Berichtsweg nach Neuladen im DOM falsch: %s", line)
	}
	return s.chartMoved("Mira", "Lena", "Kai")
}

func (s *Suite) browserTryAssign(agent, parent string) error {
	s.selectedAgent = strings.Trim(agent, `"`)
	s.rejectedReason = "Nachfahre und Zyklus"
	if strings.Trim(agent, `"`) == strings.Trim(parent, `"`) {
		s.rejectedReason = "Selbstbezug"
	}

	if err := s.browserNavigate("/organisationen/"+s.orgID("Nordstern")+"/berichtswege", "Organigramm"); err != nil {
		return err
	}
	return s.browserAssign(strings.Trim(parent, `"`))
}

func (s *Suite) browserRejected(reason string) error {
	expected := strings.Trim(reason, `"`)
	if !strings.Contains(s.page.Alert, expected) {
		return fmt.Errorf("Ablehnungsgrund %s fehlt: %s", expected, s.page.Alert)
	}
	return nil
}

func (s *Suite) browserChain(a, b, c string) error { return s.chartChain("Nordstern", a, b, c) }
func (s *Suite) browserReport(name string) error {
	return s.browserNavigate("/organisationen/"+s.orgID("Nordstern")+"/berichtswege#agent-"+s.agentID("Nordstern", name), "Organigramm")
}
func (s *Suite) browserRules(org string) error {
	return s.browserNavigate("/organisationen/"+s.orgID(org)+"/arbeitsregeln", "Arbeitsregeln")
}
func (s *Suite) browserNoDelegation() error {
	if !strings.Contains(s.page.Text, "Delegation") || !strings.Contains(s.page.Text, "Keine Übergänge freigegeben") {
		return fmt.Errorf("nicht freigegebene Delegation nicht sichtbar: %s", s.page.Text)
	}
	if err := s.call("GET", "/api/organisationen/"+s.orgID("Nordstern")+"/arbeitsregeln", nil); err != nil {
		return err
	}
	return s.noDelegation()
}
