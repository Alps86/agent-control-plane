package geheimnisreferenzen

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"

	credentialadapter "agentcontrolplane/app/internal/adapter/credentials"
	"agentcontrolplane/app/internal/adapter/model/codexauth"
	modelopenrouter "agentcontrolplane/app/internal/adapter/model/openrouter"
	"agentcontrolplane/app/internal/adapter/sqlite"
	webagent "agentcontrolplane/app/internal/adapter/web/agent"
	webcodex "agentcontrolplane/app/internal/adapter/web/modellanbieter"
	webgrant "agentcontrolplane/app/internal/adapter/web/modellfreigabe"
	webchoice "agentcontrolplane/app/internal/adapter/web/modellwahl"
	webopenrouter "agentcontrolplane/app/internal/adapter/web/openrouter"
	weborg "agentcontrolplane/app/internal/adapter/web/organisation"
	webstatus "agentcontrolplane/app/internal/adapter/web/verbindungsstatus"
	appagent "agentcontrolplane/app/internal/app/agent"
	appgrant "agentcontrolplane/app/internal/app/modellfreigabe"
	"agentcontrolplane/app/internal/app/modellverbindung"
	appchoice "agentcontrolplane/app/internal/app/modellwahl"
	"agentcontrolplane/app/internal/app/openrouterverbindung"
	apporg "agentcontrolplane/app/internal/app/organisation"
	appstatus "agentcontrolplane/app/internal/app/verbindungsstatus"
	"agentcontrolplane/ui/bridge"
	"github.com/cucumber/godog"
)

func (s *Suite) InitializeScenario(sc *godog.ScenarioContext) {
	*s = Suite{t: s.t}
	sc.After(s.cleanup)
	s.registerSteps(sc)
}

func (s *Suite) start() error {
	s.dir = s.t.TempDir()
	s.issuerState = &Issuer{}
	s.issuer = httptest.NewServer(s.issuerState)
	s.provider = httptest.NewServer(http.HandlerFunc(s.providerResponse))
	db, err := sqlite.OpenApplication(context.Background(), filepath.Join(s.dir, "app.sqlite"))
	if err != nil {
		return err
	}
	s.db = db
	return s.launch()
}

func (s *Suite) launch() error {
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		return err
	}
	s.address = listener.Addr().String()
	handler, err := s.routes()
	if err != nil {
		_ = listener.Close()
		return err
	}
	s.app = httptest.NewUnstartedServer(handler)
	_ = s.app.Listener.Close()
	s.app.Listener = listener
	s.app.Start()
	return nil
}

func (s *Suite) routes() (http.Handler, error) {
	credentialDir := filepath.Join(s.dir, "credentials")
	if err := os.Mkdir(credentialDir, 0700); err != nil && !os.IsExist(err) {
		return nil, err
	}
	store, err := credentialadapter.NewStore(filepath.Join(credentialDir, "secrets.enc"), filepath.Join(credentialDir, "master.key"))
	if err != nil {
		return nil, err
	}
	auth, err := codexauth.NewDirect(codexauth.Config{Issuer: s.issuer.URL, ClientID: codexauth.CodexCLIClientID})
	if err != nil {
		return nil, err
	}
	s.codex = modellverbindung.New(auth, store)
	s.router = openrouterverbindung.NewService(store, modelopenrouter.NewProbeAt(nil, s.provider.URL+"/api/v1/key"))
	return s.mount()
}

func (s *Suite) mount() (http.Handler, error) {
	ui, err := bridge.New()
	if err != nil {
		return nil, err
	}
	catalog, err := appchoice.NewCatalogLoader().Load("")
	if err != nil {
		return nil, err
	}
	status := webstatus.NewHandler(appstatus.NewService(catalog, s.statusSources()...), ui)
	codex, err := webcodex.New(s.codex, s.address)
	if err != nil {
		return nil, err
	}
	router := webopenrouter.NewHandler(s.router, ui, s.address)
	return s.mountRoutes(status, codex.Handler(), router, ui), nil
}

func (s *Suite) statusSources() []appstatus.Source {
	sources := []appstatus.Source{appstatus.NewCodexSource(s.codex), appstatus.NewOpenRouterSource(s.router)}
	if s.later {
		sources = append(sources, appstatus.Source{ProviderID: "later", Name: "Späterer Anbieter", AuthType: "api_key", Reference: "later-central", Port: LaterPort{}})
	}
	return sources
}

func (s *Suite) mountRoutes(status, codex, router http.Handler, ui *bridge.Bridge) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/api/settings/modellanbieter/status", status)
	mux.Handle("/settings/modellanbieter/status", status)
	mux.Handle("/settings/modelle/codex/device/", codex)
	mux.Handle("/api/settings/modellanbieter/openrouter", router)
	mux.Handle("/api/settings/modellanbieter/openrouter/", router)
	mux.Handle("/settings/modellanbieter/openrouter", router)
	s.mountEntities(mux, ui)
	return mux
}

func (s *Suite) mountEntities(mux *http.ServeMux, ui *bridge.Bridge) {
	identity := apporg.NewLocalIdentity()
	organizations := sqlite.NewOrganizationStore(s.db)
	s.grants = appgrant.NewService(sqlite.NewModellfreigabeStore(s.db), organizations, s.db, identity, s.router)
	org := weborg.NewHandler(apporg.NewService(organizations, identity), ui, s.address)
	agent := webagent.NewHandler(appagent.NewService(s.db, identity, organizations), ui, s.address)
	grant := webgrant.NewHandler(s.grants, ui, s.address)
	catalog, _ := appchoice.NewCatalogLoader().Load("")
	choiceService := appchoice.NewService(s.db, identity, ChoiceStore{s.db}, ChoiceGrant{s.grants}, catalog)
	choice := webchoice.NewHandler(choiceService, ui, s.address)
	mux.Handle("/api/organisationen", org)
	mux.Handle("/api/organisationen/", s.entityRoutes(org, agent, grant, choice))
}

func (s *Suite) entityRoutes(org, agent, grant, choice http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/modellfreigabe/openrouter") {
			grant.ServeHTTP(w, r)
			return
		}
		if strings.HasSuffix(r.URL.Path, "/modellwahl") {
			choice.ServeHTTP(w, r)
			return
		}
		if strings.Contains(r.URL.Path, "/agenten") {
			agent.ServeHTTP(w, r)
			return
		}
		org.ServeHTTP(w, r)
	})
}

func (s *Suite) cleanup(ctx context.Context, _ *godog.Scenario, _ error) (context.Context, error) {
	if s.app != nil {
		s.app.Close()
	}
	if s.provider != nil {
		s.provider.Close()
	}
	if s.issuer != nil {
		s.issuer.Close()
	}
	if s.db != nil {
		_ = s.db.Close()
	}
	return ctx, nil
}

func (s *Suite) restart() error {
	if s.app == nil {
		return fmt.Errorf("Server nicht gestartet")
	}
	if err := s.getStatus(); err != nil {
		return err
	}
	s.before = append([]byte(nil), s.last...)
	s.app.Close()
	s.app = nil
	return s.launch()
}
