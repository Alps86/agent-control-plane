package codexprofil

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
	sc.Step(`^ich öffne den Codex-CLI-Agenten "Kai" in "Nordstern" im Browser$`, s.browserAgent)
	sc.Step(`^ich sein Arbeitsprofil öffne$`, s.browserOpenProfile)
	sc.Step(`^sehe ich einen serverseitig abgeleiteten privaten Arbeitsbereich ohne Pfadeingabe$`, s.browserDerived)
	sc.Step(`^die Schalter "Arbeitsbereich" und "Schreiben" sind deaktiviert$`, s.browserBothDisabled)
	sc.Step(`^der konkrete CLI-Durchsetzungsgrund ist als "Nicht startbereit" sichtbar$`, s.browserUnverified)
	sc.Step(`^ich sein Arbeitsprofil öffne und nur "Arbeitsbereich" aktiviere und speichere$`, s.browserEnableWorkspace)
	sc.Step(`^sehe ich "Arbeitsbereich" aktiviert und "Schreiben" deaktiviert$`, s.browserWorkspaceOnly)
	sc.Step(`^ich die Profilseite neu lade$`, s.browserReload)
	sc.Step(`^der CLI-Durchsetzungsgrund bleibt als "Nicht startbereit" sichtbar$`, s.browserUnverified)
	sc.Step(`^ich sein Arbeitsprofil öffne und nur "Schreiben" aktiviere und speichere$`, s.browserWriteOnly)
	sc.Step(`^sehe ich einen Feldfehler für "Schreiben"$`, s.browserWriteError)
	sc.Step(`^die Schalter "Arbeitsbereich" und "Schreiben" bleiben deaktiviert$`, s.browserBothDisabled)
	sc.Step(`^"Kai" gehört zu "Nordstern" und "Südlicht" besteht$`, s.browserTwoOrganizations)
	sc.Step(`^ich Kais Arbeitsprofil über die Browseradresse von "Südlicht" öffne$`, s.browserForeign)
	sc.Step(`^sehe ich weder Kais Profil noch seinen Arbeitsbereich$`, s.browserNoLeak)
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

func (s *Suite) stopBrowser() {
	if s.browser == nil {
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

func (s *Suite) browserAgent() error {
	if err := s.prepare(); err != nil {
		return err
	}
	if err := s.startBrowser(); err != nil {
		return err
	}
	return s.browserNavigate("/organisationen/" + s.orgID + "/agenten/" + s.agentID)
}

func (s *Suite) browserOpenProfile() error {
	return s.browserNavigate(s.pagePath())
}

func (s *Suite) browserDerived() error {
	if s.page.PathInput || !strings.Contains(strings.ToLower(s.page.Text), "arbeitsbereich") {
		return fmt.Errorf("privater Arbeitsbereich oder Pfadverbot fehlt: %+v", s.page)
	}
	return nil
}

func (s *Suite) browserBothDisabled() error {
	if s.page.WorkspaceEnabled || s.page.WriteEnabled {
		return fmt.Errorf("Schalter nicht deaktiviert: %+v", s.page)
	}
	return nil
}

func (s *Suite) browserUnverified() error {
	text := strings.ToLower(s.page.Text)
	if !strings.Contains(text, "nicht startbereit") || !strings.Contains(text, "nicht nachgewiesen") {
		return fmt.Errorf("CLI-Durchsetzungsgrund fehlt: %+v", s.page)
	}
	return nil
}

func (s *Suite) browserSet(name string, value bool) error {
	return s.browserCommand(map[string]any{"setCheckbox": map[string]any{"name": name, "value": value}})
}

func (s *Suite) browserEnableWorkspace() error {
	if err := s.browserOpenProfile(); err != nil {
		return err
	}
	if err := s.browserSet("workspace_enabled", true); err != nil {
		return err
	}
	return s.browserCommand(map[string]any{"submit": true})
}

func (s *Suite) browserWriteOnly() error {
	if err := s.browserOpenProfile(); err != nil {
		return err
	}
	if err := s.browserSet("write_enabled", true); err != nil {
		return err
	}
	return s.browserCommand(map[string]any{"submit": true})
}

func (s *Suite) browserWorkspaceOnly() error {
	if !s.page.WorkspaceEnabled || s.page.WriteEnabled {
		return fmt.Errorf("Profilzustand im Browser falsch: %+v", s.page)
	}
	return nil
}

func (s *Suite) browserReload() error { return s.browserNavigate(s.pagePath()) }

func (s *Suite) browserWriteError() error {
	if !strings.Contains(strings.ToLower(s.page.Text), "schreiben") || !strings.Contains(strings.ToLower(s.page.Text), "arbeitsbereich") {
		return fmt.Errorf("Schreibfehler fehlt: %+v", s.page)
	}
	return nil
}

func (s *Suite) browserTwoOrganizations() error {
	if err := s.prepare(); err != nil {
		return err
	}
	if err := s.createSuedlicht(); err != nil {
		return err
	}
	return s.startBrowser()
}

func (s *Suite) browserForeign() error {
	return s.browserNavigate("/organisationen/" + s.otherOrgID + "/agenten/" + s.agentID + "/codex-profil")
}

func (s *Suite) browserNoLeak() error {
	if strings.Contains(s.page.Text, "Kai") || strings.Contains(s.page.Text, s.agentID) || strings.Contains(s.page.Text, s.orgID) {
		return fmt.Errorf("fremde Profildaten: %+v", s.page)
	}
	return nil
}
