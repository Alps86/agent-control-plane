package aktivitaet

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/cucumber/godog"
)

// NewSuite erstellt die öffentliche HTTP-Abnahmeumgebung.
func NewSuite(t *testing.T) *Suite {
	client := &http.Client{Timeout: 4 * time.Second, Transport: &http.Transport{Proxy: nil},
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error { return http.ErrUseLastResponse }}
	return &Suite{t: t, client: client}
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
	s.cleanup()
	s.database = filepath.Join(s.t.TempDir(), "aktivitaet.sqlite")
	s.organizations, s.projects, s.agents = map[string]string{}, map[string]string{}, map[string]string{}
	s.taskID, s.before = "", nil
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
	s.process.Env = append(os.Environ(), "APP_ADDR="+s.address, "APP_DB_PATH="+s.database)
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
		if err == nil && response.StatusCode == http.StatusOK {
			response.Body.Close()
			return nil
		}

		if response != nil {
			response.Body.Close()
		}

		time.Sleep(20 * time.Millisecond)
	}

	return fmt.Errorf("Server %s nicht erreichbar", s.address)
}

func (s *Suite) restart() error {
	s.stopServer()
	return s.start()
}

func (s *Suite) stopServer() {
	if s.process == nil {
		return
	}

	_ = s.process.Process.Kill()
	_ = s.process.Wait()
	s.process = nil
}

func (s *Suite) cleanup() {
	s.stopServer()
	s.stopBrowser()
	if s.fixtureServer != nil {
		s.fixtureServer.Close()
		s.fixtureServer = nil
	}

	if s.fixtureDB != nil {
		_ = s.fixtureDB.Close()
		s.fixtureDB = nil
	}
}

func (s *Suite) after(ctx context.Context, _ *godog.Scenario, _ error) (context.Context, error) {
	s.cleanup()
	s.status, s.body = 0, nil
	s.organizations, s.projects, s.agents = nil, nil, nil
	s.taskID, s.before = "", nil
	return ctx, nil
}

func (s *Suite) url(path string) string { return "http://" + s.address + path }

func (s *Suite) request(method, path, body string) error {
	request, err := http.NewRequest(method, s.url(path), strings.NewReader(body))
	if err != nil {
		return err
	}

	if body != "" {
		request.Header.Set("Content-Type", "application/json")
	}

	response, err := s.client.Do(request)
	if err != nil {
		return err
	}

	defer response.Body.Close()
	s.status, s.body = response.StatusCode, nil
	s.body, err = io.ReadAll(response.Body)
	return err
}
