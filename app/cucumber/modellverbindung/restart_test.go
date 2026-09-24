package modellverbindung

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"agentcontrolplane/app/internal/adapter/credentials"
	"agentcontrolplane/app/internal/adapter/model/codexauth"
	"agentcontrolplane/app/internal/adapter/web/modellanbieter"
	appflow "agentcontrolplane/app/internal/app/modellverbindung"
	"github.com/cucumber/godog"
)

func (s *Suite) registerRestartSteps(sc *godog.ScenarioContext) {
	sc.Step(`^ich die lokale Settings-Grenze mit demselben geschützten Speicher neu aufbaue$`, s.rebuildSettings)
	sc.Step(`^wird die Codex-Abo-Verbindung ohne neuen Gerätecode als verbunden angezeigt$`, s.connectedAfterRestart)
	sc.Step(`^der serverseitige Zugangs-Port liefert dasselbe Token und dieselbe Konto-ID$`, s.sameAccessAfterRestart)
}

func (s *Suite) rebuildSettings() error {
	if err := s.reopenConnectionStore(); err != nil {
		return err
	}
	provider, err := codexauth.NewDirect(codexauth.Config{Issuer: s.issuer.server.URL, ClientID: codexauth.CodexCLIClientID, HTTPClient: s.issuer.server.Client()})
	if err != nil {
		return err
	}
	s.service = appflow.New(provider, s.store)
	handler, err := modellanbieter.New(s.service, "127.0.0.1:8080")
	if err != nil {
		return err
	}
	s.handler = handler.Handler()
	return nil
}

func (s *Suite) reopenConnectionStore() error {
	var err error
	s.priorBundle, err = s.store.Load(context.Background(), appflow.CredentialKey)
	if err != nil {
		return err
	}
	s.store, err = credentials.NewStore(s.secretPath, s.keyPath)
	return err
}

func (s *Suite) connectedAfterRestart() error {
	if err := s.expectState("connected", ""); err != nil {
		return err
	}
	if s.response.code != http.StatusOK || s.response.data.UserCode != "" || s.response.data.VerificationURL != "" || s.issuer.starts != 1 {
		return fmt.Errorf("reconstructed Settings connection lost or issued another device code")
	}
	return nil
}

func (s *Suite) sameAccessAfterRestart() error {
	var prior appflow.TokenBundle
	if err := json.Unmarshal(s.priorBundle, &prior); err != nil {
		return err
	}
	if err := s.requestAccess(); err != nil {
		return err
	}
	if s.accessErr != nil || s.accessToken == "" || s.accessToken != prior.AccessToken || s.accountID != prior.AccountID {
		return fmt.Errorf("reconstructed server-side resolver changed token/account snapshot")
	}
	return s.tokenBundleEncrypted()
}
