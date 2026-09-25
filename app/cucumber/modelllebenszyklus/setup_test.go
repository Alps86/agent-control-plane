package modelllebenszyklus

import (
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"time"

	credentialadapter "agentcontrolplane/app/internal/adapter/credentials"
	"agentcontrolplane/app/internal/adapter/model/codexauth"
	modelopenrouter "agentcontrolplane/app/internal/adapter/model/openrouter"
	webcodex "agentcontrolplane/app/internal/adapter/web/modellanbieter"
	webopenrouter "agentcontrolplane/app/internal/adapter/web/openrouter"
	"agentcontrolplane/app/internal/app/modellverbindung"
	"agentcontrolplane/app/internal/app/openrouterverbindung"
	"github.com/cucumber/godog"
)

func (s *Suite) InitializeScenario(sc *godog.ScenarioContext) {
	*s = Suite{t: s.t}
	sc.After(s.cleanup)
	s.registerCodex(sc)
	s.registerOpenRouter(sc)
	s.registerBinding(sc)
	s.registerInvoke(sc)
	s.registerSecrets(sc)
}

func (s *Suite) setup() error {
	s.dir = s.t.TempDir()
	credentialDir := filepath.Join(s.dir, "credentials")
	if err := os.Mkdir(credentialDir, 0700); err != nil {
		return err
	}
	s.secretPath = filepath.Join(credentialDir, "secrets.enc")
	s.keyPath = filepath.Join(credentialDir, "master.key")
	store, err := credentialadapter.NewStore(s.secretPath, s.keyPath)
	if err != nil {
		return err
	}

	s.store = store
	s.issuerState = &IssuerState{}
	s.issuer = httptest.NewServer(s.issuerState)
	s.providerState = &ProviderState{}
	s.provider = httptest.NewServer(s.providerState)
	return s.startApplication()
}

func (s *Suite) startApplication() error {
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		return err
	}
	address := listener.Addr().String()
	client, err := codexauth.NewDirect(codexauth.Config{Issuer: s.issuer.URL, ClientID: codexauth.CodexCLIClientID})
	if err != nil {
		listener.Close()
		return err
	}

	s.flow = modellverbindung.New(client, s.store)
	codexHandler, err := webcodex.New(s.flow, address)
	if err != nil {
		return err
	}

	s.launchHTTP(listener, codexHandler.Handler(), s.routerHandler(address))
	return nil
}

func (s *Suite) routerHandler(address string) http.Handler {
	probe := modelopenrouter.NewProbeAt(nil, s.provider.URL+"/api/v1/key")
	s.routerService = openrouterverbindung.NewService(s.store, probe)
	return webopenrouter.NewHandler(s.routerService, s, address)
}

func (s *Suite) launchHTTP(listener net.Listener, codex, router http.Handler) {
	s.app = httptest.NewUnstartedServer(s.routes(codex, router))
	s.app.Listener.Close()
	s.app.Listener = listener
	s.app.Start()
}

func (s *Suite) routes(codex, openrouter http.Handler) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/settings/modelle/codex/device/", codex)
	mux.Handle("/api/settings/modellanbieter/openrouter", openrouter)
	mux.Handle("/api/settings/modellanbieter/openrouter/", openrouter)
	mux.Handle("/settings/modellanbieter/openrouter", openrouter)
	return mux
}

func (s *Suite) cleanup(ctx context.Context, _ *godog.Scenario, _ error) (context.Context, error) {
	s.stopProcess()
	if s.db != nil {
		_ = s.db.Close()
	}
	if s.app != nil {
		s.app.Close()
	}
	if s.issuer != nil {
		s.issuer.Close()
	}
	if s.provider != nil {
		s.provider.Close()
	}
	return ctx, nil
}

func (s *Suite) seedCodex(expiry time.Duration) error {
	bundle := modellverbindung.TokenBundle{IDToken: s.issuerState.idToken(), AccessToken: priorToken, RefreshToken: priorRefresh, AccountID: accountID, ExpiresAt: time.Now().Add(expiry)}
	encoded, err := json.Marshal(bundle)
	if err != nil {
		return err
	}
	return s.store.Save(context.Background(), modellverbindung.CredentialKey, encoded)
}

func (s *Suite) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/health" {
		w.WriteHeader(http.StatusOK)
		return
	}
	http.NotFound(w, r)
}

func (s *Suite) Render(w io.Writer, _ string, _ map[string]any) error {
	_, err := io.WriteString(w, "controlled renderer")
	return err
}
