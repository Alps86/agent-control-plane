package agentenvorlagen

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

func NewSuite(t *testing.T) *Suite {
	client := &http.Client{Timeout: 4 * time.Second, Transport: &http.Transport{Proxy: nil}, CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
		return http.ErrUseLastResponse
	}}
	return &Suite{t: t, client: client, orgIDs: map[string]string{}, agentIDs: map[string]string{}}
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
	s.dbPath = filepath.Join(s.t.TempDir(), "agenten.sqlite")
	s.orgIDs, s.agentIDs = map[string]string{}, map[string]string{}
	return s.startServer()
}

func (s *Suite) startServer() error {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return err
	}
	s.address = listener.Addr().String()
	s.listenAddress = s.address
	listener.Close()
	return s.launch()
}

func (s *Suite) startWildcardServer() error {
	s.stopServer()
	if err := s.reserveWildcardAddress(); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, s.binary)
	command.Env = append(os.Environ(), "APP_ADDR="+s.listenAddress, "APP_DB_PATH="+s.dbPath)
	output, err := command.CombinedOutput()
	if ctx.Err() != nil || err == nil {
		return fmt.Errorf("wildcard server did not fail startup")
	}
	s.wildcardRejected = strings.Contains(string(output), "APP_ADDR must be a local loopback address")
	return nil
}

func (s *Suite) reserveWildcardAddress() error {
	listener, err := net.Listen("tcp", "0.0.0.0:0")
	if err != nil {
		return err
	}
	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()
	s.listenAddress = fmt.Sprintf("0.0.0.0:%d", port)
	s.address = fmt.Sprintf("127.0.0.1:%d", port)
	return nil
}

func (s *Suite) wildcardConfigurationRejected() error {
	if !s.wildcardRejected {
		return fmt.Errorf("wildcard startup lacked APP_ADDR loopback rejection")
	}
	return nil
}

func (s *Suite) noWildcardListener() error {
	response, err := s.client.Get(s.baseURL() + "/health")
	if err != nil {
		return nil
	}
	response.Body.Close()
	return fmt.Errorf("wildcard startup left HTTP listener reachable")
}

func (s *Suite) startAlternateLoopbackServer() error {
	s.stopServer()
	listener, err := net.Listen("tcp", "127.0.0.2:0")
	if err != nil {
		return err
	}
	s.address = listener.Addr().String()
	s.listenAddress = s.address
	listener.Close()
	return s.launch()
}

func (s *Suite) launch() error {
	file, err := os.CreateTemp(s.t.TempDir(), "agent-server-log-")
	if err != nil {
		return err
	}
	s.logFile = file
	s.process = exec.Command(s.binary)
	s.process.Env = append(os.Environ(), "APP_ADDR="+s.listenAddress, "APP_DB_PATH="+s.dbPath)
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
	return s.startServer()
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
	s.stopNegative()
	s.stopBrowser()
}

func (s *Suite) afterScenario(ctx context.Context, _ *godog.Scenario, _ error) (context.Context, error) {
	s.stopServer()
	s.stopNegative()
	s.stopBrowser()
	s.dbPath, s.address, s.listenAddress = "", "", ""
	s.orgIDs, s.agentIDs = map[string]string{}, map[string]string{}
	s.response, s.initialID, s.initialURL = nil, "", ""
	s.initialProfile = Profile{}
	s.selectedTeam, s.selectedAgent, s.selectedOrg, s.enteredName = "", "", "", ""
	s.page = BrowserPage{}
	s.wildcardRejected = false
	return ctx, nil
}

func (s *Suite) baseURL() string { return "http://" + s.address }

func (s *Suite) request(method, path, body, contentType string) error {
	return s.requestWith(method, path, body, contentType, "")
}

func (s *Suite) requestWith(method, path, body, contentType, origin string) error {
	req, err := http.NewRequest(method, s.baseURL()+path, strings.NewReader(body))
	if err != nil {
		return err
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	if origin != "" {
		req.Header.Set("Origin", origin)
	}
	response, err := s.client.Do(req)
	if err != nil {
		return err
	}
	return s.recordResponse(response)
}

func (s *Suite) recordResponse(response *http.Response) error {
	defer response.Body.Close()
	data, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}
	s.response = &HTTPResponse{Status: response.StatusCode, Location: response.Header.Get("Location"), Body: data, Header: response.Header}
	return nil
}
