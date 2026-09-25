package kommentare

import (
	"context"
	"encoding/json"
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

	"agentcontrolplane/app/internal/adapter/sqlite"
	appkommentar "agentcontrolplane/app/internal/app/kommentar"
	apporganisation "agentcontrolplane/app/internal/app/organisation"
	"agentcontrolplane/app/internal/domain/rechte"
	portkommentar "agentcontrolplane/app/internal/port/kommentar"
	"github.com/cucumber/godog"
)

func NewSuite(t *testing.T) *Suite {
	client := &http.Client{Timeout: 4 * time.Second, Transport: &http.Transport{Proxy: nil},
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	return &Suite{t: t, client: client, organizations: map[string]string{}, projects: map[string]string{},
		agents: map[string]string{}, tasks: map[string]string{}}
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
	s.database = filepath.Join(s.t.TempDir(), "kommentare.sqlite")
	return s.start()
}

func (s *Suite) start() error {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return err
	}
	s.address = listener.Addr().String()
	_ = listener.Close()
	s.server = exec.Command(s.binary)
	s.server.Env = append(os.Environ(), "APP_ADDR="+s.address, "APP_DB_PATH="+s.database)
	s.server.Stdout, s.server.Stderr = os.Stderr, os.Stderr
	if err := s.server.Start(); err != nil {
		return err
	}
	return s.healthy()
}

func (s *Suite) healthy() error {
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		response, err := s.client.Get(s.url("/health"))
		if err == nil && response.StatusCode == http.StatusOK {
			_ = response.Body.Close()
			return nil
		}
		if response != nil {
			_ = response.Body.Close()
		}
		time.Sleep(20 * time.Millisecond)
	}
	return fmt.Errorf("Server %s nicht erreichbar", s.address)
}

func (s *Suite) restart() error { s.stopServer(); return s.start() }

func (s *Suite) stopServer() {
	if s.server == nil {
		return
	}
	_ = s.server.Process.Kill()
	_ = s.server.Wait()
	s.server = nil
}

func (s *Suite) cleanup() {
	s.stopServer()
	s.stopBrowser()
	if s.store != nil {
		_ = s.store.Close()
		s.store = nil
	}
}

func (s *Suite) after(ctx context.Context, _ *godog.Scenario, _ error) (context.Context, error) {
	s.cleanup()
	s.response, s.page, s.lastComment, s.lastError = Response{}, Page{}, Comment{}, nil
	s.organizations, s.projects = map[string]string{}, map[string]string{}
	s.agents, s.tasks = map[string]string{}, map[string]string{}
	s.activeAgent, s.commentID, s.createdAt, s.content = "", "", "", ""
	s.savedTimestamp = ""
	s.resolver = nil
	s.service, s.database, s.address = nil, "", ""
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
	data, err := io.ReadAll(response.Body)
	s.response = Response{response.StatusCode, data, response.Header.Get("Location")}
	return err
}

func (s *Suite) status(expected int) error {
	if s.response.Status != expected {
		return fmt.Errorf("HTTP %d statt %d: %s", s.response.Status, expected, s.response.Body)
	}
	return nil
}

func (s *Suite) key(org, name string) string     { return org + "\x00" + name }
func (s *Suite) orgID(org string) string         { return s.organizations[org] }
func (s *Suite) taskID(org, title string) string { return s.tasks[s.key(org, title)] }
func (s *Suite) taskPath(org, project, title string) string {
	return "/api/organisationen/" + s.orgID(org) + "/projekte/" + s.projects[s.key(org, project)] + "/aufgaben/" + s.taskID(org, title)
}
func (s *Suite) commentPath(org, project, title string) string {
	return s.taskPath(org, project, title) + "/kommentare"
}

func (s *Suite) post(path string, value any, out any) error {
	body, err := json.Marshal(value)
	if err != nil {
		return err
	}
	if err := s.request("POST", path, string(body)); err != nil {
		return err
	}
	if err := s.status(http.StatusCreated); err != nil {
		return err
	}
	return json.Unmarshal(s.response.Body, out)
}

func (s *Suite) ensureOrganization(name string) error {
	if s.address == "" {
		if err := s.fresh(); err != nil {
			return err
		}
	}
	if s.organizations[name] != "" {
		return nil
	}
	var item Entity
	if err := s.post("/api/organisationen", map[string]string{"name": name, "description": "Testorganisation"}, &item); err != nil {
		return err
	}
	s.organizations[name] = item.ID
	return nil
}

func (s *Suite) ensureProject(org, project string) error {
	if s.projects[s.key(org, project)] != "" {
		return nil
	}
	if err := s.ensureOrganization(org); err != nil {
		return err
	}
	var goal Entity
	if err := s.post("/api/organisationen/"+s.orgID(org)+"/ziele", map[string]string{"name": "Ziel " + project}, &goal); err != nil {
		return err
	}
	var item Entity
	path := "/api/organisationen/" + s.orgID(org) + "/projekte"
	if err := s.post(path, map[string]string{"name": project, "description": "Testprojekt", "goal_id": goal.ID}, &item); err != nil {
		return err
	}
	s.projects[s.key(org, project)] = item.ID
	return nil
}

func (s *Suite) ensureAgent(org, name string) error {
	if s.agents[s.key(org, name)] != "" {
		return nil
	}
	if err := s.ensureOrganization(org); err != nil {
		return err
	}
	var item Entity
	path := "/api/organisationen/" + s.orgID(org) + "/agenten"
	input := map[string]string{"name": name, "template_id": "recherche", "execution_kind": "eino"}
	if err := s.post(path, input, &item); err != nil {
		return err
	}
	s.agents[s.key(org, name)] = item.ID
	return nil
}

func (s *Suite) ensureTask(org, project, title string) error {
	if s.taskID(org, title) != "" {
		return nil
	}
	if err := s.ensureProject(org, project); err != nil {
		return err
	}
	if err := s.ensureAgent(org, "Assignee-"+org); err != nil {
		return err
	}
	var item Entity
	path := "/api/organisationen/" + s.orgID(org) + "/projekte/" + s.projects[s.key(org, project)] + "/aufgaben"
	input := map[string]string{"title": title, "priority": "normal", "assignee_id": s.agents[s.key(org, "Assignee-"+org)]}
	if err := s.post(path, input, &item); err != nil {
		return err
	}
	s.tasks[s.key(org, title)] = item.ID
	return nil
}

func (s *Suite) scope(org, project, agent string, read, write bool) error {
	path := "/api/organisationen/" + s.orgID(org) + "/agenten/" + s.agents[s.key(org, agent)] + "/datenbereich"
	value := map[string]any{"scopes": []map[string]any{{"project_id": s.projects[s.key(org, project)], "can_read": read, "can_write": write}}}
	body, _ := json.Marshal(value)
	if err := s.request("PUT", path, string(body)); err != nil {
		return err
	}
	return s.status(http.StatusOK)
}

func (s *Suite) agentService(org, agent string) error {
	if s.store != nil {
		_ = s.store.Close()
		s.store = nil
	}
	store, err := sqlite.OpenApplication(context.Background(), s.database)
	if err != nil {
		return err
	}
	s.store = store
	identity := AgentIdentity{Actor: rechte.Actor{Kind: rechte.Agent, ID: s.agents[s.key(org, agent)]}}
	s.service = appkommentar.NewService(store, apporganisation.NewLocalIdentity(), identity, s.artifactResolver())
	s.activeAgent = agent
	return nil
}

func (s *Suite) artifactResolver() portkommentar.ArtifactResolver {
	if s.resolver == nil {
		return nil
	}
	return s.resolver
}
