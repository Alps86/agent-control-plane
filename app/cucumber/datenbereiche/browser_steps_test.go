package datenbereiche

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
	sc.Step(`^ich "([^"]+)" in einem echten Browser öffne$`, s.browserOpen)
	sc.Step(`^ich "([^"]+)" Datenbereich in einem echten Browser öffne$`, s.browserOpen)
	sc.Step(`^sehe ich, dass keine Projektfreigabe besteht$`, s.browserNoGrant)
	sc.Step(`^ich "([^"]+)" zum Lesen freigebe und speichere$`, s.browserGrantRead)
	sc.Step(`^zeigt die Seite "([^"]+)" mit Lesen und ohne Schreiben$`, s.browserOnlyRead)
	sc.Step(`^ich die Lesefreigabe wieder entziehe und speichere$`, s.browserRevokeRead)
	sc.Step(`^zeigt die Seite erneut keinen Datenbereich für "([^"]+)"$`, s.browserNoGrantForAgent)
	sc.Step(`^sehe ich "([^"]+)" nicht als auswählbares Projekt$`, s.browserNoForeignProject)
	sc.Step(`^ich die Kennung von "([^"]+)" direkt an den Speichern-Endpunkt sende$`, s.browserPostForeign)
	sc.Step(`^wird die Änderung ohne fremde Projektdaten abgewiesen$`, s.atomicForeignDenied)
	sc.Step(`^"([^"]+)" erhält keine Freigabe für "([^"]+)"$`, s.browserNoForeignGrant)
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

func (s *Suite) browserOpen(agent string) error {
	if err := s.agentIdentity(agent); err != nil {
		return err
	}
	if err := s.startBrowser(); err != nil {
		return err
	}

	profile := s.agentListPath("Nordstern") + "/" + s.agents[agent]
	if err := s.browserCommand(map[string]any{"url": s.baseURL() + strings.TrimPrefix(profile, "/api"),
		"wait": map[string]string{"heading": agent}}); err != nil {
		return err
	}

	return s.browserCommand(map[string]any{"click": `main a[href$="/datenbereich"]`,
		"wait": map[string]string{"heading": "Datenbereich von " + agent}})
}

func (s *Suite) browserNoGrant() error {
	if !strings.Contains(s.page.Status, "Keine Projektfreigabe") {
		return fmt.Errorf("erwarteter Leerzustand fehlt: %+v", s.page)
	}

	return nil
}

func (s *Suite) browserGrantRead(project string) error {
	if err := s.browserSetRead(project, true); err != nil {
		return err
	}

	return s.browserSave()
}

func (s *Suite) browserSetRead(project string, checked bool) error {
	selector := fmt.Sprintf(`main input[name="read_project_ids"][value="%s"]`, s.projects[project])
	return s.browserCommand(map[string]any{"check": map[string]any{"selector": selector, "checked": checked}})
}

func (s *Suite) browserSave() error {
	return s.browserCommand(map[string]any{"submit": `main form[action$="/datenbereich"]`,
		"wait": map[string]string{"heading": "Datenbereich von Mira"}})
}

func (s *Suite) browserOnlyRead(project string) error {
	if !strings.Contains(s.page.Status, project+" mit Lesen und ohne Schreiben") {
		return fmt.Errorf("Lesefreigabe fehlt: %+v", s.page)
	}

	return nil
}

func (s *Suite) browserRevokeRead() error {
	if err := s.browserSetRead("Website", false); err != nil {
		return err
	}

	return s.browserSave()
}

func (s *Suite) browserNoGrantForAgent(agent string) error {
	if s.page.Heading != "Datenbereich von "+agent {
		return fmt.Errorf("falscher Agent im Browser: %+v", s.page)
	}

	return s.browserNoGrant()
}

func (s *Suite) browserNoForeignProject(project string) error {
	if strings.Contains(s.page.Text, project) {
		return fmt.Errorf("fremdes Projekt im Browser sichtbar: %+v", s.page)
	}

	return nil
}

func (s *Suite) browserPostForeign(project string) error {
	body := "read_project_ids=" + s.projects[project]
	path := strings.TrimPrefix(s.scopePath("Nordstern", "Mira"), "/api")
	return s.requestScope("POST", path, body, "application/x-www-form-urlencoded")
}

func (s *Suite) browserNoForeignGrant(agent, project string) error {
	if err := s.requestScope("GET", s.scopePath("Nordstern", agent), "", ""); err != nil {
		return err
	}

	if strings.Contains(string(s.response.Body), s.projects[project]) {
		return fmt.Errorf("fremde Freigabe sichtbar: %s", s.response.Body)
	}

	return nil
}

func (s *Suite) stopBrowser() {
	if s.browser == nil || s.browser.Process == nil {
		return
	}

	_ = syscall.Kill(-s.browser.Process.Pid, syscall.SIGKILL)
	_ = s.browser.Wait()
	s.browser = nil
}
