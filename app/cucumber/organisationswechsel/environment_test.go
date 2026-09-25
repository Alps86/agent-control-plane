package organisationswechsel

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
	s.client = &http.Client{Timeout: 4 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	s.organizations = map[string]Organization{}
	return s
}

func (s *Suite) build() error {
	s.binary = filepath.Join(s.t.TempDir(), "server")
	cmd := exec.Command("go", "build", "-o", s.binary, "./cmd/server")
	cmd.Dir = "../.."
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Server bauen: %w: %s", err, out)
	}
	return nil
}

func (s *Suite) fresh() error {
	s.stop()
	s.database = filepath.Join(s.t.TempDir(), "story31.sqlite")
	s.organizations = map[string]Organization{}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return err
	}
	s.address = listener.Addr().String()
	_ = listener.Close()
	f, err := os.CreateTemp(s.t.TempDir(), "story31-log-")
	if err != nil {
		return err
	}
	s.logFile = f
	s.process = exec.Command(s.binary)
	s.process.Env = append(os.Environ(), "APP_ADDR="+s.address, "APP_DB_PATH="+s.database)
	s.process.Stdout, s.process.Stderr = f, f
	if err := s.process.Start(); err != nil {
		return err
	}
	s.exited = make(chan error, 1)
	go func() { s.exited <- s.process.Wait() }()
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
	log, _ := os.ReadFile(f.Name())
	return fmt.Errorf("Server nicht erreichbar: %s", log)
}

func (s *Suite) stop() {
	if s.foreignClose != nil {
		s.foreignClose()
		s.foreignClose = nil
	}
	if s.browser != nil {
		_ = s.browserInput.Close()
		_ = s.browser.Wait()
		s.browser = nil
	}
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
	s.response, s.previous = Response{}, Response{}
	s.baseline, s.lastPreview = nil, nil
	s.lastArea, s.lastOrganization = "", ""
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
	r, err := http.NewRequest(method, "http://"+s.address+path, input)
	if err != nil {
		return err
	}
	if body != nil {
		r.Header.Set("Content-Type", "application/json")
	}
	response, err := s.client.Do(r)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	data, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}
	s.previous, s.response = s.response, Response{Status: response.StatusCode, Body: data, Header: response.Header.Clone()}
	return nil
}

func (s *Suite) create(name string) error {
	if err := s.call("POST", "/api/organisationen", map[string]string{"name": name, "description": "Beschreibung von " + name}); err != nil {
		return err
	}
	if s.response.Status != 201 {
		return fmt.Errorf("Anlegen %s: %d %s", name, s.response.Status, s.response.Body)
	}
	var org Organization
	if err := json.Unmarshal(s.response.Body, &org); err != nil {
		return err
	}
	if org.ID == "" {
		return fmt.Errorf("keine Organisationskennung: %s", s.response.Body)
	}
	s.organizations[name] = org
	return nil
}

func (s *Suite) organization(name string) (Organization, error) {
	org, ok := s.organizations[name]
	if !ok {
		return org, fmt.Errorf("Organisation %q fehlt", name)
	}
	return org, nil
}
