package openrouterverbindung

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	credentialsadapter "agentcontrolplane/app/internal/adapter/credentials"
	modelopenrouter "agentcontrolplane/app/internal/adapter/model/openrouter"
	webopenrouter "agentcontrolplane/app/internal/adapter/web/openrouter"
	"agentcontrolplane/app/internal/app/openrouterverbindung"
	"agentcontrolplane/ui/bridge"
	"github.com/cucumber/godog"
)

func NewSuite(t *testing.T) *Suite {
	return &Suite{t: t}
}

func (s *Suite) newStore() error {
	s.storeDir = filepath.Join(s.t.TempDir(), "credentials")
	return nil
}

func (s *Suite) startApplication() error {
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		return err
	}

	handler, err := s.newHandler(listener.Addr().String())
	if err != nil {
		listener.Close()
		return err
	}

	s.productHandler = handler
	server := httptest.NewUnstartedServer(http.HandlerFunc(s.observeRequest))
	server.Listener.Close()
	server.Listener = listener
	server.Start()
	s.app = server
	return nil
}

func (s *Suite) newHandler(bindAddress string) (http.Handler, error) {
	store, err := credentialsadapter.NewStore(filepath.Join(s.storeDir, "secrets.enc"), filepath.Join(s.storeDir, "master.key"))
	if err != nil {
		return nil, err
	}

	renderer, err := bridge.New()
	if err != nil {
		return nil, err
	}

	probe := modelopenrouter.NewProbeAt(nil, s.provider.URL+"/api/v1/key")
	service := openrouterverbindung.NewService(store, probe)
	return webopenrouter.NewHandler(service, renderer, bindAddress), nil
}

func (s *Suite) observeRequest(w http.ResponseWriter, r *http.Request) {
	observed := s.mutationArrived != nil && r.Header.Get("X-Race-Probe") == "1"
	if observed {
		select {
		case s.mutationArrived <- struct{}{}:
		default:
		}
	}

	s.productHandler.ServeHTTP(w, r)
	if observed {
		s.mutationHandled <- struct{}{}
	}
}

func (s *Suite) restartApplication() error {
	s.app.Close()
	return s.startApplication()
}

func (s *Suite) cleanup(ctx context.Context, _ *godog.Scenario, _ error) (context.Context, error) {
	if s.probeRelease != nil {
		s.releaseOnce.Do(s.closeProbe)
	}

	if s.app != nil {
		s.app.Close()
	}

	if s.provider != nil {
		s.provider.Close()
	}

	if s.remoteServer != nil {
		s.remoteServer.Close()
	}

	s.resetScenarioState()
	return ctx, nil
}

func (s *Suite) resetScenarioState() {
	s.app, s.provider = nil, nil
	s.remoteServer, s.remotePeerAddr = nil, nil
	s.providerCalls = nil
	s.responseBody = nil
	s.currentKey, s.previousKey = "", ""
	s.probeEntered, s.probeRelease, s.mutationArrived, s.mutationHandled = nil, nil, nil, nil
	s.checkDone, s.mutationDone = nil, nil
}

func (s *Suite) readyMockProvider() error {
	s.startProvider()
	return s.startApplication()
}
