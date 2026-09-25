package projektarchiv

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/cucumber/godog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

func (s *Suite) registerBrowser(sc *godog.ScenarioContext) {
	sc.Step(`^ich öffne die Anwendung mit "([^"]+)" und dem Projekt "([^"]+)" im Browser$`, s.browserSetup)
	sc.Step(`^ich die Projektübersicht von "([^"]+)" öffne$`, s.browserList)
	sc.Step(`^ich "([^"]+)" über die sichtbare Aktion archiviere$`, s.browserArchive)
	sc.Step(`^sehe ich "([^"]+)" nicht mehr in der aktiven Projektliste$`, s.browserAbsent)
	sc.Step(`^ich das Projektarchiv von "([^"]+)" öffne$`, s.browserOpenArchive)
	sc.Step(`^sehe ich "([^"]+)" mit dem Status "([^"]+)"$`, s.browserArchived)
	sc.Step(`^ich kann die bestehende Aufgabe "([^"]+)" weiterhin öffnen$`, s.browserTask)
	sc.Step(`^ich "([^"]+)" über die sichtbare Aktion wiederherstelle$`, s.browserRestore)
	sc.Step(`^sehe ich "([^"]+)" erneut in der aktiven Projektliste$`, s.browserActive)
	sc.Step(`^ich die direkte Browseradresse für eine neue Aufgabe in "([^"]+)" öffne$`, s.browserTaskForm)
	sc.Step(`^sehe ich einen verständlichen Archivhinweis statt eines speicherbaren Aufgabenformulars$`, s.browserBlockedForm)
	sc.Step(`^im Projekt "([^"]+)" entsteht keine neue Aufgabe$`, s.browserNoTask)
}

func (s *Suite) browserSetup(org, project string) error {
	if err := s.projectSetup(org, project); err != nil {
		return err
	}
	if err := s.startBrowser(); err != nil {
		return err
	}
	return s.browserList(org)
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
	return s.browserPipes()
}

func (s *Suite) browserPipes() error {
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

func (s *Suite) browserCommand(command map[string]string) error {
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
	if !s.output.Scan() {
		return fmt.Errorf("Chrome endete: %v", s.output.Err())
	}
	return s.browserReply()
}

func (s *Suite) browserReply() error {
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

func (s *Suite) browserNavigate(path, heading string) error {
	return s.browserCommand(map[string]string{"url": s.url(path), "expected": heading})
}

func (s *Suite) browserList(org string) error {
	s.activeOrg = org
	return s.browserNavigate("/organisationen/"+s.organizations[org]+"/projekte", "Projektübersicht")
}

func (s *Suite) browserArchive(project string) error {
	detail := "/organisationen/" + s.organizations[s.activeOrg] + "/projekte/" + s.projects[s.key(s.activeOrg, project)]
	if !s.hasLink(detail) {
		return fmt.Errorf("Projektlink fehlt: %+v", s.page)
	}

	if err := s.browserCommand(map[string]string{"click": `a[href="` + detail + `"]`, "expected": project}); err != nil {
		return err
	}

	path := "/organisationen/" + s.organizations[s.activeOrg] + "/projekte/" + s.projects[s.key(s.activeOrg, project)] + "/archivieren"
	if !s.hasForm(path) {
		return fmt.Errorf("sichtbare Archivaktion fehlt: %+v", s.page)
	}

	return s.browserCommand(map[string]string{"submit": `form[action="` + path + `"]`, "expected": "Projektübersicht"})
}

func (s *Suite) hasForm(path string) bool {
	for _, form := range s.page.Forms {
		if form == path {
			return true
		}
	}

	return false
}

func (s *Suite) browserAbsent(project string) error {
	if strings.Contains(s.page.Text, project) {
		return fmt.Errorf("Aktive Liste enthält %s: %+v", project, s.page)
	}
	return nil
}

func (s *Suite) browserOpenArchive(org string) error {
	path := "/organisationen/" + s.organizations[org] + "/projekte/archiv"
	if !s.hasLink(path) {
		return fmt.Errorf("Archivlink fehlt: %+v", s.page)
	}
	return s.browserCommand(map[string]string{"click": `a[href="` + path + `"]`, "expected": "Projektarchiv"})
}

func (s *Suite) hasLink(path string) bool {
	for _, link := range s.page.Links {
		if link == path {
			return true
		}
	}
	return false
}

func (s *Suite) browserArchived(project, status string) error {
	if !strings.Contains(s.page.Text, project) || !strings.Contains(s.page.Text, status) {
		return fmt.Errorf("Archivansicht: %+v", s.page)
	}
	return nil
}

func (s *Suite) browserTask(title string) error {
	projectPath := "/organisationen/" + s.organizations[s.activeOrg] + "/projekte/" + s.projects[s.key(s.activeOrg, s.activeProject)]
	if err := s.browserNavigate(projectPath, s.activeProject); err != nil {
		return err
	}
	path := projectPath + "/aufgaben/" + s.tasks[s.key(s.activeOrg, title)]
	if !s.hasLink(path) {
		return fmt.Errorf("Aufgabenlink fehlt: %+v", s.page)
	}
	return s.browserCommand(map[string]string{"click": `a[href="` + path + `"]`, "expected": title})
}

func (s *Suite) browserRestore(project string) error {
	archivePath := "/organisationen/" + s.organizations[s.activeOrg] + "/projekte/archiv"
	if err := s.browserNavigate(archivePath, "Projektarchiv"); err != nil {
		return err
	}
	path := "/organisationen/" + s.organizations[s.activeOrg] + "/projekte/" + s.projects[s.key(s.activeOrg, project)] + "/restaurieren"
	return s.browserCommand(map[string]string{"submit": `form[action="` + path + `"]`, "expected": "Projektübersicht"})
}

func (s *Suite) browserActive(project string) error {
	if !strings.Contains(s.page.Text, project) {
		return fmt.Errorf("Projekt nicht wieder aktiv: %+v", s.page)
	}
	return nil
}

func (s *Suite) browserTaskForm(project string) error {
	path := "/organisationen/" + s.organizations[s.activeOrg] + "/projekte/" + s.projects[s.key(s.activeOrg, project)] + "/aufgaben/neu"
	return s.browserNavigate(path, "*")
}

func (s *Suite) browserBlockedForm() error {
	if !strings.Contains(strings.ToLower(s.page.Text), "archiv") {
		return fmt.Errorf("Archivhinweis fehlt: %+v", s.page)
	}
	for _, form := range s.page.Forms {
		if strings.Contains(form, "/aufgaben") {
			return fmt.Errorf("Aufgabenformular noch offen: %+v", s.page)
		}
	}
	return nil
}

func (s *Suite) browserNoTask(project string) error {
	s.rejectedTitle = "Verbotener Auftrag"
	return s.noNewTask(project)
}
