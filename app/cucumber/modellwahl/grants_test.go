package modellwahl

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func (s *Suite) grantPath() string {
	return "/api/organisationen/" + s.orgID + "/modellfreigabe/openrouter"
}

func (s *Suite) openRouterReady() error {
	if s.providerServer == nil {
		return fmt.Errorf("kontrollierter lokaler Probe-Anbieter fehlt")
	}
	if err := s.request(http.MethodPost, "/api/settings/modellanbieter/openrouter", `{"key":"`+s.secret+`"}`); err != nil {
		return err
	}
	if s.last.Status != 200 {
		return fmt.Errorf("Settings-Save HTTP %d: %s", s.last.Status, s.last.Body)
	}
	s.secretStored = true
	if err := s.noSecrets(); err != nil {
		return err
	}
	return s.checkOpenRouter()
}

func (s *Suite) checkOpenRouter() error {
	before := s.providerCalls.Load()
	if err := s.request(http.MethodPost, "/api/settings/modellanbieter/openrouter/pruefen", ""); err != nil {
		return err
	}
	if s.last.Status != 200 {
		return fmt.Errorf("Settings-Check HTTP %d: %s", s.last.Status, s.last.Body)
	}
	view := OpenRouterStatus{}
	if err := json.Unmarshal(s.last.Body, &view); err != nil {
		return err
	}
	if s.blockedExternal.Load() != 0 {
		return fmt.Errorf("Live-Provider-Route blockiert; lokale Probe nicht verdrahtet")
	}
	if view.Reference != "openrouter-central" || view.Status != "einsatzbereit" || s.providerCalls.Load() != before+1 {
		return fmt.Errorf("kontrollierte OpenRouter-Prüfung fehlt: %+v, lokale Aufrufe %d", view, s.providerCalls.Load()-before)
	}
	return s.noSecrets()
}

func (s *Suite) grantOrganization() error {
	if err := s.request(http.MethodPost, s.grantPath()+"/organisation", ""); err != nil {
		return err
	}
	if s.last.Status != 200 {
		return fmt.Errorf("Organisationsfreigabe HTTP %d: %s", s.last.Status, s.last.Body)
	}
	return s.noSecrets()
}

func (s *Suite) grantAgent() error {
	if err := s.request(http.MethodPost, s.grantPath()+"/agenten/"+s.agentID, ""); err != nil {
		return err
	}
	if s.last.Status != 200 {
		return fmt.Errorf("Agentenfreigabe HTTP %d: %s", s.last.Status, s.last.Body)
	}
	return s.noSecrets()
}

func (s *Suite) grantStatus() (GrantStatus, error) {
	status := GrantStatus{}
	if err := s.request(http.MethodGet, s.grantPath()+"/agenten/"+s.agentID, ""); err != nil {
		return status, err
	}
	if s.last.Status != 200 {
		return status, fmt.Errorf("Freigabestatus HTTP %d: %s", s.last.Status, s.last.Body)
	}
	if err := json.Unmarshal(s.last.Body, &status); err != nil {
		return status, err
	}
	return status, nil
}

func (s *Suite) openRouterGranted() error {
	if err := s.openRouterReady(); err != nil {
		return err
	}
	if err := s.grantOrganization(); err != nil {
		return err
	}
	if err := s.grantAgent(); err != nil {
		return err
	}
	status, err := s.grantStatus()
	if err != nil {
		return err
	}
	if !status.Allowed || !status.OrganizationGranted || !status.AgentGranted || status.Reference != "openrouter-central" {
		return fmt.Errorf("öffentliche Freigaben unvollständig: %+v", status)
	}
	return s.noSecrets()
}

func (s *Suite) missingGrant(level string) error {
	if level != "Organisation" && level != "Agent" {
		return fmt.Errorf("unbekannte Freigabeebene %q", level)
	}
	if level == "Agent" {
		if err := s.grantOrganization(); err != nil {
			return err
		}
	}
	status, err := s.grantStatus()
	if err != nil {
		return err
	}
	return s.assertMissingGrant(level, status)
}

func (s *Suite) assertMissingGrant(level string, status GrantStatus) error {
	if status.Allowed || status.AgentGranted || status.OrganizationGranted != (level == "Agent") {
		return fmt.Errorf("falsche fehlende %s-Freigabe: %+v", level, status)
	}
	reason := "Organisationsfreigabe fehlt"
	if level == "Agent" {
		reason = "Agentenfreigabe fehlt"
	}
	if status.Reason != reason {
		return fmt.Errorf("fehlender Freigabegrund: %+v", status)
	}
	return s.noSecrets()
}
