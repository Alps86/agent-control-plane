package aufgaben

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

func (s *Suite) registerBrowser(sc *godog.ScenarioContext) {
	s.registerBrowserSetup(sc)
	s.registerBrowserActions(sc)
	s.registerBrowserAssertions(sc)
}

func (s *Suite) registerBrowserSetup(sc *godog.ScenarioContext) {
	sc.Step(`^ich öffne die Anwendung mit der Organisation "([^"]+)" und dem Projekt "([^"]+)" im Browser$`, s.browserProjectSetup)
	sc.Step(`^ich öffne das Formular "Aufgabe anlegen" für das Projekt "([^"]+)" in "([^"]+)" im Browser$`, s.browserFormSetup)
}

func (s *Suite) registerBrowserActions(sc *godog.ScenarioContext) {
	sc.Step(`^ich das Projektdetail von "([^"]+)" öffne$`, s.browserProjectDetail)
	sc.Step(`^ich "Aufgabe anlegen" öffne$`, s.browserOpenForm)
	sc.Step(`^ich den Titel "([^"]+)" und die Beschreibung "([^"]+)" eingebe$`, s.browserFillDetails)
	sc.Step(`^ich die Priorität "([^"]+)" und genau "([^"]+)" als zuständigen Agenten auswähle$`, s.browserSelect)
	sc.Step(`^ich das Aufgabenformular speichere$`, s.browserSubmit)
	sc.Step(`^ich das Aufgabendetail von "([^"]+)" öffne$`, s.browserOpenTask)
	sc.Step(`^ich die Anwendung mit derselben SQLite-Datenbank neu starte und die Seite neu lade$`, s.browserRestart)
	sc.Step(`^ich die Aufgabe "([^"]+)" mit "([^"]+)" als Empfänger vorbelege$`, s.browserPrefill)
	sc.Step(`^ich "([^"]+)" im Aufgabenformular leere$`, s.browserClearField)
	sc.Step(`^ich die Auswahl "Zuständig" öffne$`, s.browserAssigneeOptions)
	sc.Step(`^ich eine Aufgabe mit dem Titel "([^"]+)" und dem Empfänger "([^"]+)" für ein ungültiges Projekt speichere$`, s.browserInvalidProject)
}

func (s *Suite) registerBrowserAssertions(sc *godog.ScenarioContext) {
	sc.Step(`^sehe ich die leere Aufgabenliste und die Aktion "Aufgabe anlegen"$`, s.browserEmptyList)
	sc.Step(`^sehe ich die Felder "Titel", "Beschreibung", "Priorität", "Projekt" und "Zuständig"$`, s.browserFields)
	sc.Step(`^"([^"]+)" ist als Projekt ausgewählt$`, s.browserProjectSelected)
	sc.Step(`^sehe ich "([^"]+)" genau einmal in der Aufgabenliste von "([^"]+)"$`, s.browserTaskOnce)
	sc.Step(`^sehe ich Beschreibung, Priorität "([^"]+)", Projekt "([^"]+)", "([^"]+)" mit "([^"]+)" und Status "([^"]+)"$`, s.browserTaskDetailFields)
	sc.Step(`^sehe ich dieselbe Aufgabe unter derselben Detailadresse mit denselben Angaben$`, s.browserSameTask)
	sc.Step(`^sehe ich einen Hinweis am Aufgabenfeld "([^"]+)"$`, s.browserFieldError)
	sc.Step(`^in der Aufgabenliste von "([^"]+)" erscheint keine neue Aufgabe$`, s.browserNoTask)
	sc.Step(`^kann ich "([^"]+)" genau einmal auswählen$`, s.browserAgentOnce)
	sc.Step(`^ich kann "([^"]+)" und "([^"]+)" nicht auswählen$`, s.browserAgentsAbsent)
	sc.Step(`^ich kann für eine Aufgabe nicht zwei Agenten zugleich auswählen$`, s.browserSingleSelect)
}

func (s *Suite) browserProjectSetup(org, project string) error {
	if err := s.fresh(); err != nil {
		return err
	}
	if err := s.ensureProject(org, project); err != nil {
		return err
	}
	if err := s.startBrowser(); err != nil {
		return err
	}
	return s.browserProjectDetail(project)
}

func (s *Suite) browserFormSetup(project, org string) error {
	if err := s.browserProjectSetup(org, project); err != nil {
		return err
	}
	return s.browserOpenForm()
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
	s.input, s.output = bufio.NewWriter(input), bufio.NewScanner(output)
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

func (s *Suite) browserCommand(command map[string]any) error {
	if err := s.writeBrowserCommand(command); err != nil {
		return err
	}
	return s.readBrowserReply()
}

func (s *Suite) writeBrowserCommand(command map[string]any) error {
	data, err := json.Marshal(command)
	if err != nil {
		return err
	}
	if _, err := s.input.Write(append(data, '\n')); err != nil {
		return err
	}
	if err := s.input.Flush(); err != nil {
		return err
	}
	return nil
}

func (s *Suite) readBrowserReply() error {
	if !s.output.Scan() {
		return fmt.Errorf("Chrome endete: %v", s.output.Err())
	}
	var reply BrowserReply
	if err := json.Unmarshal(s.output.Bytes(), &reply); err != nil {
		return err
	}
	if !reply.OK {
		return fmt.Errorf("Chrome: %s", reply.Error)
	}
	s.page = reply.Page
	return nil
}

func (s *Suite) browserNavigate(path, mode string) error {
	return s.browserCommand(map[string]any{"url": s.url(path), "mode": mode})
}

func (s *Suite) browserProjectDetail(project string) error {
	if project != s.activeProject {
		return fmt.Errorf("Projekt %s statt %s", project, s.activeProject)
	}
	path := "/organisationen/" + s.organizations[s.activeOrg] + "/projekte/" + s.projects[s.key(s.activeOrg, project)]
	return s.browserNavigate(path, "project-detail")
}

func (s *Suite) browserEmptyList() error {
	if !strings.Contains(s.page.Text, "Noch keine Aufgaben") || !strings.Contains(s.page.Text, "Aufgabe anlegen") {
		return fmt.Errorf("Leere Aufgabenliste: %+v", s.page)
	}
	return nil
}

func (s *Suite) browserOpenForm() error {
	path := s.browserProjectPath() + "/aufgaben/neu"
	return s.browserCommand(map[string]any{"click": `a[href="` + path + `"]`, "mode": "task-form"})
}

func (s *Suite) browserProjectPath() string {
	return "/organisationen/" + s.organizations[s.activeOrg] + "/projekte/" + s.projects[s.key(s.activeOrg, s.activeProject)]
}

func (s *Suite) browserFields() error {
	for _, label := range []string{"Titel", "Beschreibung", "Priorität", "Projekt", "Zuständig"} {
		if !strings.Contains(s.page.Text, label) {
			return fmt.Errorf("Feld %s fehlt: %+v", label, s.page)
		}
	}
	return nil
}

func (s *Suite) browserProjectSelected(project string) error {
	if project != s.activeProject || s.page.ProjectName != project || !s.page.ProjectReadOnly {
		return fmt.Errorf("Projekt fehlt: %+v", s.page)
	}
	return nil
}

func (s *Suite) browserFillDetails(title, description string) error {
	if err := s.browserCommand(map[string]any{"fill": title, "field": `input[name=title]`, "mode": "task-form"}); err != nil {
		return err
	}
	return s.browserCommand(map[string]any{"fill": description, "field": `textarea[name=description]`, "mode": "task-form"})
}

func (s *Suite) browserSelect(priority, agent string) error {
	if err := s.browserCommand(map[string]any{"select": s.priority(priority), "field": `select[name=priority]`, "mode": "task-form"}); err != nil {
		return err
	}
	return s.browserCommand(map[string]any{"select": s.agents[s.key(s.activeOrg, agent)], "field": `select[name=assignee_id]`, "mode": "task-form"})
}

func (s *Suite) browserSubmit() error {
	mode := "task-detail"
	if s.page.Title == "" || s.page.Assignee == "" {
		mode = "form-error"
	}
	return s.browserCommand(map[string]any{"submit": true, "mode": mode})
}

func (s *Suite) browserTaskOnce(title, project string) error {
	if err := s.browserProjectDetail(project); err != nil {
		return err
	}
	count := strings.Count(s.page.Text, title)
	if count != 1 {
		return fmt.Errorf("Aufgabe %q %d-mal im Projekt: %+v", title, count, s.page)
	}
	return nil
}

func (s *Suite) browserOpenTask(title string) error {
	if err := s.browserTaskOnce(title, s.activeProject); err != nil {
		return err
	}
	path := s.browserProjectPath() + "/aufgaben/"
	return s.browserCommand(map[string]any{"click": `a[href^="` + path + `"]`, "mode": "task-detail"})
}

func (s *Suite) browserTaskDetailFields(priority, project, agent, kind, status string) error {
	for _, value := range []string{"Navigation und Texte prüfen", priority, project, agent, kind, status} {
		if !strings.Contains(s.page.Text, value) {
			return fmt.Errorf("Detailfeld %s fehlt: %+v", value, s.page)
		}
	}
	return nil
}

func (s *Suite) browserRestart() error {
	path := strings.TrimPrefix(s.page.URL, s.url(""))
	s.detailPath, s.detailText = path, s.page.Text
	if err := s.restart(); err != nil {
		return err
	}
	return s.browserNavigate(path, "task-detail")
}

func (s *Suite) browserSameTask() error {
	path := strings.TrimPrefix(s.page.URL, s.url(""))
	if path != s.detailPath || s.page.Text != s.detailText || s.page.Heading != "Startseite prüfen" {
		return fmt.Errorf("Aufgabe nach Neustart: %+v", s.page)
	}
	return nil
}

func (s *Suite) browserPrefill(title, agent string) error {
	if err := s.browserNavigate(s.browserProjectPath()+"/aufgaben/neu", "task-form"); err != nil {
		return err
	}
	if err := s.browserCommand(map[string]any{"fill": title, "field": `input[name=title]`, "mode": "task-form"}); err != nil {
		return err
	}
	return s.browserCommand(map[string]any{"select": s.agents[s.key(s.activeOrg, agent)], "field": `select[name=assignee_id]`, "mode": "task-form"})
}

func (s *Suite) browserClearField(field string) error {
	selector := `input[name=title]`
	if field == "Zuständig" {
		selector = `select[name=assignee_id]`
	}
	return s.browserCommand(map[string]any{"fill": "", "field": selector, "mode": "task-form"})
}

func (s *Suite) browserFieldError(field string) error {
	message := s.page.TitleError
	if field == "Zuständig" {
		message = s.page.AssigneeError
	}
	if field == "Projekt" {
		message = s.page.ProjectError
	}
	if message == "" {
		return fmt.Errorf("Hinweis am Feld %s fehlt: %+v", field, s.page)
	}
	return nil
}

func (s *Suite) browserNoTask(project string) error {
	if err := s.browserProjectDetail(project); err != nil {
		return err
	}
	if strings.Contains(s.page.Text, "Startseite prüfen") {
		return fmt.Errorf("Unerwartete Aufgabe: %+v", s.page)
	}
	return nil
}

func (s *Suite) browserAssigneeOptions() error {
	if err := s.browserNavigate(s.browserProjectPath()+"/aufgaben/neu", "task-form"); err != nil {
		return err
	}
	if s.page.Heading != "Aufgabe anlegen" {
		return fmt.Errorf("Formular fehlt: %+v", s.page)
	}
	return nil
}

func (s *Suite) browserAgentOnce(agent string) error {
	count := 0
	for _, option := range s.page.AssigneeOptions {
		if strings.Contains(option, agent) {
			count++
		}
	}
	if count != 1 {
		return fmt.Errorf("Agent %s %d-mal: %+v", agent, count, s.page)
	}
	return nil
}

func (s *Suite) browserAgentsAbsent(first, second string) error {
	for _, option := range s.page.AssigneeOptions {
		if strings.Contains(option, first) || strings.Contains(option, second) {
			return fmt.Errorf("Unzulässiger Agent: %s", option)
		}
	}
	return nil
}

func (s *Suite) browserSingleSelect() error {
	if s.page.AssigneeSelectCount != 1 || s.page.AssigneeMultiple {
		return fmt.Errorf("Mehrfachauswahl unklar: %+v", s.page)
	}
	return nil
}

func (s *Suite) browserInvalidProject(title, agent string) error {
	if err := s.browserPrefill(title, agent); err != nil {
		return err
	}
	if err := s.browserCommand(map[string]any{"fill": "Fremdprojekt", "field": `input[name=project_name]`, "mode": "task-form"}); err != nil {
		return err
	}
	return s.browserCommand(map[string]any{"submit": true, "mode": "form-error"})
}
