package uibridge

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"testing"

	"agentcontrolplane/ui/bridge"
)

func NewSuite(t *testing.T) *Suite {
	return &Suite{t: t}
}

func (s *Suite) start() error {
	var err error
	s.bridge, err = bridge.New()
	if err != nil {
		return err
	}

	mux := http.NewServeMux()
	mux.Handle("/", s.bridge.Assets())
	mux.HandleFunc("GET /{$}", s.serve)
	mux.HandleFunc("GET /fragment", s.serve)
	s.server = httptest.NewServer(mux)
	return s.startBrowser()
}

func (s *Suite) serve(w http.ResponseWriter, r *http.Request) {
	name := "organization.json"
	if r.URL.Query().Get("view") == "projects" {
		name = "projects.json"
	}

	data, err := s.bridge.LoadFixture(name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	part := "page"
	if r.URL.Path == "/fragment" {
		part = "content"
	}

	if err := s.bridge.Render(w, part, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
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

	s.stdin = bufio.NewWriter(input)
	s.stdout = bufio.NewScanner(output)
	return s.browser.Start()
}

func (s *Suite) browserPage(url string) (BrowserPage, error) {
	return s.browserCommand(map[string]any{"url": url})
}

func (s *Suite) browserCommand(request map[string]any) (BrowserPage, error) {
	message, err := json.Marshal(request)
	if err != nil {
		return BrowserPage{}, err
	}

	if _, err := s.stdin.Write(append(message, '\n')); err != nil {
		return BrowserPage{}, err
	}

	if err := s.stdin.Flush(); err != nil {
		return BrowserPage{}, err
	}

	return s.readBrowser()
}

func (s *Suite) readBrowser() (BrowserPage, error) {
	if !s.stdout.Scan() {
		return BrowserPage{}, fmt.Errorf("Chrome ended: %w", s.stdout.Err())
	}

	var reply BrowserReply
	if err := json.Unmarshal(s.stdout.Bytes(), &reply); err != nil {
		return BrowserPage{}, err
	}

	if !reply.OK {
		return BrowserPage{}, fmt.Errorf("Chrome: %s", reply.Error)
	}

	return reply.Page, nil
}

func (s *Suite) stop() {
	if s.browser != nil && s.browser.Process != nil {
		syscall.Kill(-s.browser.Process.Pid, syscall.SIGKILL)
		s.browser.Wait()
	}

	if s.server != nil {
		s.server.Close()
	}
}

func (s *Suite) fetch(path string) (string, int, error) {
	response, err := http.Get(s.server.URL + path)
	if err != nil {
		return "", 0, err
	}

	defer response.Body.Close()
	content, err := io.ReadAll(response.Body)
	if err != nil {
		return "", 0, err
	}

	return string(content), response.StatusCode, nil
}
