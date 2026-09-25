package berichtsweg

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/cucumber/godog"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func (s *Suite) newSuite() *Suite {
	s.client = &http.Client{Timeout: 4 * time.Second, Transport: &http.Transport{Proxy: nil}, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	s.organizations, s.agents, s.projects, s.tasks = map[string]string{}, map[string]string{}, map[string]string{}, map[string]string{}
	return s
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
	s.stop()
	s.database = filepath.Join(s.t.TempDir(), "story37.sqlite")
	s.organizations, s.agents, s.projects, s.tasks = map[string]string{}, map[string]string{}, map[string]string{}, map[string]string{}
	return s.start()
}

func (s *Suite) start() error {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return err
	}
	s.address = listener.Addr().String()
	_ = listener.Close()
	file, err := os.CreateTemp(s.t.TempDir(), "story37-log-")
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
		r, err := s.client.Get("http://" + s.address + "/health")
		if err == nil {
			_ = r.Body.Close()
			if r.StatusCode == 200 {
				return nil
			}
		}
		time.Sleep(20 * time.Millisecond)
	}
	data, _ := os.ReadFile(s.logFile.Name())
	return fmt.Errorf("Server nicht erreichbar: %s", data)
}

func (s *Suite) stop() {
	s.stopBrowser()
	if s.foreignClose != nil {
		s.foreignClose()
		s.foreignClose = nil
	}
	if s.process == nil {
		return
	}
	_ = s.process.Process.Kill()
	<-s.exited
	_ = s.logFile.Close()
	s.process = nil
}

func (s *Suite) restart() error {
	s.stopProcess()
	return s.start()
}

func (s *Suite) stopProcess() {
	if s.process == nil {
		return
	}
	_ = s.process.Process.Kill()
	<-s.exited
	_ = s.logFile.Close()
	s.process = nil
}

func (s *Suite) after(ctx context.Context, _ *godog.Scenario, _ error) (context.Context, error) {
	s.stop()
	s.response = Response{}
	s.foreignReplies, s.unknownReplies = nil, nil
	return ctx, nil
}

func (s *Suite) call(method, path string, body any) error {
	var input io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return err
		}
		input = strings.NewReader(string(data))
	}
	request, err := http.NewRequest(method, "http://"+s.address+path, input)
	if err != nil {
		return err
	}
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
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
	s.response = Response{Status: response.StatusCode, Body: data, Header: response.Header.Clone()}
	return nil
}

func (s *Suite) status(want int) error {
	if s.response.Status != want {
		return fmt.Errorf("HTTP %d statt %d: %s", s.response.Status, want, s.response.Body)
	}
	return nil
}

func (s *Suite) key(org, name string) string     { return org + "\x00" + name }
func (s *Suite) orgID(name string) string        { return s.organizations[name] }
func (s *Suite) agentID(org, name string) string { return s.agents[s.key(org, name)] }
func (s *Suite) chartPath(org string) string {
	return "/api/organisationen/" + s.orgID(org) + "/berichtswege"
}
func (s *Suite) profilePath(org, name string) string {
	return "/api/organisationen/" + s.orgID(org) + "/agenten/" + s.agentID(org, name)
}
