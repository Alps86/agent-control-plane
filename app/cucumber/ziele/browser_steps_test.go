package ziele

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
	sc.Step(`^ich öffne die Anwendung mit einer neuen SQLite-Datenbank im Browser$`, s.browserFresh)
	sc.Step(`^ich lege die Organisation "([^"]*)" an$`, s.browserCreateOrganization)
	sc.Step(`^ich lege die Organisationen "([^"]*)" und "([^"]*)" an$`, s.browserCreateOrganizations)
	sc.Step(`^ich die Organisationsseite von "([^"]*)" öffne$`, s.browserOpenOrganization)
	sc.Step(`^ich die Aktion "Ziele" öffne$`, s.browserOpenGoals)
	sc.Step(`^sehe ich die leere Zielübersicht von "([^"]*)" mit der Aktion "Ziel anlegen"$`, s.browserEmptyGoals)
	sc.Step(`^ich "Ziel anlegen" öffne$`, s.browserOpenGoalForm)
	sc.Step(`^ich öffne das Formular "Ziel anlegen" von "([^"]*)"$`, s.browserOpenGoalFormFor)
	sc.Step(`^sehe ich das Feld "Name" für ein Stammziel in "([^"]*)"$`, s.browserGoalField)
	sc.Step(`^ich als Zielname "([^"]*)" eingebe$`, s.browserFillGoal)
	sc.Step(`^ich das Zielformular speichere$`, s.browserSubmitGoal)
	s.registerBrowserAssertions(sc)
}

func (s *Suite) registerBrowserAssertions(sc *godog.ScenarioContext) {
	sc.Step(`^sehe ich "([^"]*)" in der Zielübersicht von "([^"]*)"$`, s.browserGoalVisible)
	sc.Step(`^ich sehe das Ziel als Stammziel ohne übergeordnetes Ziel$`, s.browserRootGoal)
	sc.Step(`^ich die Anwendung mit derselben SQLite-Datenbank neu starte und die Seite neu lade$`, s.browserRestart)
	sc.Step(`^sehe ich "([^"]*)" erneut in der Zielübersicht von "([^"]*)"$`, s.browserGoalVisible)
	sc.Step(`^sehe ich einen Hinweis am Feld "Name"$`, s.browserNameError)
	sc.Step(`^in der Zielübersicht von "([^"]*)" erscheint kein Ziel$`, s.browserNoGoal)
	sc.Step(`^ich lege das Stammziel "([^"]*)" in "([^"]*)" an$`, s.browserCreateGoal)
	sc.Step(`^ich die Zielübersicht von "([^"]*)" öffne$`, s.browserOpenGoalsFor)
	sc.Step(`^sehe ich dort "([^"]*)" nicht$`, s.browserGoalAbsent)
	sc.Step(`^ich sehe, dass die Zielübersicht zu "([^"]*)" gehört$`, s.browserGoalsBelongTo)
}

func (s *Suite) browserFresh() error {
	if err := s.freshServer(); err != nil {
		return err
	}

	if err := s.startBrowser(); err != nil {
		return err
	}

	return s.browserNavigate("/organisationen", "org-list")
}

func (s *Suite) startBrowser() error {
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

func (s *Suite) stopBrowser() {
	if s.browser == nil || s.browser.Process == nil {
		return
	}

	_ = syscall.Kill(-s.browser.Process.Pid, syscall.SIGKILL)
	_ = s.browser.Wait()
	s.browser = nil
}

func (s *Suite) browserCommand(input map[string]any) error {
	data, err := json.Marshal(input)
	if err != nil {
		return err
	}

	if _, err := s.stdin.Write(append(data, '\n')); err != nil {
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

func (s *Suite) browserNavigate(path, mode string) error {
	return s.browserCommand(map[string]any{"url": s.baseURL() + path, "mode": mode})
}

func (s *Suite) browserClick(selector, mode string) error {
	return s.browserCommand(map[string]any{"click": selector, "mode": mode})
}

func (s *Suite) browserCreateOrganization(name string) error {
	if err := s.browserNavigate("/organisationen", "org-list"); err != nil {
		return err
	}

	if err := s.browserClick(`a[href="/organisationen/neu"]`, "org-form"); err != nil {
		return err
	}

	if err := s.browserCommand(map[string]any{"fill": name, "mode": "org-form"}); err != nil {
		return err
	}

	if err := s.browserCommand(map[string]any{"submit": true, "mode": "org-detail"}); err != nil {
		return err
	}

	return s.rememberBrowserOrganization(name)
}

func (s *Suite) rememberBrowserOrganization(name string) error {
	id := strings.TrimPrefix(strings.TrimPrefix(s.page.URL, s.baseURL()), "/organisationen/")
	if id == "" || s.page.Heading != name {
		return fmt.Errorf("Organisationsdetail fehlt: %+v", s.page)
	}

	s.organizations[name], s.active = id, name
	return nil
}

func (s *Suite) browserCreateOrganizations(first, second string) error {
	if err := s.browserCreateOrganization(first); err != nil {
		return err
	}

	return s.browserCreateOrganization(second)
}

func (s *Suite) browserOpenOrganization(name string) error {
	if err := s.browserNavigate("/organisationen", "org-list"); err != nil {
		return err
	}

	s.active = name
	return s.browserClick(`a[href="/organisationen/`+s.organizations[name]+`"]`, "org-detail")
}

func (s *Suite) browserOpenGoals() error {
	path := "/organisationen/" + s.organizations[s.active] + "/ziele"
	return s.browserClick(`a[href="`+path+`"]`, "goals")
}

func (s *Suite) browserEmptyGoals(name string) error {
	if s.page.Heading != "Zielübersicht" || !strings.Contains(s.page.Text, "Noch keine Ziele") {
		return fmt.Errorf("leere Zielübersicht fehlt: %+v", s.page)
	}

	if !strings.Contains(s.page.Text, name) || !s.hasLink("Ziel anlegen") || len(s.page.Goals) != 0 {
		return fmt.Errorf("Aktion, Organisation oder Leerzustand fehlt: %+v", s.page)
	}

	return nil
}

func (s *Suite) hasLink(label string) bool {
	for _, link := range s.page.Links {
		if strings.Contains(link, label) {
			return true
		}
	}

	return false
}

func (s *Suite) browserOpenGoalForm() error {
	path := "/organisationen/" + s.organizations[s.active] + "/ziele/neu"
	return s.browserClick(`a[href="`+path+`"]`, "goal-form")
}

func (s *Suite) browserOpenGoalFormFor(name string) error {
	if err := s.browserOpenOrganization(name); err != nil {
		return err
	}

	if err := s.browserOpenGoals(); err != nil {
		return err
	}

	return s.browserOpenGoalForm()
}

func (s *Suite) browserGoalField(name string) error {
	if s.page.Heading != "Ziel anlegen" || !strings.Contains(s.page.Text, name) || !strings.Contains(s.page.Text, "Name") {
		return fmt.Errorf("Zielformular fehlt: %+v", s.page)
	}

	return nil
}

func (s *Suite) browserFillGoal(name string) error {
	if err := s.browserCommand(map[string]any{"fill": name, "mode": "goal-form"}); err != nil {
		return err
	}

	if s.page.Name != name {
		return fmt.Errorf("Zielname %q statt %q", s.page.Name, name)
	}

	return nil
}

func (s *Suite) browserSubmitGoal() error {
	mode := "goals"
	if strings.TrimSpace(s.page.Name) == "" {
		mode = "error"
	}

	return s.browserCommand(map[string]any{"submit": true, "mode": mode})
}

func (s *Suite) browserGoalVisible(goal, organization string) error {
	if s.page.Heading != "Zielübersicht" || !strings.Contains(s.page.Text, organization) {
		return fmt.Errorf("Zielübersicht fehlt: %+v", s.page)
	}

	count := 0
	for _, item := range s.page.Goals {
		if item == goal {
			count++
		}
	}

	if count != 1 {
		return fmt.Errorf("Ziel %q %d-mal sichtbar: %+v", goal, count, s.page)
	}

	return nil
}

func (s *Suite) browserRootGoal() error {
	if !strings.Contains(s.page.Text, "Stammziel") || !strings.Contains(s.page.Text, "Kein übergeordnetes Ziel") {
		return fmt.Errorf("Stammzielhinweis fehlt: %s", s.page.Text)
	}

	return nil
}

func (s *Suite) browserRestart() error {
	if err := s.restartServer(); err != nil {
		return err
	}

	return s.browserNavigate("/organisationen/"+s.organizations[s.active]+"/ziele", "goals")
}

func (s *Suite) browserNameError() error {
	if s.page.Heading != "Ziel anlegen" || strings.TrimSpace(s.page.Alert) == "" {
		return fmt.Errorf("Feldhinweis fehlt: %+v", s.page)
	}

	return nil
}

func (s *Suite) browserNoGoal(name string) error {
	if err := s.browserNavigate("/organisationen/"+s.organizations[name]+"/ziele", "goals"); err != nil {
		return err
	}

	if len(s.page.Goals) != 0 {
		return fmt.Errorf("ungültiges Ziel gespeichert: %+v", s.page.Goals)
	}

	return nil
}

func (s *Suite) browserCreateGoal(goal, organization string) error {
	if err := s.browserOpenGoalFormFor(organization); err != nil {
		return err
	}

	if err := s.browserFillGoal(goal); err != nil {
		return err
	}

	return s.browserSubmitGoal()
}

func (s *Suite) browserOpenGoalsFor(name string) error {
	if err := s.browserOpenOrganization(name); err != nil {
		return err
	}

	return s.browserOpenGoals()
}

func (s *Suite) browserGoalAbsent(goal string) error {
	for _, item := range s.page.Goals {
		if item == goal {
			return fmt.Errorf("fremdes Ziel sichtbar: %+v", s.page.Goals)
		}
	}

	return nil
}

func (s *Suite) browserGoalsBelongTo(name string) error {
	if !strings.Contains(s.page.Text, name) || !strings.Contains(s.page.URL, s.organizations[name]) {
		return fmt.Errorf("falsche Organisation in Übersicht: %+v", s.page)
	}

	return nil
}
