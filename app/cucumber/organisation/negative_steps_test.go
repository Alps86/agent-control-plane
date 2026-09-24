package organisation

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"

	"agentcontrolplane/app/internal/adapter/sqlite"
	weborganisation "agentcontrolplane/app/internal/adapter/web/organisation"
	apporganisation "agentcontrolplane/app/internal/app/organisation"
	"agentcontrolplane/app/internal/domain/rechte"
	"agentcontrolplane/ui/bridge"
)

func (s *Suite) noIdentity() error {
	db, err := sqlite.OpenWithMigrations(context.Background(), s.dbPath, sqlite.RunMigration(2), sqlite.OrganizationMigration(), sqlite.GoalMigration(), sqlite.AgentMigration(), sqlite.ProjectMigration())
	if err != nil {
		return err
	}
	service := apporganisation.NewService(sqlite.NewOrganizationStore(db), nil)
	ui, err := bridge.New()
	if err != nil {
		_ = db.Close()
		return err
	}
	server := httptest.NewUnstartedServer(nil)
	server.Config.Handler = weborganisation.NewHandler(service, ui, server.Listener.Addr().String())
	server.Start()
	s.negative = &NegativeServer{DB: db, Server: server}
	return nil
}

func (s *Suite) deniedList() error {
	for _, path := range []string{"/api/organisationen", "/api/organisationen?actor=local-operator&organization=" + s.organization.ID, "/api/organisationen/" + s.organization.ID} {
		response, err := s.negativeRequest("GET", path, "")
		if err != nil {
			return err
		}
		if err := s.assertDenied(response); err != nil {
			return err
		}
	}
	return nil
}

func (s *Suite) deniedCreate() error {
	body := `{"name":"Spoofed","description":"Must not appear"}`
	response, err := s.negativeRequest("POST", "/api/organisationen?actor=local-operator", body)
	if err != nil {
		return err
	}
	if err := s.assertDenied(response); err != nil {
		return err
	}
	return s.descriptionAbsent("Must not appear")
}

func (s *Suite) negativeRequest(method, path, body string) (*HTTPResponse, error) {
	req, err := http.NewRequest(method, s.negative.Server.URL+path, strings.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Actor-ID", "local-operator")
	req.Header.Set("X-Operator-ID", "local-operator")
	response, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	data, err := io.ReadAll(response.Body)
	return &HTTPResponse{Status: response.StatusCode, Body: data}, err
}

func (s *Suite) assertDenied(response *HTTPResponse) error {
	if response.Status != http.StatusForbidden {
		return fmt.Errorf("Identität: Status %d: %s", response.Status, response.Body)
	}
	var failure FieldError
	if err := json.Unmarshal(response.Body, &failure); err != nil {
		return err
	}
	if failure.Error != "access_denied" {
		return fmt.Errorf("Identität: %s", response.Body)
	}
	if strings.Contains(string(response.Body), s.organization.Name) {
		return fmt.Errorf("Organisationsdaten im 403: %s", response.Body)
	}
	return nil
}

func (s *Suite) stopNegative() {
	if s.negative == nil {
		return
	}
	s.negative.Server.Close()
	_ = s.negative.DB.Close()
	s.negative = nil
}

func (i *FixedIdentity) Actors() []rechte.Actor { return []rechte.Actor{i.Actor} }

func (s *Suite) getUnknown() error {
	return s.request("GET", "/api/organisationen/gibt-es-nicht", "", "")
}

func (s *Suite) unknown404() error {
	if err := s.assertNotFound(s.response); err != nil {
		return err
	}
	s.unknown = s.response
	return nil
}

func (s *Suite) getUnassigned() error {
	db, err := sqlite.OpenWithMigrations(context.Background(), s.dbPath, sqlite.RunMigration(2), sqlite.OrganizationMigration(), sqlite.GoalMigration(), sqlite.AgentMigration(), sqlite.ProjectMigration())
	if err != nil {
		return err
	}
	identity := &FixedIdentity{Actor: rechte.NewActor("other-operator", rechte.Operator)}
	service := apporganisation.NewService(sqlite.NewOrganizationStore(db), identity)
	ui, err := bridge.New()
	if err != nil {
		_ = db.Close()
		return err
	}
	server := httptest.NewUnstartedServer(nil)
	server.Config.Handler = weborganisation.NewHandler(service, ui, server.Listener.Addr().String())
	server.Start()
	s.negative = &NegativeServer{DB: db, Server: server}
	return s.readUnassigned()
}

func (s *Suite) readUnassigned() error {
	response, err := s.facadeGet("/api/organisationen/" + s.organization.ID)
	if err != nil {
		return err
	}
	s.response = response
	return nil
}

func (s *Suite) same404() error {
	if err := s.assertNotFound(s.response); err != nil {
		return err
	}
	if s.unknown == nil {
		return fmt.Errorf("Vergleichsantwort fehlt")
	}
	if s.unknown.Status != s.response.Status || string(s.unknown.Body) != string(s.response.Body) {
		return fmt.Errorf("404-Antworten verschieden: %q / %q", s.unknown.Body, s.response.Body)
	}
	return nil
}

func (s *Suite) facadeGet(path string) (*HTTPResponse, error) {
	response, err := s.client.Get(s.negative.Server.URL + path)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	return &HTTPResponse{Status: response.StatusCode, Body: body}, err
}

func (s *Suite) assertNotFound(response *HTTPResponse) error {
	if response == nil || response.Status != http.StatusNotFound {
		return fmt.Errorf("kein 404: %+v", response)
	}
	var failure FieldError
	if err := json.Unmarshal(response.Body, &failure); err != nil {
		return err
	}
	if failure.Error != "not_found" || strings.Contains(string(response.Body), s.organization.Name) || strings.Contains(string(response.Body), s.organization.ID) || strings.Contains(string(response.Body), s.organization.Description) {
		return fmt.Errorf("404 enthält Organisationsdaten: %s", response.Body)
	}
	return nil
}
