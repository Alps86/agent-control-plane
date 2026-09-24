package modellzugang

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"

	"agentcontrolplane/app/internal/adapter/credentials"
	"agentcontrolplane/app/internal/adapter/model/codexauth"
	"agentcontrolplane/app/internal/adapter/web/modellanbieter"
	webprobe "agentcontrolplane/app/internal/adapter/web/modellpruefung"
	"agentcontrolplane/app/internal/app/modellpruefung"
	"agentcontrolplane/app/internal/app/modellverbindung"
)

func (s *httpE2ESuite) setup(mode string) error {
	if !s.routeOnly || !s.probeRequired {
		return fmt.Errorf("öffentliche Abo-Probe nicht eingerichtet")
	}
	s.prepareIssuer()
	s.issuerServer = httptest.NewServer(s.issuer)
	s.t.Cleanup(s.issuerServer.Close)
	s.responses = &e2eResponses{mode: mode, firstDelta: make(chan struct{}), cancelled: make(chan struct{})}
	if mode == "echo" {
		s.responses.echoID, s.responses.echoModel = s.issuer.accountID, s.issuer.accessToken
	}
	s.nachweis = modellpruefung.NewNachweis()
	s.responseServer = httptest.NewServer(s.responses)
	s.t.Cleanup(s.responseServer.Close)
	return s.connectApp()
}

func (s *httpE2ESuite) connectApp() error {
	directory := filepath.Join(s.t.TempDir(), "credentials")
	if err := os.Mkdir(directory, 0700); err != nil {
		return err
	}
	s.secretPath = filepath.Join(directory, "connections.enc")
	store, err := credentials.NewStore(s.secretPath, filepath.Join(directory, "master.key"))
	if err != nil {
		return err
	}
	direct, err := codexauth.NewDirect(codexauth.Config{Issuer: s.issuerServer.URL,
		ClientID: codexauth.CodexCLIClientID, HTTPClient: s.issuerServer.Client()})
	if err != nil {
		return err
	}
	s.service = modellverbindung.New(direct, store)
	return s.mountApp()
}

func (s *httpE2ESuite) mountApp() error {
	mux := http.NewServeMux()
	s.appServer = httptest.NewUnstartedServer(mux)
	address := s.appServer.Listener.Addr().String()
	if err := s.mountHandlers(mux, address); err != nil {
		return err
	}
	s.appServer.Start()
	s.t.Cleanup(s.appServer.Close)
	s.client = s.appServer.Client()
	return nil
}

func (s *httpE2ESuite) mountHandlers(mux *http.ServeMux, address string) error {
	device, err := modellanbieter.New(s.service, address)
	if err != nil {
		return err
	}
	client, err := s.providerClient()
	if err != nil {
		return err
	}
	probe, err := webprobe.NewHandler(s.service, modellpruefung.NewVerbindungspruefung(s.service),
		"local-e2e-model", address, client)
	if err != nil {
		return err
	}
	mux.Handle("/settings/modelle/codex/device/", device.Handler())
	mux.Handle("/api/settings/modelle/codex/probe", probe.Handler())
	return nil
}

func (s *httpE2ESuite) providerClient() (*http.Client, error) {
	target, err := url.Parse(s.responseServer.URL)
	if err != nil {
		return nil, fmt.Errorf("lokale Responses-Route ungültig: %w", err)
	}
	return &http.Client{Transport: &e2eReroute{base: s.responseServer.Client().Transport, target: target}}, nil
}
