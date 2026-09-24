package codexagent

import (
	"context"
	"fmt"
	"github.com/cucumber/godog"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func NewSuite(t *testing.T) *Suite {
	client := &http.Client{Timeout: 3 * time.Second, CheckRedirect: func(_ *http.Request, _ []*http.Request) error { return http.ErrUseLastResponse }}
	return &Suite{t: t, client: client}
}

func (s *Suite) build() error {
	s.binary = filepath.Join(s.t.TempDir(), "server")
	cmd := exec.Command("go", "build", "-o", s.binary, "./cmd/server")
	cmd.Dir = "../.."
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Server-Build: %w: %s", err, output)
	}
	return nil
}

func (s *Suite) fresh() error {
	s.stopServer()
	s.dbPath = filepath.Join(s.t.TempDir(), "codex-agent.sqlite")
	return s.start()
}

func (s *Suite) start() error {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return err
	}
	s.address = listener.Addr().String()
	listener.Close()
	s.process = exec.Command(s.binary)
	s.process.Env = append(os.Environ(), "APP_ADDR="+s.address, "APP_DB_PATH="+s.dbPath, "APP_CODEX_CLI_REFERENCE="+s.cliReference)
	s.process.Stdout, s.process.Stderr = os.Stderr, os.Stderr
	if err := s.process.Start(); err != nil {
		return err
	}
	return s.healthy()
}

func (s *Suite) healthy() error {
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		response, err := s.client.Get(s.url("/health"))
		if err == nil {
			response.Body.Close()
			if response.StatusCode == 200 {
				return nil
			}
		}
		time.Sleep(20 * time.Millisecond)
	}
	return fmt.Errorf("Server unter %s nicht erreichbar", s.address)
}

func (s *Suite) restart() error { s.stopServer(); return s.start() }
func (s *Suite) stopServer() {
	if s.process == nil {
		return
	}
	_ = s.process.Process.Kill()
	_ = s.process.Wait()
	s.process = nil
}
func (s *Suite) cleanup() { s.stopServer(); s.stopBrowser(); s.stopNegative() }
func (s *Suite) after(ctx context.Context, _ *godog.Scenario, _ error) (context.Context, error) {
	s.stopServer()
	s.stopBrowser()
	s.stopNegative()
	s.orgID, s.agentID, s.agentName, s.dbPath, s.cliReference = "", "", "", "", ""
	s.response, s.page = Response{}, Page{}
	return ctx, nil
}
func (s *Suite) stopNegative() {
	if s.negative != nil {
		s.negative.Close()
		s.negative = nil
	}
	if s.negativeDB != nil {
		_ = s.negativeDB.Close()
		s.negativeDB = nil
	}
}
func (s *Suite) url(path string) string { return "http://" + s.address + path }
func (s *Suite) request(method, path, body, contentType string) error {
	request, err := http.NewRequest(method, s.url(path), strings.NewReader(body))
	if err != nil {
		return err
	}
	if contentType != "" {
		request.Header.Set("Content-Type", contentType)
	}
	response, err := s.client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	data, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}
	s.response = Response{response.StatusCode, data, response.Header.Get("Location")}
	return nil
}
