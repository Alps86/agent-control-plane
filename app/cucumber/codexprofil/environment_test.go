package codexprofil

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"agentcontrolplane/app/internal/adapter/agent/codexcli"
	"agentcontrolplane/app/internal/adapter/runtime/codexprofil"
	"agentcontrolplane/app/internal/adapter/sqlite"
	webagent "agentcontrolplane/app/internal/adapter/web/agent"
	webprofil "agentcontrolplane/app/internal/adapter/web/codexprofil"
	weborganisation "agentcontrolplane/app/internal/adapter/web/organisation"
	appagent "agentcontrolplane/app/internal/app/agent"
	appprofil "agentcontrolplane/app/internal/app/codexprofil"
	apporganisation "agentcontrolplane/app/internal/app/organisation"
	"agentcontrolplane/ui/bridge"
	"github.com/cucumber/godog"
)

func NewSuite(t *testing.T) *Suite {
	client := &http.Client{Timeout: 30 * time.Second, CheckRedirect: func(_ *http.Request, _ []*http.Request) error { return http.ErrUseLastResponse }}
	return &Suite{t: t, client: client}
}

func (s *Suite) fresh() error {
	s.stopServer()
	s.dbPath = filepath.Join(s.t.TempDir(), "data", "app.sqlite")
	if err := os.MkdirAll(filepath.Dir(s.dbPath), 0700); err != nil {
		return err
	}
	return s.start()
}

func (s *Suite) start() error {
	db, err := sqlite.OpenWithMigrations(context.Background(), s.dbPath,
		sqlite.RunMigration(2), sqlite.OrganizationMigration(), sqlite.GoalMigration(),
		sqlite.AgentMigration(), sqlite.ProjectMigration(),
		sqlite.DataScopeMigration(), sqlite.CodexProfileMigration())
	if err != nil {
		return err
	}
	s.db = db
	ui, err := bridge.New()
	if err != nil {
		return err
	}
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		return err
	}
	s.server = &httptest.Server{Listener: listener, Config: &http.Server{Handler: http.NotFoundHandler()}}
	s.address = s.server.Listener.Addr().String()
	handler, err := s.handler(ui)
	if err != nil {
		return err
	}
	s.server.Config.Handler = handler
	s.server.Start()
	return nil
}

func (s *Suite) handler(ui *bridge.Bridge) (http.Handler, error) {
	mux := http.NewServeMux()
	identity := apporganisation.NewLocalIdentity()
	organizations := sqlite.NewOrganizationStore(s.db)
	orgService := apporganisation.NewService(organizations, identity)
	agentService := appagent.NewService(s.db, identity, organizations)
	agentService.RegisterTemplate(codexcli.NewTemplate())
	agentService.RegisterProbe("codex_cli", codexcli.NewProbe("codex"))
	profileService := appprofil.NewService(s.db, s.db, identity, organizations, codexprofil.NewLocator(s.dbPath))
	runner, err := s.runtimeRunner(profileService)
	if err != nil {
		return nil, err
	}
	s.mount(mux, weborganisation.NewHandler(orgService, ui, s.address), webagent.NewHandler(agentService, ui, s.address),
		webprofil.NewHandler(profileService, agentService, ui, s.address, runner))
	return mux, nil
}

func (s *Suite) mount(mux *http.ServeMux, organizations, agents, profiles http.Handler) {
	for _, path := range []string{"/api/organisationen", "/api/organisationen/", "/organisationen", "/organisationen/"} {
		mux.Handle(path, organizations)
	}
	for _, path := range []string{"/api/organisationen/{id}/agenten", "/api/organisationen/{id}/agenten/", "/organisationen/{id}/agenten", "/organisationen/{id}/agenten/"} {
		mux.Handle(path, agents)
	}
	for _, path := range []string{"/api/organisationen/{id}/agenten/{agentID}/codex-profil", "/organisationen/{id}/agenten/{agentID}/codex-profil"} {
		mux.Handle(path, profiles)
	}
	mux.Handle("/api/organisationen/{id}/agenten/{agentID}/codex-profil/pruefen", profiles)
}

func (s *Suite) restart() error { s.stopServer(); return s.start() }

func (s *Suite) stopServer() {
	if s.server != nil {
		s.server.Close()
		s.server = nil
	}
	if s.db != nil {
		_ = s.db.Close()
		s.db = nil
	}
}

func (s *Suite) cleanup() { s.stopServer(); s.stopBrowser(); s.stopNegative() }

func (s *Suite) after(ctx context.Context, _ *godog.Scenario, _ error) (context.Context, error) {
	s.cleanup()
	s.orgID, s.otherOrgID, s.agentID, s.dbPath = "", "", "", ""
	s.response, s.page = Response{}, Page{}
	s.runtimeScenario = ""
	return ctx, nil
}

func (s *Suite) url(path string) string { return "http://" + s.address + path }
func (s *Suite) profilePath() string {
	return "/api/organisationen/" + s.orgID + "/agenten/" + s.agentID + "/codex-profil"
}
func (s *Suite) pagePath() string {
	return "/organisationen/" + s.orgID + "/agenten/" + s.agentID + "/codex-profil"
}

func (s *Suite) request(method, path, body, origin string) error {
	request, err := http.NewRequest(method, s.url(path), strings.NewReader(body))
	if err != nil {
		return err
	}
	if body != "" {
		request.Header.Set("Content-Type", "application/json")
	}
	if origin != "" {
		request.Header.Set("Origin", origin)
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
	s.response = Response{Method: method, Status: response.StatusCode, Body: data, Location: response.Header.Get("Location")}
	return nil
}

func (s *Suite) createOrganization(name string) error {
	body, _ := json.Marshal(map[string]string{"name": name, "description": "Testorganisation"})
	if err := s.request("POST", "/api/organisationen", string(body), ""); err != nil {
		return err
	}
	if s.response.Status != 201 {
		return fmt.Errorf("Organisation: HTTP %d: %s", s.response.Status, s.response.Body)
	}
	var organization Organization
	if err := json.Unmarshal(s.response.Body, &organization); err != nil {
		return err
	}
	if name == "Nordstern" {
		s.orgID = organization.ID
	} else {
		s.otherOrgID = organization.ID
	}
	return nil
}

func (s *Suite) createAgent() error {
	body := `{"name":"Kai","template_id":"codex-cli","execution_kind":"codex_cli"}`
	path := "/api/organisationen/" + s.orgID + "/agenten"
	if err := s.request("POST", path, body, ""); err != nil {
		return err
	}
	if s.response.Status != 201 {
		return fmt.Errorf("Agent: HTTP %d: %s", s.response.Status, s.response.Body)
	}
	var agent Agent
	if err := json.Unmarshal(s.response.Body, &agent); err != nil {
		return err
	}
	s.agentID = agent.ID
	return nil
}

func (s *Suite) prepare() error {
	if err := s.fresh(); err != nil {
		return err
	}
	if err := s.createOrganization("Nordstern"); err != nil {
		return err
	}
	return s.createAgent()
}
