package projektarchiv

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
	client := &http.Client{Timeout: 5 * time.Second, Transport: &http.Transport{Proxy: nil}, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	return &Suite{t: t, client: client, organizations: map[string]string{}, projects: map[string]string{}, agents: map[string]string{}, tasks: map[string]string{}}
}

func (s *Suite) build() error {
	s.binary = filepath.Join(s.t.TempDir(), "server")
	cmd := exec.Command("go", "build", "-o", s.binary, "./cmd/server")
	cmd.Dir = "../.."
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Server bauen: %w: %s", err, output)
	}
	return nil
}

func (s *Suite) fresh() error {
	s.stopServer()
	s.database = filepath.Join(s.t.TempDir(), "projektarchiv.sqlite")
	return s.start()
}

func (s *Suite) start() error {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return err
	}
	s.address = listener.Addr().String()
	listener.Close()
	return s.launch()
}

func (s *Suite) launch() error {
	file, err := os.CreateTemp(s.t.TempDir(), "archiv-server-")
	if err != nil {
		return err
	}
	s.log = file
	s.server = exec.Command(s.binary)
	s.server.Env = append(os.Environ(), "APP_ADDR="+s.address, "APP_DB_PATH="+s.database)
	s.server.Stdout, s.server.Stderr = file, file
	if err := s.server.Start(); err != nil {
		return err
	}
	return s.healthy()
}

func (s *Suite) healthy() error {
	for i := 0; i < 200; i++ {
		response, err := s.client.Get(s.url("/health"))
		if err == nil && response.StatusCode == http.StatusOK {
			response.Body.Close()
			return nil
		}
		if response != nil {
			response.Body.Close()
		}
		time.Sleep(25 * time.Millisecond)
	}
	return fmt.Errorf("Server nicht bereit: %s", s.serverLog())
}

func (s *Suite) serverLog() string {
	if s.log == nil {
		return ""
	}
	data, _ := os.ReadFile(s.log.Name())
	return string(data)
}

func (s *Suite) restart() error { s.stopServer(); return s.start() }

func (s *Suite) stopServer() {
	if s.server == nil {
		return
	}
	_ = s.server.Process.Kill()
	_ = s.server.Wait()
	_ = s.log.Close()
	s.server = nil
}

func (s *Suite) cleanup() { s.stopServer(); s.stopBrowser() }

func (s *Suite) after(ctx context.Context, _ *godog.Scenario, _ error) (context.Context, error) {
	s.cleanup()
	s.organizations, s.projects = map[string]string{}, map[string]string{}
	s.agents, s.tasks = map[string]string{}, map[string]string{}
	s.response, s.page = Response{}, BrowserPage{}
	s.activeOrg, s.activeProject, s.rejectedTitle = "", "", ""
	s.runID = ""
	s.runDenied = false
	s.activity = ActivityEvent{}
	s.raceArchive, s.raceTask, s.raceTitle = Response{}, Response{}, ""
	s.address, s.database, s.log = "", "", nil
	return ctx, nil
}

func (s *Suite) url(path string) string { return "http://" + s.address + path }

func (s *Suite) request(method, path, body string) error {
	req, err := http.NewRequest(method, s.url(path), strings.NewReader(body))
	if err != nil {
		return err
	}
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	response, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	data, err := io.ReadAll(response.Body)
	s.response = Response{Status: response.StatusCode, Body: data, Location: response.Header.Get("Location")}
	return err
}

func (s *Suite) status(expected int) error {
	if s.response.Status != expected {
		return fmt.Errorf("HTTP %d statt %d: %s", s.response.Status, expected, s.response.Body)
	}
	return nil
}
