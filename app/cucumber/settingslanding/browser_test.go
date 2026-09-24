package settingslanding

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

func (s *Suite) browseProvider(provider string) error {
	id := ""
	if provider == "Codex-Abo" {
		id = "settings-codex-link"
	}
	if provider == "OpenRouter" {
		id = "settings-openrouter-link"
	}
	if id == "" {
		return fmt.Errorf("unknown provider %q", provider)
	}
	if err := s.startBrowser(); err != nil {
		return err
	}
	return s.browserAction(map[string]string{"action": "open", "url": s.baseURL + "/settings", "id": id})
}

func (s *Suite) startBrowser() error {
	script, err := filepath.Abs("browser.mjs")
	if err != nil {
		return err
	}
	s.browserProcess = exec.Command("node", script)
	s.browserProcess.Stderr = os.Stderr
	s.browserProcess.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	s.browserInput, err = s.browserProcess.StdinPipe()
	if err != nil {
		return err
	}
	output, err := s.browserProcess.StdoutPipe()
	if err != nil {
		return err
	}
	s.browserOutput = bufio.NewScanner(output)
	return s.browserProcess.Start()
}

func (s *Suite) browserAction(action map[string]string) error {
	message, err := json.Marshal(action)
	if err != nil {
		return err
	}
	if _, err := s.browserInput.Write(append(message, '\n')); err != nil {
		return err
	}
	if !s.browserOutput.Scan() {
		return fmt.Errorf("Chrome exited: %v", s.browserOutput.Err())
	}
	if err := json.Unmarshal(s.browserOutput.Bytes(), &s.browser); err != nil {
		return err
	}
	if s.browser.Error != "" {
		return fmt.Errorf("Chrome: %s", s.browser.Error)
	}
	return nil
}

func (s *Suite) providerPage(path string) error {
	if s.browser.Path != path {
		return fmt.Errorf("browser provider path %q, want %q", s.browser.Path, path)
	}
	if strings.Contains(path, "codex") && !strings.Contains(s.browser.ProviderTitle, "Codex") {
		return fmt.Errorf("Codex page heading missing")
	}
	if strings.Contains(path, "openrouter") && !strings.Contains(s.browser.ProviderTitle, "OpenRouter") {
		return fmt.Errorf("OpenRouter page heading missing")
	}
	return nil
}

func (s *Suite) returnLink() error { return s.browserAction(map[string]string{"action": "return"}) }

func (s *Suite) returnedSettings() error {
	if s.browser.Path != "/settings" || s.browser.Links != 2 {
		return fmt.Errorf("browser did not return to Settings with two links: %+v", s.browser)
	}
	return nil
}

func (s *Suite) stopBrowser() {
	if s.browserProcess == nil || s.browserProcess.Process == nil {
		return
	}
	_ = s.browserInput.Close()
	done := make(chan error, 1)
	go func() { done <- s.browserProcess.Wait() }()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		_ = syscall.Kill(-s.browserProcess.Process.Pid, syscall.SIGKILL)
		<-done
	}
	s.browserProcess = nil
}
