package projekte

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
	s.registerBrowserSetup(sc)
	s.registerBrowserProjectActions(sc)
	s.registerBrowserAssertions(sc)
	s.registerBrowserSettings(sc)
}

func (s *Suite) registerBrowserSettings(sc *godog.ScenarioContext) {
	sc.Step(`^ich in der leeren Organisationsübersicht "Settings" öffne$`, s.browserOpenSettingsFromEmpty)
	sc.Step(`^ich dort "Settings" öffne$`, s.browserOpenSettingsFromDetail)
	sc.Step(`^sehe ich die globale Seite "/settings" mit Codex-Abo und OpenRouter$`, s.browserSettingsShown)
}

func (s *Suite) registerBrowserSetup(sc *godog.ScenarioContext) {
	sc.Step(`^ich öffne die Anwendung mit einer neuen SQLite-Datenbank im Browser$`, s.browserFresh)
	sc.Step(`^ich lege die Organisation "([^"]*)" an$`, s.browserCreateOrganization)
	sc.Step(`^ich lege die Organisationen "([^"]*)" und "([^"]*)" an$`, s.browserCreateOrganizations)
	sc.Step(`^ich lege das Stammziel "([^"]*)" in "([^"]*)" an$`, s.browserCreateGoal)
}

func (s *Suite) registerBrowserProjectActions(sc *godog.ScenarioContext) {
	sc.Step(`^ich die Organisationsseite von "([^"]*)" öffne$`, s.browserOpenOrganization)
	sc.Step(`^ich die Aktion "Projekte" öffne$`, s.browserOpenProjects)
	sc.Step(`^sehe ich die leere Projektübersicht von "([^"]*)" mit der Aktion "Projekt anlegen"$`, s.browserEmptyProjects)
	sc.Step(`^ich "Projekt anlegen" öffne$`, s.browserOpenProjectForm)
	sc.Step(`^ich öffne das Formular "Projekt anlegen" von "([^"]*)"$`, s.browserOpenProjectFormFor)
	sc.Step(`^ich das Formular "Projekt anlegen" von "([^"]*)" öffne$`, s.browserOpenProjectFormFor)
	sc.Step(`^sehe ich die Felder "Name" und "Beschreibung"$`, s.browserProjectFields)
	sc.Step(`^ich kann das Ziel "([^"]*)" aus "([^"]*)" auswählen$`, s.browserCanSelectGoal)
	sc.Step(`^ich als Projektnamen "([^"]*)" und als Beschreibung "([^"]*)" eingebe$`, s.browserFillProject)
	sc.Step(`^ich als Projektnamen "([^"]*)" eingebe$`, s.browserFillProjectName)
	sc.Step(`^ich das Ziel "([^"]*)" auswähle$`, s.browserSelectGoal)
	sc.Step(`^ich das Projektformular speichere$`, s.browserSubmitProject)
	sc.Step(`^ich das Projektformular ohne Zielauswahl speichere$`, s.browserSubmitWithoutGoal)
}

func (s *Suite) registerBrowserAssertions(sc *godog.ScenarioContext) {
	sc.Step(`^sehe ich "([^"]*)" mit "([^"]*)" in der Projektübersicht von "([^"]*)"$`, s.browserProjectVisible)
	sc.Step(`^ich das Projektdetail von "([^"]*)" öffne$`, s.browserOpenProjectDetail)
	sc.Step(`^sehe ich die Zuordnung zum Ziel "([^"]*)"$`, s.browserDetailGoal)
	sc.Step(`^ich sehe eine leere Aufgabenliste$`, s.browserEmptyTasks)
	sc.Step(`^ich die Anwendung mit derselben SQLite-Datenbank neu starte und die Seite neu lade$`, s.browserRestart)
	sc.Step(`^sehe ich "([^"]*)" mit "([^"]*)" und dem Ziel "([^"]*)" erneut$`, s.browserProjectAfterRestart)
	sc.Step(`^sehe ich einen Hinweis am Projektfeld "([^"]*)"$`, s.browserFieldError)
	sc.Step(`^in der Projektübersicht von "([^"]*)" erscheint kein Projekt$`, s.browserNoProject)
	sc.Step(`^kann ich "([^"]*)" nicht als Projektziel auswählen$`, s.browserCannotSelectGoal)
	sc.Step(`^ich sehe, dass das Projektformular zu "([^"]*)" gehört$`, s.browserFormBelongsTo)
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
	return s.startBrowserPipes()
}

func (s *Suite) startBrowserPipes() error {
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
	if err := s.writeBrowserCommand(input); err != nil {
		return err
	}

	return s.readBrowserReply()
}

func (s *Suite) writeBrowserCommand(input map[string]any) error {
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

	return nil
}

func (s *Suite) readBrowserReply() error {
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

	return s.browserSubmitOrganization(name)
}

func (s *Suite) browserSubmitOrganization(name string) error {
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

func (s *Suite) browserCreateGoal(goal, organization string) error {
	if err := s.browserOpenOrganization(organization); err != nil {
		return err
	}

	path := "/organisationen/" + s.organizations[organization] + "/ziele"
	if err := s.browserClick(`a[href="`+path+`"]`, "goals"); err != nil {
		return err
	}

	if err := s.browserClick(`a[href="`+path+`/neu"]`, "goal-form"); err != nil {
		return err
	}

	if err := s.browserCommand(map[string]any{"fill": goal, "mode": "goal-form"}); err != nil {
		return err
	}

	return s.browserCommand(map[string]any{"submit": true, "mode": "goals"})
}

func (s *Suite) browserOpenProjects() error {
	path := "/organisationen/" + s.organizations[s.active] + "/projekte"
	return s.browserClick(`a[href="`+path+`"]`, "projects")
}

func (s *Suite) browserEmptyProjects(name string) error {
	if s.page.Heading != "Projektübersicht" || !strings.Contains(s.page.Text, "Noch keine Projekte") || !strings.Contains(s.page.Text, name) || !s.hasLink("Projekt anlegen") || len(s.page.Projects) != 0 {
		return fmt.Errorf("leere Projektübersicht fehlt: %+v", s.page)
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

func (s *Suite) browserOpenProjectForm() error {
	path := "/organisationen/" + s.organizations[s.active] + "/projekte/neu"
	return s.browserClick(`a[href="`+path+`"]`, "project-form")
}

func (s *Suite) browserOpenProjectFormFor(name string) error {
	if err := s.browserOpenOrganization(name); err != nil {
		return err
	}

	if err := s.browserOpenProjects(); err != nil {
		return err
	}

	return s.browserOpenProjectForm()
}

func (s *Suite) browserProjectFields() error {
	if s.page.Heading != "Projekt anlegen" || !strings.Contains(s.page.Text, "Name") || !strings.Contains(s.page.Text, "Beschreibung") {
		return fmt.Errorf("Projektfelder fehlen: %+v", s.page)
	}

	return nil
}

func (s *Suite) browserCanSelectGoal(goal, organization string) error {
	if s.active != organization {
		return fmt.Errorf("falsche Organisation: %s", s.active)
	}

	for _, option := range s.page.Options {
		if option == goal {
			return nil
		}
	}

	return fmt.Errorf("Zieloption %q fehlt: %+v", goal, s.page.Options)
}

func (s *Suite) browserFillProject(name, description string) error {
	if err := s.browserFillProjectName(name); err != nil {
		return err
	}

	if err := s.browserCommand(map[string]any{"fill": description, "field": "textarea[name=description]", "mode": "project-form"}); err != nil {
		return err
	}

	if s.page.Description != description {
		return fmt.Errorf("Beschreibung nicht gesetzt: %+v", s.page)
	}

	return nil
}

func (s *Suite) browserFillProjectName(name string) error {
	if err := s.browserCommand(map[string]any{"fill": name, "mode": "project-form"}); err != nil {
		return err
	}

	if s.page.Name != name {
		return fmt.Errorf("Projektname nicht gesetzt: %+v", s.page)
	}

	return nil
}

func (s *Suite) browserSelectGoal(goal string) error {
	if err := s.browserCommand(map[string]any{"select": goal, "mode": "project-form"}); err != nil {
		return err
	}

	if s.page.Selected == "" {
		return fmt.Errorf("Projektziel nicht ausgewählt: %+v", s.page)
	}

	return nil
}

func (s *Suite) browserSubmitProject() error {
	mode := "projects"
	if strings.TrimSpace(s.page.Name) == "" || s.page.Selected == "" {
		mode = "error"
	}

	return s.browserCommand(map[string]any{"submit": true, "mode": mode})
}

func (s *Suite) browserSubmitWithoutGoal() error {
	if s.page.Selected != "" {
		return fmt.Errorf("Formular hat bereits Zielauswahl: %+v", s.page)
	}

	return s.browserSubmitProject()
}

func (s *Suite) browserProjectVisible(name, description, organization string) error {
	if s.page.Heading != "Projektübersicht" || !strings.Contains(s.page.Text, organization) || !strings.Contains(s.page.Text, description) {
		return fmt.Errorf("Projektübersicht fehlt: %+v", s.page)
	}

	count := 0
	for _, item := range s.page.Projects {
		if item == name {
			count++
		}
	}

	if count != 1 {
		return fmt.Errorf("Projekt %q %d-mal sichtbar: %+v", name, count, s.page)
	}

	return nil
}

func (s *Suite) browserOpenProjectDetail(name string) error {
	if err := s.browserClick(`main article h2 a`, "project-detail"); err != nil {
		return err
	}

	if s.page.Heading != name {
		return fmt.Errorf("Projektdetail %q fehlt: %+v", name, s.page)
	}

	return nil
}

func (s *Suite) browserDetailGoal(goal string) error {
	if !strings.Contains(s.page.Text, goal) || !strings.Contains(s.page.Text, "Ziel") {
		return fmt.Errorf("Zielbezug im Projektdetail fehlt: %+v", s.page)
	}

	return nil
}

func (s *Suite) browserEmptyTasks() error {
	if !strings.Contains(s.page.Text, "Aufgabenliste") || !strings.Contains(s.page.Text, "Noch keine Aufgaben") {
		return fmt.Errorf("leere Aufgabenliste fehlt: %+v", s.page)
	}

	return nil
}

func (s *Suite) browserRestart() error {
	path := strings.TrimPrefix(s.page.URL, s.baseURL())
	if err := s.restartServer(); err != nil {
		return err
	}

	return s.browserNavigate(path, "project-detail")
}

func (s *Suite) browserProjectAfterRestart(project, description, goal string) error {
	if s.page.Heading != project || !strings.Contains(s.page.Text, description) || !strings.Contains(s.page.Text, goal) {
		return fmt.Errorf("Projektdetail nach Neustart fehlt: %+v", s.page)
	}

	return nil
}

func (s *Suite) browserFieldError(field string) error {
	if s.page.Heading != "Projekt anlegen" {
		return fmt.Errorf("Projektformular fehlt: %+v", s.page)
	}

	message := map[string]string{"Name": s.page.NameError, "Ziel": s.page.GoalError}[field]
	if strings.TrimSpace(message) == "" {
		return fmt.Errorf("Feldhinweis für %q fehlt: %+v", field, s.page)
	}

	return nil
}

func (s *Suite) browserNoProject(organization string) error {
	path := "/organisationen/" + s.organizations[organization] + "/projekte"
	if err := s.browserNavigate(path, "projects"); err != nil {
		return err
	}

	if len(s.page.Projects) != 0 {
		return fmt.Errorf("ungültiges Projekt gespeichert: %+v", s.page.Projects)
	}

	return nil
}

func (s *Suite) browserCannotSelectGoal(goal string) error {
	for _, option := range s.page.Options {
		if option == goal {
			return fmt.Errorf("fremdes Ziel auswählbar: %+v", s.page.Options)
		}
	}

	return nil
}

func (s *Suite) browserFormBelongsTo(organization string) error {
	if !strings.Contains(s.page.Text, organization) || !strings.Contains(s.page.URL, s.organizations[organization]) {
		return fmt.Errorf("Projektformular gehört nicht zu %q: %+v", organization, s.page)
	}

	return nil
}

func (s *Suite) browserOpenSettingsFromEmpty() error {
	if s.page.Heading != "Organisationsübersicht" || !strings.Contains(s.page.Text, "Noch keine Organisation") || !s.hasLink("Settings") {
		return fmt.Errorf("Settings-Link im Leerzustand fehlt: %+v", s.page)
	}

	return s.browserClick(`main a[href="/settings"]`, "settings")
}

func (s *Suite) browserOpenSettingsFromDetail() error {
	if !strings.Contains(s.page.URL, "/organisationen/") || !s.hasLink("Settings") {
		return fmt.Errorf("Settings-Link im Organisationsdetail fehlt: %+v", s.page)
	}

	return s.browserClick(`main a[href="/settings"]`, "settings")
}

func (s *Suite) browserSettingsShown() error {
	if s.page.URL != s.baseURL()+"/settings" || s.page.Heading != "Settings" || !s.hasLink("Codex-Abo") || !s.hasLink("OpenRouter") {
		return fmt.Errorf("globale Settings-Seite fehlt: %+v", s.page)
	}

	return nil
}
