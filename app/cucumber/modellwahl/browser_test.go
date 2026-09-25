package modellwahl

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
	sc.Step(`^ich Miras Modellwahl im echten Browser öffne$`, s.browserOpen)
	sc.Step(`^sehe ich "Codex-Abo" als vorausgewählten Anbieter$`, s.browserCodex)
	sc.Step(`^ich sehe "Eino" als Ausführungsart$`, s.browserEino)
	sc.Step(`^die Seite zeigt nur nicht geheime Verbindungsreferenzen$`, s.browserReferencesOnly)
	sc.Step(`^die Seite zeigt keine Zugangsdaten$`, s.browserNoSecrets)
	sc.Step(`^ich im echten Browser OpenRouter und "openai/gpt-4" für "Mira" auswähle und speichere$`, s.browserChoose)
	sc.Step(`^sehe ich "OpenRouter" und "openai/gpt-4" in Miras Agentenansicht$`, s.browserSelected)
	sc.Step(`^ich Miras Modellfähigkeiten im echten Browser öffne$`, s.browserOpenCapabilities)
	sc.Step(`^sehe ich Modellquelle, Stand und Prüfstatus für "openai/gpt-4"$`, s.browserEvidence)
	sc.Step(`^ich sehe den Textbeleg mit Quelle, Prüfzeit und Status$`, s.browserTextEvidence)
	sc.Step(`^ich sehe Text, Tools, Audioeingabe und Audioausgabe als getrennte Fähigkeiten$`, s.browserCapabilities)
	sc.Step(`^bidirektionale Echtzeit ist als "Nicht nachgewiesen" markiert$`, s.browserRealtime)
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
	if s.browser == nil {
		return
	}
	_ = s.browser.Process.Kill()
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
	return s.browserReply()
}

func (s *Suite) browserReply() error {
	if !s.output.Scan() {
		return fmt.Errorf("Browser endete: %v", s.output.Err())
	}
	reply := BrowserReply{}
	if err := json.Unmarshal(s.output.Bytes(), &reply); err != nil {
		return err
	}
	if !reply.OK {
		return fmt.Errorf("Browser: %s", reply.Error)
	}
	s.page = reply.Page
	return nil
}

func (s *Suite) browserOpen() error {
	if err := s.startBrowser(); err != nil {
		return err
	}
	return s.browserCommand(map[string]string{"url": s.url(s.pagePath())})
}
func (s *Suite) browserChoose() error {
	if err := s.browserOpen(); err != nil {
		return err
	}
	return s.browserCommand(map[string]string{"provider": "openrouter", "connection": "openrouter-central", "model": "openai/gpt-4", "submit": "1"})
}
func (s *Suite) browserOpenCapabilities() error {
	if err := s.browserOpen(); err != nil {
		return err
	}
	return s.browserCommand(map[string]string{"expand": "openai/gpt-4"})
}
func (s *Suite) pageContains(parts ...string) error {
	for _, part := range parts {
		if !strings.Contains(strings.ToLower(s.page.Text), strings.ToLower(part)) {
			return fmt.Errorf("Browserseite ohne %q: %s", part, s.page.Text)
		}
	}
	return nil
}
func (s *Suite) browserCodex() error {
	if !strings.Contains(s.page.Selected, "Codex-Abo") {
		return fmt.Errorf("vorausgewählt %q", s.page.Selected)
	}
	return s.pageContains("Codex-Abo")
}
func (s *Suite) browserEino() error { return s.pageContains("Eino") }
func (s *Suite) browserSelected() error {
	if !strings.Contains(s.page.Selected, "OpenRouter") || !strings.Contains(s.page.Selected, "openai/gpt-4") {
		return fmt.Errorf("aktuelle Browserauswahl falsch: %s", s.page.Selected)
	}
	return nil
}
func (s *Suite) browserEvidence() error {
	item := s.page.ModelEvidence
	if !s.page.Expanded || item.Source != "https://openrouter.ai/docs/api/api-reference/models/get-models" || item.ObservedAt != "2026-09-24T23:55:13Z" || item.Status != "Nicht nachgewiesen" {
		return fmt.Errorf("Modellnachweis im Browser unvollständig: %+v", item)
	}
	return nil
}
func (s *Suite) browserTextEvidence() error {
	return s.browserCapabilityEvidence("Text")
}
func (s *Suite) browserCapabilityEvidence(name string) error {
	item, found := s.page.Capabilities[name]
	if !found || !strings.Contains(item.Text, "Nicht nachgewiesen") || !strings.Contains(item.Text, "Quelle: https://openrouter.ai/docs/api/api-reference/models/get-models") || item.CheckedAt != "2026-09-24T23:55:13Z" {
		return fmt.Errorf("Fähigkeitsreihe %q unvollständig: %+v", name, item)
	}
	return nil
}
func (s *Suite) browserCapabilities() error {
	for _, name := range []string{"Text", "Tools", "Audioeingabe", "Audioausgabe"} {
		if err := s.browserCapabilityEvidence(name); err != nil {
			return err
		}
	}
	return nil
}
func (s *Suite) browserRealtime() error {
	return s.browserCapabilityEvidence("Bidirektionale Echtzeit")
}
func (s *Suite) browserReferencesOnly() error {
	if !strings.Contains(s.page.Selected, "codex-chatgpt") {
		return fmt.Errorf("öffentliche Codex-Referenz fehlt")
	}
	for _, marker := range []string{`type="password"`, `name="access_token"`, `name="refresh_token"`} {
		if strings.Contains(s.page.HTML, marker) {
			return fmt.Errorf("Geheimnisfeld %s in Modellwahl", marker)
		}
	}
	return nil
}
func (s *Suite) browserNoSecrets() error {
	if !s.secretStored {
		return fmt.Errorf("Testschlüssel wurde nicht öffentlich hinterlegt; Browser-Redaction unbelegt")
	}
	if strings.Contains(s.page.HTML, s.secret) {
		return fmt.Errorf("Browser-DOM enthält Zugangsdaten")
	}
	return nil
}
