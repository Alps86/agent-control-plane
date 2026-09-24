package agentenvorlagen

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/cucumber/godog"
)

func (s *Suite) registerBrowserSteps(sc *godog.ScenarioContext) {
	sc.Step(`^ich öffne die Anwendung mit der Organisation "([^"]+)" im Browser$`, s.browserOpenApplication)
	sc.Step(`^ich die Agentenübersicht von "([^"]+)" öffne$`, s.browserOpenList)
	sc.Step(`^sehe ich einen leeren Zustand mit der Aktion "Agent anlegen"$`, s.browserEmpty)
	sc.Step(`^ich "Agent anlegen" öffne$`, s.browserOpenForm)
	sc.Step(`^ich das Formular "Agent anlegen" für "([^"]+)" im Browser öffne$`, s.browserFormForOrg)
	sc.Step(`^ich öffne das Formular "Agent anlegen" für "([^"]+)" im Browser$`, s.browserFormForOrg)
	sc.Step(`^kann ich die Agentenvorlage "([^"]+)" auswählen$`, s.browserTemplateAvailable)
	sc.Step(`^ich die Vorlage "([^"]+)" auswähle$`, s.browserSelectTemplate)
	sc.Step(`^sehe ich vor dem Speichern Rolle, Anweisung und benannte Fachfähigkeiten der Vorlage$`, s.browserPreview)
	sc.Step(`^die Ausführungsart ist "Eino"$`, s.browserEino)
	sc.Step(`^ich den Namen "([^"]+)" eingebe und den Agenten speichere$`, s.browserCreate)
	sc.Step(`^sehe ich "([^"]+)" in der Agentenübersicht von "([^"]+)"$`, s.browserListContains)
	sc.Step(`^ich das Profil von "([^"]+)" öffne$`, s.browserOpenProfile)
	sc.Step(`^sehe ich Rolle, Anweisung, benannte Fachfähigkeiten und "Eino"$`, s.browserProfile)
	sc.Step(`^ich sehe "Modell nicht verbunden" sowie "Nicht startbereit"$`, s.browserNotReady)
	sc.Step(`^ich sehe keine direkte Shell-, Terminal-, Prozess-, Interpreter-, HTTP- oder Dateisystemfähigkeit$`, s.browserNoDirectCapabilities)
	s.registerBrowserFinishSteps(sc)
}

func (s *Suite) registerBrowserFinishSteps(sc *godog.ScenarioContext) {
	sc.Step(`^ich die Anwendung mit derselben SQLite-Datenbank neu starte und die Seite neu lade$`, s.browserRestart)
	sc.Step(`^sehe ich "([^"]+)" mit derselben Profiladresse und demselben Auftrag erneut$`, s.browserSameProfile)
	s.registerBrowserOtherSteps(sc)
}

func (s *Suite) registerBrowserOtherSteps(sc *godog.ScenarioContext) {
	sc.Step(`^ich öffne das Profil des unverbundenen Eino-Agenten "([^"]+)" im Browser$`, s.browserOpenUnconnectedProfile)
	sc.Step(`^ich einen Start für "([^"]+)" auslöse$`, s.browserStart)
	sc.Step(`^sehe ich den Hinweis "Modellverbindung fehlt"$`, s.browserStartReason)
	sc.Step(`^ich sehe keine Startbestätigung und keine Laufadresse$`, s.browserNoStartConfirmation)
	sc.Step(`^"([^"]+)" bleibt "Nicht startbereit"$`, s.browserStillNotReady)
	sc.Step(`^ich die Teamvorlage "([^"]+)" als Ausgangspunkt auswähle$`, s.browserSelectTeam)
	sc.Step(`^sehe ich die Agentenvorlage "([^"]+)" mit Rolle und Auftrag als Vorschau$`, s.browserTeamPreview)
	sc.Step(`^ich daraus den Agenten "([^"]+)" anlege$`, s.browserCreateFromTeam)
	sc.Step(`^sehe ich nur "([^"]+)" neu in der Agentenübersicht von "([^"]+)"$`, s.browserListOnly)
	sc.Step(`^ich als Agentennamen "([^"]*)" eingebe$`, s.browserFillName)
	sc.Step(`^ich das Formular speichere$`, s.browserSubmit)
	sc.Step(`^sehe ich einen Hinweis am Feld "Name"$`, s.browserNameError)
	sc.Step(`^in der Agentenübersicht von "([^"]+)" erscheint kein neuer Agent$`, s.browserListEmpty)
	sc.Step(`^ich im Formular "Agent anlegen" erneut die Vorlage "([^"]+)" und den Namen "([^"]+)" speichere$`, s.browserDuplicate)
	sc.Step(`^sehe ich einen Konflikthinweis am Feld "Name"$`, s.browserNameConflict)
	sc.Step(`^die Agentenübersicht von "([^"]+)" enthält "([^"]+)" genau einmal$`, s.browserListOnce)
}

func (s *Suite) startBrowser() error {
	if s.browser != nil {
		return nil
	}
	script, err := filepath.Abs("browser.mjs")
	if err != nil {
		return err
	}
	s.browser = exec.Command("node", script)
	s.browser.Stderr = os.Stderr
	s.browser.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	return s.connectBrowser()
}

func (s *Suite) connectBrowser() error {
	input, err := s.browser.StdinPipe()
	if err != nil {
		return err
	}
	output, err := s.browser.StdoutPipe()
	if err != nil {
		return err
	}
	s.stdin, s.stdout = bufio.NewWriter(input), bufio.NewScanner(output)
	return s.browser.Start()
}

func (s *Suite) browserCommand(input map[string]any) error {
	message, err := json.Marshal(input)
	if err != nil {
		return err
	}
	if _, err := s.stdin.Write(append(message, '\n')); err != nil {
		return err
	}
	if err := s.stdin.Flush(); err != nil {
		return err
	}
	return s.readBrowser()
}

func (s *Suite) readBrowser() error {
	if !s.stdout.Scan() {
		return fmt.Errorf("Chrome endete: %v", s.stdout.Err())
	}
	var reply BrowserReply
	if err := json.Unmarshal(s.stdout.Bytes(), &reply); err != nil {
		return err
	}
	if !reply.OK {
		return fmt.Errorf("Chrome: %s", reply.Error)
	}
	s.page = reply.Page
	return nil
}

func (s *Suite) browserNavigate(path, heading string) error {
	return s.browserCommand(map[string]any{"url": s.baseURL() + path, "wait": map[string]string{"heading": heading}})
}

func (s *Suite) browserOpenApplication(org string) error {
	if err := s.freshServer(); err != nil {
		return err
	}
	if err := s.ensureOrganization(org); err != nil {
		return err
	}
	if err := s.startBrowser(); err != nil {
		return err
	}
	s.selectedOrg = org
	return s.browserNavigate("/organisationen/"+s.orgIDs[org], org)
}

func (s *Suite) browserOpenList(org string) error {
	if strings.HasSuffix(s.page.URL, "/organisationen/"+s.orgIDs[org]) {
		return s.browserCommand(map[string]any{"click": `main a[href="/organisationen/` + s.orgIDs[org] + `/agenten"]`, "wait": map[string]string{"heading": "Agentenübersicht"}})
	}
	return s.browserNavigate("/organisationen/"+s.orgIDs[org]+"/agenten", "Agentenübersicht")
}

func (s *Suite) browserEmpty() error {
	if !strings.Contains(s.page.Text, "Noch keine Agenten") || !strings.Contains(s.page.Text, "Agent anlegen") || len(s.page.Cards) != 0 {
		return fmt.Errorf("kein leerer Agentenzustand: %+v", s.page)
	}
	return nil
}

func (s *Suite) browserOpenForm() error {
	return s.browserCommand(map[string]any{"click": `main a[href$="/agenten/neu"]`, "wait": map[string]string{"heading": "Agent anlegen"}})
}

func (s *Suite) browserFormForOrg(org string) error {
	if err := s.browserOpenApplication(org); err != nil {
		return err
	}
	if err := s.browserNavigate("/organisationen/"+s.orgIDs[org]+"/agenten", "Agentenübersicht"); err != nil {
		return err
	}
	return s.browserOpenForm()
}

func (s *Suite) browserTemplateAvailable(name string) error {
	if !strings.Contains(s.page.Text, name) {
		return fmt.Errorf("Vorlage %s nicht auswählbar: %s", name, s.page.Text)
	}
	return nil
}

func (s *Suite) browserSelectTemplate(name string) error {
	s.selectedAgent = name
	return s.browserCommand(map[string]any{"click": `main a[href*="vorlage=` + strings.ToLower(name) + `"]`, "wait": map[string]string{"text": "Vorschau: " + name}})
}

func (s *Suite) browserPreview() error {
	for _, part := range []string{"Vorschau:", "Rolle:", "Anweisung:", "Benannte Fachfähigkeiten"} {
		if !strings.Contains(s.page.Text, part) {
			return fmt.Errorf("Vorlagendetail %s fehlt: %s", part, s.page.Text)
		}
	}
	return nil
}

func (s *Suite) browserEino() error {
	if !strings.Contains(s.page.Text, "Ausführungsart: Eino") {
		return fmt.Errorf("Eino-Vorlage fehlt: %s", s.page.Text)
	}
	return nil
}

func (s *Suite) browserFillName(name string) error {
	s.enteredName = name
	return s.browserCommand(map[string]any{"fill": map[string]string{"name": "name", "value": name}})
}

func (s *Suite) browserSubmit() error {
	return s.browserCommand(map[string]any{"submit": `main form[action$="/agenten"]`, "wait": map[string]string{"text": s.expectedSubmitText()}})
}

func (s *Suite) expectedSubmitText() string {
	if strings.TrimSpace(s.enteredName) == "" {
		return "Bitte geben Sie einen Namen ein."
	}
	if strings.EqualFold(strings.TrimSpace(s.enteredName), "mira") && s.initialID != "" {
		return "bereits vergeben"
	}
	return "Modell nicht verbunden"
}

func (s *Suite) browserCreate(name string) error {
	if err := s.browserFillName(name); err != nil {
		return err
	}
	if err := s.browserSubmit(); err != nil {
		return err
	}
	s.initialURL = s.page.URL
	return nil
}

func (s *Suite) browserListContains(agent, org string) error {
	if err := s.browserOpenList(org); err != nil {
		return err
	}
	return s.browserCount(agent, 1)
}

func (s *Suite) browserCount(agent string, expected int) error {
	count := 0
	for _, card := range s.page.Cards {
		if strings.Contains(card, agent) {
			count++
		}
	}
	if count != expected {
		return fmt.Errorf("%s in %d statt %d Karten: %+v", agent, count, expected, s.page.Cards)
	}
	return nil
}

func (s *Suite) browserOpenProfile(agent string) error {
	return s.browserCommand(map[string]any{"click": `main article a`, "wait": map[string]string{"heading": agent}})
}

func (s *Suite) browserProfile() error {
	for _, part := range []string{"Ausführungsart: Eino", "Rolle:", "Anweisung:", "Benannte Fachfähigkeiten"} {
		if !strings.Contains(s.page.Text, part) {
			return fmt.Errorf("Profilangabe %s fehlt: %s", part, s.page.Text)
		}
	}
	return nil
}

func (s *Suite) browserNotReady() error {
	if !strings.Contains(s.page.Text, "Modell nicht verbunden") || !strings.Contains(s.page.Text, "Nicht startbereit") {
		return fmt.Errorf("Bereitschaft falsch: %s", s.page.Text)
	}
	return nil
}

func (s *Suite) browserNoDirectCapabilities() error {
	for _, forbidden := range []string{"shell", "terminal", "prozess", "interpreter", "http", "dateisystem"} {
		if strings.Contains(strings.ToLower(s.page.Text), forbidden) {
			return fmt.Errorf("direkte Fähigkeit %s im Profil: %s", forbidden, s.page.Text)
		}
	}
	return nil
}

func (s *Suite) browserRestart() error {
	path := strings.TrimPrefix(s.page.URL, s.baseURL())
	if err := s.restartServer(); err != nil {
		return err
	}
	s.initialURL = s.baseURL() + path
	return s.browserCommand(map[string]any{"url": s.initialURL, "wait": map[string]string{"heading": "Mira"}})
}

func (s *Suite) browserSameProfile(name string) error {
	if s.page.URL != s.initialURL || s.page.Heading != name {
		return fmt.Errorf("Profiladresse nach Neustart verändert: %+v", s.page)
	}
	return s.browserProfile()
}

func (s *Suite) stopBrowser() {
	if s.browser == nil || s.browser.Process == nil {
		return
	}
	_ = syscall.Kill(-s.browser.Process.Pid, syscall.SIGKILL)
	_ = s.browser.Wait()
	s.browser = nil
}
