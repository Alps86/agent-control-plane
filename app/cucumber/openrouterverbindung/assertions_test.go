package openrouterverbindung

import (
	"fmt"
	"net/http"
	"strings"
)

func (s *Suite) expectConnection() error {
	if err := s.expectStatus(http.StatusOK); err != nil {
		return err
	}

	if !s.last.Connected || s.last.Reference != "openrouter-central" {
		return fmt.Errorf("keine einzelne zentrale Referenz: %s", s.responseBody)
	}

	return s.expectNoSecret()
}

func (s *Suite) rememberReference() error {
	if err := s.expectConnection(); err != nil {
		return err
	}

	s.firstReference = s.last.Reference
	return nil
}

func (s *Suite) expectSameReference() error {
	if err := s.expectConnection(); err != nil {
		return err
	}

	if s.last.Reference != s.firstReference {
		return fmt.Errorf("zentrale Referenz geändert: %q zu %q", s.firstReference, s.last.Reference)
	}

	return nil
}

func (s *Suite) expectNoProviderCall() error {
	if len(s.providerCalls) != 0 {
		return fmt.Errorf("unerwartete Anbieteranfragen: %d", len(s.providerCalls))
	}

	return nil
}

func (s *Suite) expectLatestKey() error {
	if len(s.providerCalls) == 0 {
		return fmt.Errorf("keine Anbieteranfrage")
	}

	last := s.providerCalls[len(s.providerCalls)-1]
	if last != "Bearer "+s.currentKey {
		return fmt.Errorf("Anbieter erhielt nicht den aktuellen Schlüssel")
	}

	return nil
}

func (s *Suite) expectNoGrants() error {
	body := strings.ToLower(string(s.responseBody))
	for _, field := range []string{"organization", "organisation", "agent", "grant", "freigabe"} {
		if strings.Contains(body, `"`+field+`"`) {
			return fmt.Errorf("Nutzungsfreigabe in Settings-Antwort: %s", field)
		}
	}

	return nil
}

func (s *Suite) getHTML() error {
	return s.send(http.MethodGet, htmlPath, "", nil)
}
