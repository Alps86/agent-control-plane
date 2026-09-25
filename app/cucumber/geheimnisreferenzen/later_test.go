package geheimnisreferenzen

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	appstatus "agentcontrolplane/app/internal/app/verbindungsstatus"
)

func (LaterPort) Read(context.Context) (appstatus.State, error) {
	return appstatus.State{Status: "not_ready"}, nil
}

func (s *Suite) registerLaterProvider() error {
	s.later = true
	return s.restart()
}

func (s *Suite) laterProviderAuth() error { return s.auth("later", "api_key") }

func (s *Suite) noOtherAuthAction() error {
	view, err := s.statusView()
	if err != nil {
		return err
	}
	provider, err := view.provider("later")
	if err != nil {
		return err
	}
	if provider.AuthType != "api_key" || provider.Ready {
		return fmt.Errorf("falsche spätere Auth-/Bereitschaft: %+v", provider)
	}
	encoded, _ := json.Marshal(provider)
	if strings.Contains(string(encoded), "device_code") || strings.Contains(string(encoded), "subscription") {
		return fmt.Errorf("späterer Anbieter zeigt andere Authart")
	}
	return nil
}
