package datenbereiche

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

	"agentcontrolplane/app/internal/domain/rechte"

	"github.com/cucumber/godog"
)

func NewSuite(t *testing.T) *Suite {
	client := &http.Client{Timeout: 4 * time.Second, Transport: &http.Transport{Proxy: nil}, CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
		return http.ErrUseLastResponse
	}}
	return &Suite{t: t, client: client, organizations: map[string]string{}, agents: map[string]string{}, projects: map[string]string{}, goals: map[string]string{}}
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

func (s *Suite) freshServer() error {
	s.stopServer()
	s.database = filepath.Join(s.t.TempDir(), "datenbereiche.sqlite")
	s.organizations, s.agents, s.projects, s.goals = map[string]string{}, map[string]string{}, map[string]string{}, map[string]string{}
	return s.startServer()
}

func (s *Suite) startServer() error {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return err
	}

	s.address = listener.Addr().String()
	listener.Close()
	return s.launch()
}

func (s *Suite) launch() error {
	file, err := os.CreateTemp(s.t.TempDir(), "datenbereich-server-")
	if err != nil {
		return err
	}

	s.logFile = file
	s.process = exec.Command(s.binary)
	s.process.Env = append(os.Environ(), "APP_ADDR="+s.address, "APP_DB_PATH="+s.database)
	s.process.Stdout, s.process.Stderr = file, file
	if err := s.process.Start(); err != nil {
		return err
	}

	s.exited = make(chan error, 1)
	go func() { s.exited <- s.process.Wait() }()
	return s.waitHealthy()
}

func (s *Suite) waitHealthy() error {
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		response, err := s.client.Get(s.baseURL() + "/health")
		if err == nil && response.StatusCode == http.StatusOK {
			response.Body.Close()
			return nil
		}

		if response != nil {
			response.Body.Close()
		}

		time.Sleep(20 * time.Millisecond)
	}

	return fmt.Errorf("Server nicht erreichbar: %s", s.serverLog())
}

func (s *Suite) serverLog() string {
	if s.logFile == nil {
		return ""
	}

	data, _ := os.ReadFile(s.logFile.Name())
	return string(data)
}

func (s *Suite) restartServer() error {
	s.stopServer()
	s.closeDataFacade()
	if err := s.startServer(); err != nil {
		return err
	}

	return s.openDataFacade()
}

func (s *Suite) stopServer() {
	if s.process == nil {
		return
	}

	_ = s.process.Process.Kill()
	<-s.exited
	_ = s.logFile.Close()
	s.process = nil
}

func (s *Suite) cleanup() {
	s.stopServer()
	s.closeDataFacade()
	s.stopBrowser()
}

func (s *Suite) afterScenario(ctx context.Context, _ *godog.Scenario, _ error) (context.Context, error) {
	s.cleanup()
	s.response, s.previous = nil, nil
	s.organizations, s.agents, s.projects, s.goals = map[string]string{}, map[string]string{}, map[string]string{}, map[string]string{}
	s.page = BrowserPage{}
	s.agentActor, s.lastError = rechte.Actor{}, nil
	return ctx, nil
}

func (s *Suite) baseURL() string { return "http://" + s.address }

func (s *Suite) request(method, path, body, contentType string) error {
	return s.requestTo(s.baseURL()+path, method, body, contentType)
}

func (s *Suite) requestTo(target, method, body, contentType string) error {
	req, err := http.NewRequest(method, target, strings.NewReader(body))
	if err != nil {
		return err
	}

	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	response, err := s.client.Do(req)
	if err != nil {
		return err
	}

	return s.recordResponse(response)
}

func (s *Suite) recordResponse(response *http.Response) error {
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}

	s.previous = s.response
	s.response = &HTTPResponse{Status: response.StatusCode, Body: body, Header: response.Header}
	return nil
}
