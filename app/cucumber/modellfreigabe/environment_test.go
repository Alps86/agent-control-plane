package modellfreigabe

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	credentialsadapter "agentcontrolplane/app/internal/adapter/credentials"
	modelopenrouter "agentcontrolplane/app/internal/adapter/model/openrouter"
	"agentcontrolplane/app/internal/adapter/sqlite"
	webagent "agentcontrolplane/app/internal/adapter/web/agent"
	webfreigabe "agentcontrolplane/app/internal/adapter/web/modellfreigabe"
	webopenrouter "agentcontrolplane/app/internal/adapter/web/openrouter"
	weborganisation "agentcontrolplane/app/internal/adapter/web/organisation"
	appagent "agentcontrolplane/app/internal/app/agent"
	appfreigabe "agentcontrolplane/app/internal/app/modellfreigabe"
	appopenrouter "agentcontrolplane/app/internal/app/openrouterverbindung"
	apporganisation "agentcontrolplane/app/internal/app/organisation"
	"agentcontrolplane/ui/bridge"
	"github.com/cucumber/godog"
)

const firstKey = "sk-or-story24-SYNTHETIC-Alpha-629451"
const secondKey = "sk-or-story24-SYNTHETIC-Beta-728453"

type suite struct {
	t                 *testing.T
	db                *sqlite.Database
	app               *httptest.Server
	provider          *httptest.Server
	service           *appfreigabe.Service
	browser           chromeBrowser
	client            *http.Client
	orgIDs            map[string]string
	agentIDs          map[string]string
	agentOrgs         map[string]string
	reference         string
	status            int
	body              []byte
	modelPosts        atomic.Int64
	reconnectBaseline int
	callerEntered     chan struct{}
	callerRelease     chan struct{}
	invokeDone        chan error
	lastDecision      string
	lastError         error
	waitingAgent      string
	pageHTML          []byte
	jsonBody          []byte
	browserHTML       []string
}

func newSuite(t *testing.T) *suite {
	return &suite{t: t, client: &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }},
		orgIDs: map[string]string{}, agentIDs: map[string]string{}, agentOrgs: map[string]string{}}
}

func (s *suite) afterScenario(ctx context.Context, _ *godog.Scenario, _ error) (context.Context, error) {
	if s.callerRelease != nil {
		close(s.callerRelease)
		select {
		case <-s.invokeDone:
		case <-time.After(3 * time.Second):
		}
	}
	s.browser.stop()
	if s.app != nil {
		s.app.Close()
	}
	if s.provider != nil {
		s.provider.Close()
	}
	if s.db != nil {
		_ = s.db.Close()
	}
	s.app, s.provider, s.db, s.service = nil, nil, nil, nil
	s.orgIDs, s.agentIDs, s.agentOrgs = map[string]string{}, map[string]string{}, map[string]string{}
	s.reference, s.lastDecision, s.waitingAgent, s.body, s.pageHTML, s.jsonBody = "", "", "", nil, nil, nil
	s.browserHTML = nil
	s.status, s.reconnectBaseline, s.lastError = 0, 0, nil
	s.modelPosts.Store(0)
	s.callerEntered, s.callerRelease, s.invokeDone = nil, nil, nil
	return ctx, nil
}

func (s *suite) start() error {
	if s.app != nil {
		return nil
	}
	s.provider = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && r.URL.Path == "/api/v1/chat/completions" {
			s.modelPosts.Add(1)
			if r.Header.Get("Authorization") != "Bearer "+firstKey && r.Header.Get("Authorization") != "Bearer "+secondKey {
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, `{"id":"synthetic-chat","choices":[{"message":{"content":"ok"}}]}`)
			return
		}
		if r.Method != http.MethodGet || r.URL.Path != "/api/v1/key" ||
			(r.Header.Get("Authorization") != "Bearer "+firstKey && r.Header.Get("Authorization") != "Bearer "+secondKey) {
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"data":{"label":"synthetic-test-provider"}}`)
	}))

	var err error
	s.db, err = sqlite.OpenApplication(context.Background(), filepath.Join(s.t.TempDir(), "story24.sqlite"))
	if err != nil {
		return fmt.Errorf("kanonische SQLite-Migrationskette noch nicht vollständig: %w", err)
	}
	ui, err := bridge.New()
	if err != nil {
		return err
	}
	credentialDir := s.t.TempDir()
	if err := os.Chmod(credentialDir, 0700); err != nil {
		return err
	}
	credentials, err := credentialsadapter.NewStore(filepath.Join(credentialDir, "secrets.enc"), filepath.Join(credentialDir, "master.key"))
	if err != nil {
		return err
	}
	connection := appopenrouter.NewService(credentials, modelopenrouter.NewProbeAt(nil, s.provider.URL+"/api/v1/key"))
	identity := apporganisation.NewLocalIdentity()
	orgStore := sqlite.NewOrganizationStore(s.db)
	orgService := apporganisation.NewService(orgStore, identity)
	agentService := appagent.NewService(s.db, identity, orgStore)
	s.service = appfreigabe.NewService(sqlite.NewModellfreigabeStore(s.db), orgStore, s.db, identity, connection)

	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		return err
	}
	address := listener.Addr().String()
	orgHandler := weborganisation.NewHandler(orgService, ui, address)
	agentHandler := webagent.NewHandler(agentService, ui, address)
	grantHandler := webfreigabe.NewHandler(s.service, ui, address)
	settingsHandler := webopenrouter.NewHandler(connection, ui, address)
	s.app = httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasPrefix(r.URL.Path, "/assets/"), strings.HasPrefix(r.URL.Path, "/fragments/"):
			ui.Assets().ServeHTTP(w, r)
		case strings.HasPrefix(r.URL.Path, "/settings/modellanbieter/openrouter"), strings.HasPrefix(r.URL.Path, "/api/settings/modellanbieter/openrouter"):
			settingsHandler.ServeHTTP(w, r)
		case strings.Contains(r.URL.Path, "/modellfreigabe/openrouter"):
			grantHandler.ServeHTTP(w, r)
		case strings.Contains(r.URL.Path, "/agenten"):
			agentHandler.ServeHTTP(w, r)
		default:
			orgHandler.ServeHTTP(w, r)
		}
	}))
	s.app.Listener.Close()
	s.app.Listener = listener
	s.app.Start()
	return nil
}

func (s *suite) request(method, path string, payload any) error {
	var reader io.Reader
	if payload != nil {
		body, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(body)
	}
	request, err := http.NewRequest(method, s.app.URL+path, reader)
	if err != nil {
		return err
	}
	if method != http.MethodGet {
		request.Header.Set("Origin", s.app.URL)
	}
	if payload != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	response, err := s.client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	s.status = response.StatusCode
	s.body, err = io.ReadAll(io.LimitReader(response.Body, 1<<20))
	return err
}

func (s *suite) settingsReady() error {
	if err := s.start(); err != nil {
		return err
	}
	if err := s.request(http.MethodPost, "/api/settings/modellanbieter/openrouter", map[string]string{"key": firstKey}); err != nil {
		return err
	}
	if s.status != http.StatusOK {
		return fmt.Errorf("Settings-Save: HTTP %d: %s", s.status, s.body)
	}
	if err := s.request(http.MethodPost, "/api/settings/modellanbieter/openrouter/pruefen", nil); err != nil {
		return err
	}
	if s.status != http.StatusOK {
		return fmt.Errorf("Settings-Check: HTTP %d: %s", s.status, s.body)
	}
	var view struct{ Reference, Status string }
	if err := json.Unmarshal(s.body, &view); err != nil {
		return err
	}
	if view.Reference == "" || view.Status != "einsatzbereit" {
		return fmt.Errorf("kontrollierter Status nicht einsatzbereit: %s", s.body)
	}
	s.reference = view.Reference
	return nil
}

func (s *suite) modelCalls() int { return int(s.modelPosts.Load()) }

func (s *suite) createOrganization(name string) error {
	if err := s.start(); err != nil {
		return err
	}
	if s.orgIDs[name] != "" {
		return nil
	}
	if err := s.request(http.MethodPost, "/api/organisationen", map[string]string{"name": name, "description": "Synthetische Organisation " + name}); err != nil {
		return err
	}
	if s.status != http.StatusCreated {
		return fmt.Errorf("Organisation %s: HTTP %d: %s", name, s.status, s.body)
	}
	var item struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(s.body, &item); err != nil {
		return err
	}
	if item.ID == "" {
		return fmt.Errorf("Organisationskennung fehlt")
	}
	s.orgIDs[name] = item.ID
	return nil
}

func (s *suite) createAgent(name, org string) error {
	if err := s.createOrganization(org); err != nil {
		return err
	}
	if s.agentIDs[name] != "" {
		return nil
	}
	path := "/api/organisationen/" + url.PathEscape(s.orgIDs[org]) + "/agenten"
	if err := s.request(http.MethodPost, path, map[string]string{"name": name, "template_id": "recherche", "execution_kind": "eino"}); err != nil {
		return err
	}
	if s.status != http.StatusCreated {
		return fmt.Errorf("Agent %s: HTTP %d: %s", name, s.status, s.body)
	}
	var item struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(s.body, &item); err != nil {
		return err
	}
	if item.ID == "" {
		return fmt.Errorf("Agentenkennung fehlt")
	}
	s.agentIDs[name], s.agentOrgs[name] = item.ID, org
	return nil
}
