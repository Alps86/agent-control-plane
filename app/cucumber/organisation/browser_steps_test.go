package organisation

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
	sc.Step(`^sehe ich eine leere Organisationsübersicht mit der Aktion "Organisation anlegen"$`, s.browserEmpty)
	sc.Step(`^ich "Organisation anlegen" öffne$`, s.browserOpenForm)
	sc.Step(`^ich öffne das Formular "Organisation anlegen"$`, s.browserOpenForm)
	sc.Step(`^ich erneut "Organisation anlegen" öffne$`, s.browserOpenForm)
	sc.Step(`^sehe ich die Felder "Name" und "Beschreibung"$`, s.browserFields)
	sc.Step(`^ich sehe, dass Arbeits- und Delegationsübergänge zunächst restriktiv sind$`, s.browserRestricted)
	sc.Step(`^ich als Namen "([^"]*)" und als Beschreibung "([^"]*)" eingebe$`, s.browserFill)
	sc.Step(`^ich das Formular speichere$`, s.browserSubmit)
	sc.Step(`^ich lege "([^"]*)" mit der Beschreibung "([^"]*)" an$`, s.browserCreate)
	s.registerBrowserAssertions(sc)
}

func (s *Suite) registerBrowserAssertions(sc *godog.ScenarioContext) {
	sc.Step(`^sehe ich "([^"]*)" und "([^"]*)" in der Organisationsübersicht$`, s.browserListContains)
	sc.Step(`^sehe ich "([^"]*)" und "([^"]*)" erneut$`, s.browserListContains)
	sc.Step(`^ich sehe den restriktiven Standard für Arbeits- und Delegationsübergänge$`, s.browserRestricted)
	sc.Step(`^ich sehe weiterhin den restriktiven Standard für Arbeits- und Delegationsübergänge$`, s.browserRestricted)
	sc.Step(`^ich die Anwendung mit derselben SQLite-Datenbank neu starte und die Seite neu lade$`, s.browserRestart)
	sc.Step(`^sehe ich einen Hinweis am Feld "Name"$`, s.browserNameError)
	sc.Step(`^sehe ich einen Konflikthinweis am Feld "Name"$`, s.browserConflict)
	sc.Step(`^in der Organisationsübersicht erscheint keine neue Organisation$`, s.browserListEmpty)
	sc.Step(`^in der Organisationsübersicht steht "([^"]*)" mit "([^"]*)" genau einmal$`, s.browserListOnce)
	sc.Step(`^ich sehe dort keine Organisation mit der Beschreibung "([^"]*)"$`, s.browserDescriptionAbsent)
}

func (s *Suite) browserFresh() error {
	if err := s.freshServer(); err != nil {
		return err
	}
	if err := s.startBrowser(); err != nil {
		return err
	}
	return s.browserNavigate("/organisationen", "list")
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

func (s *Suite) browserNavigate(path, mode string) error {
	return s.browserCommand(map[string]any{"url": s.baseURL() + path, "mode": mode})
}

func (s *Suite) browserEmpty() error {
	if s.page.Heading != "Organisationsübersicht" {
		return fmt.Errorf("Überschrift %q", s.page.Heading)
	}
	if !strings.Contains(s.page.Text, "Noch keine Organisation") || !strings.Contains(s.page.Text, "Organisation anlegen") {
		return fmt.Errorf("leere Übersicht: %s", s.page.Text)
	}
	if len(s.page.Cards) != 0 {
		return fmt.Errorf("neue Datenbank enthält %d Organisationen", len(s.page.Cards))
	}
	return nil
}

func (s *Suite) browserOpenForm() error {
	if err := s.browserNavigate("/organisationen", "list"); err != nil {
		return err
	}
	return s.browserCommand(map[string]any{"click": `a[href="/organisationen/neu"]`, "mode": "form"})
}

func (s *Suite) browserFields() error {
	if !strings.Contains(s.page.Text, "Name") || !strings.Contains(s.page.Text, "Beschreibung") {
		return fmt.Errorf("Formularfelder fehlen: %s", s.page.Text)
	}
	if s.page.Heading != "Organisation anlegen" {
		return fmt.Errorf("Formularüberschrift %q", s.page.Heading)
	}
	return nil
}

func (s *Suite) browserFill(name, description string) error {
	if err := s.browserCommand(map[string]any{"fill": map[string]string{"name": name, "description": description}, "mode": "form"}); err != nil {
		return err
	}
	if s.page.Name != name || s.page.Description != description {
		return fmt.Errorf("Formularwerte %q/%q", s.page.Name, s.page.Description)
	}
	return nil
}

func (s *Suite) browserSubmit() error {
	mode := "detail"
	if strings.TrimSpace(s.page.Name) == "" || strings.EqualFold(strings.TrimSpace(s.page.Name), "nordstern") && s.page.Description == "Duplikat" {
		mode = "error"
	}
	return s.browserCommand(map[string]any{"submit": true, "mode": mode})
}

func (s *Suite) browserCreate(name, description string) error {
	if err := s.browserOpenForm(); err != nil {
		return err
	}
	if err := s.browserFill(name, description); err != nil {
		return err
	}
	return s.browserSubmit()
}

func (s *Suite) browserListContains(name, description string) error {
	if err := s.browserNavigate("/organisationen", "list"); err != nil {
		return err
	}
	return s.browserCardOnce(name, description)
}

func (s *Suite) browserCardOnce(name, description string) error {
	count := 0
	for _, card := range s.page.Cards {
		if card.Name == name && card.Description == description {
			count++
		}
	}
	if count != 1 {
		return fmt.Errorf("%q/%q: %d Karten: %+v", name, description, count, s.page.Cards)
	}
	return nil
}

func (s *Suite) browserRestricted() error {
	text := s.page.Text
	if !strings.Contains(text, "Arbeits") || !strings.Contains(text, "Delegations") {
		return fmt.Errorf("Policy-Bezeichnungen fehlen: %s", text)
	}
	if !strings.Contains(text, "nicht freigegeben") && strings.Count(text, "keine freigegeben") < 2 {
		return fmt.Errorf("restriktive Policy fehlt: %s", text)
	}
	return nil
}

func (s *Suite) browserRestart() error {
	if err := s.restartServer(); err != nil {
		return err
	}
	return s.browserNavigate("/organisationen", "list")
}

func (s *Suite) browserNameError() error {
	if s.page.Heading != "Organisation anlegen" || s.page.Alert == "" {
		return fmt.Errorf("Feldhinweis fehlt: %+v", s.page)
	}
	return nil
}

func (s *Suite) browserConflict() error {
	if err := s.browserNameError(); err != nil {
		return err
	}
	if !strings.Contains(strings.ToLower(s.page.Alert), "vergeben") {
		return fmt.Errorf("kein Konflikthinweis: %s", s.page.Alert)
	}
	return nil
}

func (s *Suite) browserListEmpty() error {
	if err := s.browserNavigate("/organisationen", "list"); err != nil {
		return err
	}
	return s.browserEmpty()
}

func (s *Suite) browserListOnce(name, description string) error {
	if err := s.browserNavigate("/organisationen", "list"); err != nil {
		return err
	}
	return s.browserCardOnce(name, description)
}

func (s *Suite) browserDescriptionAbsent(description string) error {
	for _, card := range s.page.Cards {
		if card.Description == description {
			return fmt.Errorf("unerlaubte Karte: %+v", card)
		}
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
