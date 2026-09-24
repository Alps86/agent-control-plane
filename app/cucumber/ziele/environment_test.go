package ziele

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
	return &Suite{t: t, client: client, organizations: map[string]string{}}
}

func (s *Suite) build() error {
	s.binary = filepath.Join(s.t.TempDir(), "server")
	cmd := exec.Command("go", "build", "-o", s.binary, "./cmd/server")
	cmd.Dir = "../.."
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Go-Server bauen: %w: %s", err, output)
	}

	return nil
}

func (s *Suite) freshServer() error {
	s.stopServer()
	s.database = filepath.Join(s.t.TempDir(), "ziele.sqlite")
	s.bindHost = "127.0.0.1"
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
	file, err := os.CreateTemp(s.t.TempDir(), "ziele-server-")
	if err != nil {
		return err
	}

	s.logFile = file
	s.process = exec.Command(s.binary)
	s.process.Env = append(os.Environ(), "APP_ADDR="+net.JoinHostPort(s.bindHost, s.listenerPort()), "APP_DB_PATH="+s.database)
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

	return fmt.Errorf("Server unter %s nicht erreichbar: %s", s.address, s.serverLog())
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

func (s *Suite) listenerPort() string {
	_, port, _ := net.SplitHostPort(s.address)
	return port
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
	s.response, s.createdID, s.active = nil, "", ""
	s.unknownBody = nil
	s.organizations = map[string]string{}
	return ctx, nil
}

func (s *Suite) baseURL() string { return "http://" + s.address }

func (s *Suite) request(method, path, body string, headers http.Header, host string) error {
	return s.requestTo(method, s.baseURL()+path, body, headers, host)
}

func (s *Suite) requestTo(method, target, body string, headers http.Header, host string) error {
	req, err := http.NewRequest(method, target, strings.NewReader(body))
	if err != nil {
		return err
	}

	req.Header = headers.Clone()
	if host != "" {
		req.Host = host
	}

	response, err := s.client.Do(req)
	if err != nil {
		return err
	}

	defer response.Body.Close()
	data, err := io.ReadAll(response.Body)
	s.response = &HTTPResponse{Status: response.StatusCode, Location: response.Header.Get("Location"), Body: data, Header: response.Header}
	return err
}
