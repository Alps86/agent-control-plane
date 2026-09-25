package kommentare

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	appkommentar "agentcontrolplane/app/internal/app/kommentar"
	domainkommentar "agentcontrolplane/app/internal/domain/kommentar"
	portkommentar "agentcontrolplane/app/internal/port/kommentar"
	"github.com/cucumber/godog"
)

func (s *Suite) registerBrowser(sc *godog.ScenarioContext) {
	s.registerBrowser1(sc)
	s.registerBrowser2(sc)
}

func (s *Suite) registerBrowser1(sc *godog.ScenarioContext) {
	sc.Step(`^ich öffne die Kommentaransicht "/kommentare" von "([^"]+)" im Projekt "([^"]+)" der Organisation "([^"]+)" in einem echten Browser$`, s.browserSetup)
	sc.Step(`^ich im Aufgabenthread "([^"]+)" eingebe und absende$`, s.browserSubmit)
	sc.Step(`^sehe ich "([^"]+)" mit Quelle "([^"]+)" und Zeitpunkt im Thread$`, s.browserShows)
	sc.Step(`^sehe ich "([^"]+)" mit Quelle "([^"]+)" und Zeitpunkt$`, s.browserShows)
	sc.Step(`^die Kommentaransicht bleibt derselben Aufgabenkennung zugeordnet$`, s.browserSameTask)
	sc.Step(`^ich die Anwendung mit derselben SQLite-Datenbank neu starte und die Kommentaransicht neu lade$`, s.browserRestart)
	sc.Step(`^sehe ich denselben Kommentar mit derselben Herkunft und demselben Zeitpunkt genau einmal$`, s.browserSameComment)
	sc.Step(`^"([^"]+)" hat als berechtigter Agent "([^"]+)" mit Artefaktkennung "([^"]+)" an "([^"]+)" kommentiert$`, s.browserArtifactPrecomment)
	sc.Step(`^ich die Kommentaransicht "/kommentare" von "([^"]+)" im Browser öffne$`, s.browserOpenComment)
	sc.Step(`^ich sehe Kennung, Anzeigenamen "([^"]+)" und den internen Link an genau diesem Kommentar$`, s.browserArtifactLink)
}

func (s *Suite) registerBrowser2(sc *godog.ScenarioContext) {
	sc.Step(`^die Anzeige bietet keinen Datei- oder Downloadlink an$`, s.browserNoDownload)
	sc.Step(`^ich öffne die Kommentaransicht "/kommentare" von "([^"]+)" im Browser$`, s.browserOpenComment)
	sc.Step(`^ich nur Leerzeichen im Kommentarfeld eingebe und absende$`, s.browserEmptySubmit)
	sc.Step(`^sehe ich einen Hinweis am Kommentarfeld$`, s.browserContentError)
	sc.Step(`^im Aufgabenthread erscheint kein neuer Kommentar$`, s.browserNoNewComment)
	sc.Step(`^ich betrachte die Organisation "([^"]+)" im Browser$`, s.browserOrganization)
	sc.Step(`^"([^"]+)" gehört zur Organisation "([^"]+)"$`, s.browserForeignTask)
	sc.Step(`^ich die direkte Kommentaradresse "/kommentare" von "([^"]+)" im Browser öffne$`, s.browserForeignOpen)
	sc.Step(`^sehe ich weder Aufgabeninhalt noch Kommentare oder Artefaktbezüge von "([^"]+)"$`, s.browserForeignHidden)
}

func (s *Suite) browserPath(org, project, title string) string {
	return strings.TrimPrefix(s.commentPath(org, project, title), "/api")
}

func (s *Suite) browserSetup(title, project, org string) error {
	if err := s.ensureTask(org, project, title); err != nil {
		return err
	}
	return s.browserOpenComment(title)
}

func (s *Suite) browserOpenComment(title string) error {
	if err := s.ensureTask("Nordstern", "Website", title); err != nil {
		return err
	}
	if err := s.startBrowser(); err != nil {
		return err
	}
	return s.browserNavigate(s.browserPath("Nordstern", "Website", title))
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
	return s.readBrowserReply()
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

func (s *Suite) browserNavigate(path string) error {
	return s.browserCommand(map[string]any{"url": s.url(path), "mode": "comment"})
}

func (s *Suite) browserSubmit(content string) error {
	if err := s.browserCommand(map[string]any{"fill": content, "field": `textarea[name=content]`, "mode": "comment"}); err != nil {
		return err
	}
	return s.browserCommand(map[string]any{"submit": true, "mode": "comment"})
}

func (s *Suite) browserShows(content, source string) error {
	if !strings.Contains(s.page.Text, content) || !strings.Contains(s.page.Text, source) {
		return fmt.Errorf("Kommentar nicht sichtbar: %+v", s.page)
	}
	if len(s.page.Timestamps) != 1 || s.page.Timestamps[0].Value == "" || s.page.Timestamps[0].Value != s.page.Timestamps[0].Text {
		return fmt.Errorf("Zeitstempel fehlt oder weicht ab: %+v", s.page.Timestamps)
	}
	return nil
}

func (s *Suite) browserSameTask() error {
	if !strings.Contains(s.page.URL, s.taskID("Nordstern", "Texte prüfen")) {
		return fmt.Errorf("Falscher Task: %s", s.page.URL)
	}
	return nil
}

func (s *Suite) browserRestart() error {
	path := strings.TrimPrefix(s.page.URL, s.url(""))
	if len(s.page.Timestamps) != 1 {
		return fmt.Errorf("Zeitstempel vor Neustart fehlt: %+v", s.page.Timestamps)
	}
	s.savedTimestamp = s.page.Timestamps[0].Value
	if err := s.restart(); err != nil {
		return err
	}
	return s.browserNavigate(path)
}

func (s *Suite) browserSameComment() error {
	if s.page.CommentCount != 1 || strings.Count(s.page.Text, "Fachprüfung folgt") != 1 {
		return fmt.Errorf("Kommentar nicht genau einmal sichtbar: %+v", s.page)
	}
	if len(s.page.Timestamps) != 1 || s.page.Timestamps[0].Value != s.savedTimestamp {
		return fmt.Errorf("Zeitstempel nach Neustart geändert: %+v", s.page.Timestamps)
	}
	return s.browserShows("Fachprüfung folgt", "Betreiberin")
}

func (s *Suite) browserArtifactPrecomment(agent, content, id, title string) error {
	if err := s.ensureTask("Nordstern", "Website", title); err != nil {
		return err
	}
	if err := s.setupScopedAgent(agent, "Nordstern", "Website"); err != nil {
		return err
	}
	if err := s.agentService("Nordstern", agent); err != nil {
		return err
	}
	ref := &domainkommentar.Reference{Type: "artifact", ID: id, OrganizationID: s.orgID("Nordstern")}
	value, err := s.service.CreateForAgent(context.Background(), s.orgID("Nordstern"), s.taskID("Nordstern", title), appkommentar.ActionTaskComment, appkommentar.CreateInput{Content: content, Reference: ref})
	if err != nil {
		return err
	}
	s.lastComment, s.commentID = s.fromDomain(value), value.ID
	return nil
}

func (s *Suite) browserArtifactLink(name string) error {
	if !strings.Contains(s.page.Text, "entwurf-7") || !strings.Contains(s.page.Text, name) {
		return fmt.Errorf("Artefaktmetadaten fehlen: %+v", s.page)
	}
	for _, link := range s.page.Links {
		if link.Href == "/artefakte/entwurf-7" && link.Text == name {
			return nil
		}
	}
	return fmt.Errorf("Interner Artefaktlink fehlt: %+v", s.page.Links)
}

func (s *Suite) browserNoDownload() error {
	for _, link := range s.page.Links {
		if strings.Contains(strings.ToLower(link.Href), "download") || strings.Contains(link.Href, "/dateien/") {
			return fmt.Errorf("Datei-/Downloadlink angeboten: %+v", link)
		}
	}
	return nil
}

func (s *Suite) browserEmptySubmit() error { return s.browserSubmit("   ") }
func (s *Suite) browserContentError() error {
	if s.page.Alert == "" {
		return fmt.Errorf("Kommentarfehler fehlt: %+v", s.page)
	}
	return nil
}

func (s *Suite) browserNoNewComment() error           { return s.threadEmpty() }
func (s *Suite) browserOrganization(org string) error { return s.ensureOrganization(org) }
func (s *Suite) browserForeignTask(title, org string) error {
	if err := s.ensureTask(org, "Fremd", title); err != nil {
		return err
	}
	if err := s.setupScopedAgent("Fremd", org, "Fremd"); err != nil {
		return err
	}
	s.resolver = &ArtifactFixture{metadata: portkommentar.ArtifactMetadata{ID: "foreign-secret-42", OrganizationID: s.orgID(org), TaskID: s.taskID(org, title), DisplayName: "Geheimes Dokument", Link: "/artefakte/foreign-secret-42"}}
	if err := s.agentService(org, "Fremd"); err != nil {
		return err
	}
	ref := &domainkommentar.Reference{Type: "artifact", ID: "foreign-secret-42", OrganizationID: s.orgID(org)}
	_, err := s.service.CreateForAgent(context.Background(), s.orgID(org), s.taskID(org, title), appkommentar.ActionTaskComment, appkommentar.CreateInput{Content: "Geheime Suedsternnotiz", Reference: ref})
	return err
}
func (s *Suite) browserForeignOpen(title string) error {
	if err := s.startBrowser(); err != nil {
		return err
	}
	path := "/organisationen/" + s.orgID("Nordstern") + "/projekte/" + s.projects[s.key("Suedstern", "Fremd")] + "/aufgaben/" + s.taskID("Suedstern", title) + "/kommentare"
	return s.browserNavigate(path)
}
func (s *Suite) browserForeignHidden(title string) error {
	if strings.Contains(s.page.Text, title) || strings.Contains(s.page.Text, "Geheime Suedsternnotiz") || strings.Contains(s.page.Text, "foreign-secret-42") || strings.Contains(s.page.Text, "Geheimes Dokument") {
		return fmt.Errorf("Fremde Daten sichtbar: %+v", s.page)
	}
	return nil
}
