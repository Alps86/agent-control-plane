package codexagent

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
	sc.Step(`^ich öffne die Anwendung für die Organisation "([^"]+)" im Browser$`, s.browserOrganization)
	sc.Step(`^ich "Agent anlegen" öffne$`, s.browserForm)
	sc.Step(`^ich "Codex CLI" als eigene Ausführungsart auswählen$`, s.browserChoice)
	sc.Step(`^kann ich "Codex CLI" als eigene Ausführungsart auswählen$`, s.browserChoice)
	sc.Step(`^ich "([^"]+)" eingebe, "Codex CLI" auswähle und speichere$`, s.browserSave)
	sc.Step(`^"([^"]+)" in der Agentenübersicht von "([^"]+)" mit "Codex CLI"$`, s.browserListed)
	sc.Step(`^erscheint "([^"]+)" in der Agentenübersicht von "([^"]+)" mit "Codex CLI"$`, s.browserListed)
	sc.Step(`^ich "([^"]+)" öffne$`, s.browserDetail)
	sc.Step(`^ich "([^"]+)" als Organisationskontext$`, s.browserOrganizationText)
	sc.Step(`^sehe ich "([^"]+)" als Organisationskontext$`, s.browserOrganizationText)
	sc.Step(`^ich seine freigegebenen benannten Fachfähigkeiten oder "Keine Fachfähigkeiten freigegeben"$`, s.browserCapabilities)
	sc.Step(`^ich sehe seine freigegebenen benannten Fachfähigkeiten oder "Keine Fachfähigkeiten freigegeben"$`, s.browserCapabilities)
	sc.Step(`^ich den konkreten Grund, warum er noch nicht startbereit ist$`, s.browserReason)
	sc.Step(`^ich sehe den konkreten Grund, warum er noch nicht startbereit ist$`, s.browserReason)
	sc.Step(`^ich öffne den Codex-CLI-Agenten "([^"]+)" mit fehlender erforderlicher Konfiguration im Browser$`, s.browserMissingConfig)
	sc.Step(`^ich "Nicht startbereit"$`, s.browserNotReady)
	sc.Step(`^sehe ich "Nicht startbereit"$`, s.browserNotReady)
	sc.Step(`^ich sehe, welche Konfiguration fehlt$`, s.browserConfigReason)
	sc.Step(`^ich die Seite neu lade$`, s.browserReload)
	sc.Step(`^ich weiterhin "Nicht startbereit" mit dem fehlenden Konfigurationsgrund$`, s.browserConfigStillMissing)
	sc.Step(`^sehe ich weiterhin "Nicht startbereit" mit dem fehlenden Konfigurationsgrund$`, s.browserConfigStillMissing)
	sc.Step(`^ich öffne den Codex-CLI-Agenten "([^"]+)" im Browser$`, s.browserMissingConfig)
	sc.Step(`^die Sperre direkter Shell-, Terminal-, Prozess-, Interpreter-, HTTP- und Dateisystemwege ist für sein Profil nicht nachgewiesen$`, s.browserNoEnforcement)
	sc.Step(`^ich "Nicht startbereit" und den fehlenden Durchsetzungsnachweis$`, s.browserEnforcementReason)
	sc.Step(`^sehe ich "Nicht startbereit" und den fehlenden Durchsetzungsnachweis$`, s.browserEnforcementReason)
	sc.Step(`^die Oberfläche zeigt keine erfolgreiche Ausführung oder einen abgeschlossenen Lauf$`, s.browserNoRun)
}
func (s *Suite) browserOrganization(name string) error {
	if err := s.fresh(); err != nil {
		return err
	}
	if err := s.createOrganization(name); err != nil {
		return err
	}
	if err := s.startBrowser(); err != nil {
		return err
	}
	return s.browserNavigate("/organisationen/" + s.orgID + "/agenten")
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
	if !s.output.Scan() {
		return fmt.Errorf("Browser endete: %v", s.output.Err())
	}
	var reply BrowserReply
	if err := json.Unmarshal(s.output.Bytes(), &reply); err != nil {
		return err
	}
	if !reply.OK {
		return fmt.Errorf("Browser: %s", reply.Error)
	}
	s.page = reply.Page
	return nil
}
func (s *Suite) browserNavigate(path string) error {
	return s.browserCommand(map[string]any{"url": s.url(path)})
}
func (s *Suite) browserForm() error {
	return s.browserCommand(map[string]any{"click": "a[href=\"/organisationen/" + s.orgID + "/agenten/neu\"]"})
}
func (s *Suite) browserChoice() error {
	if s.page.Heading != "Agent anlegen" || !strings.Contains(s.page.Text, "Codex CLI") {
		return fmt.Errorf("CLI-Vorlage nicht angeboten: %+v", s.page)
	}
	if err := s.browserCommand(map[string]any{"clickTemplate": "codex-cli"}); err != nil {
		return err
	}
	if s.page.ExecutionKind != "codex_cli" || s.page.TemplateID != "codex-cli" {
		return fmt.Errorf("CLI-Auswahl fehlt: %+v", s.page)
	}
	return nil
}
func (s *Suite) browserSave(name string) error {
	s.agentName = name
	if err := s.browserCommand(map[string]any{"fill": name}); err != nil {
		return err
	}
	if s.page.Name != name {
		return fmt.Errorf("Name nicht gefüllt: %+v", s.page)
	}
	if err := s.browserCommand(map[string]any{"submit": true}); err != nil {
		return err
	}
	if !strings.Contains(s.page.Text, name) {
		return fmt.Errorf("nach Speichern kein Agent: %+v", s.page)
	}
	return nil
}
func (s *Suite) browserListed(name, organization string) error {
	if err := s.browserNavigate("/organisationen/" + s.orgID + "/agenten"); err != nil {
		return err
	}
	if organization == "" || !strings.Contains(s.page.URL, "/organisationen/"+s.orgID+"/agenten") || !strings.Contains(s.page.Text, name) || !strings.Contains(s.page.Text, "Codex CLI") {
		return fmt.Errorf("Agentenliste unvollständig: %+v", s.page)
	}
	return nil
}
func (s *Suite) browserDetail(name string) error {
	if name != s.agentName {
		return fmt.Errorf("Agent %q unbekannt", name)
	}
	if err := s.browserCommand(map[string]any{"clickText": name}); err != nil {
		return err
	}
	if s.page.Heading != name {
		return fmt.Errorf("Agentendetail fehlt: %+v", s.page)
	}
	return nil
}
func (s *Suite) browserOrganizationText(name string) error {
	if !strings.Contains(s.page.Text, name) {
		return fmt.Errorf("Organisationskontext fehlt: %+v", s.page)
	}
	return nil
}
func (s *Suite) browserCapabilities() error {
	if !strings.Contains(s.page.Text, "Keine Fachfähigkeiten freigegeben") {
		return fmt.Errorf("Fachfähigkeiten fehlen: %+v", s.page)
	}
	return nil
}
func (s *Suite) browserReason() error {
	if !strings.Contains(s.page.Text, "Nicht startbereit") || !strings.Contains(s.page.Text, "Konfiguration") {
		return fmt.Errorf("Sperrgrund fehlt: %+v", s.page)
	}
	return nil
}
func (s *Suite) browserMissingConfig(name string) error {
	if err := s.browserOrganization("Nordstern"); err != nil {
		return err
	}
	if err := s.createAgent("Nordstern", name); err != nil {
		return err
	}
	if err := s.agentCreated(); err != nil {
		return err
	}
	return s.browserNavigate("/organisationen/" + s.orgID + "/agenten/" + s.agentID)
}
func (s *Suite) browserNotReady() error {
	if !strings.Contains(s.page.Text, "Nicht startbereit") {
		return fmt.Errorf("fälschlich bereit: %+v", s.page)
	}
	return nil
}
func (s *Suite) browserConfigReason() error {
	if !strings.Contains(s.page.Text, "Konfiguration") {
		return fmt.Errorf("Konfigurationsgrund fehlt: %+v", s.page)
	}
	return nil
}
func (s *Suite) browserReload() error {
	return s.browserNavigate("/organisationen/" + s.orgID + "/agenten/" + s.agentID)
}
func (s *Suite) browserConfigStillMissing() error {
	if err := s.browserNotReady(); err != nil {
		return err
	}
	return s.browserConfigReason()
}
func (s *Suite) browserNoEnforcement() error {
	if err := s.noEnforcement(); err != nil {
		return err
	}
	return s.browserReload()
}
func (s *Suite) browserEnforcementReason() error {
	if err := s.browserNotReady(); err != nil {
		return err
	}
	if !strings.Contains(s.page.Text, "nicht nachgewiesen") {
		return fmt.Errorf("Durchsetzungsnachweis fehlt: %+v", s.page)
	}
	return nil
}
func (s *Suite) browserNoRun() error {
	if strings.Contains(s.page.Text, "Lauf abgeschlossen") || strings.Contains(s.page.Text, "Erfolgreich ausgeführt") {
		return fmt.Errorf("unzulässiger Lauf: %+v", s.page)
	}
	return s.startDenied()
}
func (s *Suite) stopBrowser() {
	if s.browser == nil || s.browser.Process == nil {
		return
	}
	_ = syscall.Kill(-s.browser.Process.Pid, syscall.SIGKILL)
	_ = s.browser.Wait()
	s.browser = nil
}
