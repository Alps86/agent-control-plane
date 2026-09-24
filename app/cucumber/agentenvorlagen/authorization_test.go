package agentenvorlagen

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"agentcontrolplane/app/internal/adapter/sqlite"
	webagent "agentcontrolplane/app/internal/adapter/web/agent"
	appagent "agentcontrolplane/app/internal/app/agent"
	apporganisation "agentcontrolplane/app/internal/app/organisation"
	"agentcontrolplane/app/internal/domain/rechte"
)

func (i *FixedIdentity) Actors() []rechte.Actor { return []rechte.Actor{i.Actor} }

func (s *Suite) startNegative() error {
	if s.negative != nil {
		return nil
	}
	db, err := sqlite.OpenWithMigrations(context.Background(), s.dbPath,
		sqlite.RunMigration(2), sqlite.OrganizationMigration(), sqlite.RunMigration(4), sqlite.AgentMigration(), sqlite.ProjectMigration())
	if err != nil {
		return err
	}
	s.negativeDB = db
	identity := &FixedIdentity{Actor: rechte.NewActor("other-operator", rechte.Operator)}
	service := appagent.NewService(db, identity, sqlite.NewOrganizationStore(db))
	server := httptest.NewUnstartedServer(nil)
	server.Config.Handler = webagent.NewHandler(service, nil, server.Listener.Addr().String())
	server.Start()
	s.negative = server
	return nil
}

func (s *Suite) stopNegative() {
	if s.negative != nil {
		s.negative.Close()
		s.negative = nil
	}
	if s.negativeDB != nil {
		_ = s.negativeDB.Close()
		s.negativeDB = nil
	}
}

func (s *Suite) negativeRequest(method, path, body string) error {
	req, err := http.NewRequest(method, s.negative.URL+path, strings.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	response, err := s.client.Do(req)
	if err != nil {
		return err
	}
	return s.recordResponse(response)
}

func (s *Suite) unassignedCreate(org string) error {
	if err := s.startNegative(); err != nil {
		return err
	}
	body, _ := json.Marshal(map[string]string{"name": "Unbefugt", "template_id": "recherche", "execution_kind": "eino"})
	return s.negativeRequest("POST", s.orgPath(org), string(body))
}

func (s *Suite) unassignedDenied() error {
	if s.response.Status != http.StatusNotFound && s.response.Status != http.StatusForbidden {
		return fmt.Errorf("fremde Anlage nicht verweigert: %d: %s", s.response.Status, s.response.Body)
	}
	return s.noAgentFromRequest("Nordstern")
}

func (s *Suite) unassignedList(org string) error {
	return s.negativeRequest("GET", s.orgPath(org), "")
}

func (s *Suite) noAgentData() error {
	if s.response.Status != http.StatusNotFound && s.response.Status != http.StatusForbidden {
		return fmt.Errorf("fremde Liste lesbar: %d: %s", s.response.Status, s.response.Body)
	}
	if strings.Contains(string(s.response.Body), "Unbefugt") || strings.Contains(string(s.response.Body), "Nordstern") {
		return fmt.Errorf("fremde Daten in Antwort: %s", s.response.Body)
	}
	return nil
}

func (s *Suite) remotePeerSpoofedPost() error {
	db, err := sqlite.OpenWithMigrations(context.Background(), s.dbPath,
		sqlite.RunMigration(2), sqlite.OrganizationMigration(), sqlite.RunMigration(4), sqlite.AgentMigration(), sqlite.ProjectMigration())
	if err != nil {
		return err
	}
	defer db.Close()
	service := appagent.NewService(db, apporganisation.NewLocalIdentity(), sqlite.NewOrganizationStore(db))
	server := httptest.NewUnstartedServer(nil)
	defer server.Close()
	handler := webagent.NewHandler(service, nil, server.Listener.Addr().String())
	return s.serveRemoteSpoof(handler, server)
}

func (s *Suite) serveRemoteSpoof(handler http.Handler, server *httptest.Server) error {
	address := server.Listener.Addr().String()
	body := `{"name":"Fern","template_id":"recherche","execution_kind":"eino"}`
	req := httptest.NewRequest("POST", "http://localhost:"+strings.Split(address, ":")[1]+s.orgPath("Nordstern"), strings.NewReader(body))
	req.RemoteAddr = "203.0.113.8:41234"
	req = req.WithContext(context.WithValue(req.Context(), http.LocalAddrContextKey, server.Listener.Addr()))
	req.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, req)
	return s.recordResponse(response.Result())
}
