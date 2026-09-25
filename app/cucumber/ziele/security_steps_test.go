package ziele

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"

	"agentcontrolplane/app/internal/adapter/sqlite"
	webziel "agentcontrolplane/app/internal/adapter/web/ziel"
	apporganisation "agentcontrolplane/app/internal/app/organisation"
	appziel "agentcontrolplane/app/internal/app/ziel"
	"agentcontrolplane/app/internal/domain/rechte"
	portorganisation "agentcontrolplane/app/internal/port/organisation"
)

func (s *Suite) notFound() error {
	if err := s.assertNotFound(); err != nil {
		return err
	}

	s.unknownBody = append([]byte(nil), s.response.Body...)
	return nil
}

func (s *Suite) sameNotFound() error {
	if err := s.assertNotFound(); err != nil {
		return err
	}

	if len(s.unknownBody) == 0 || string(s.unknownBody) != string(s.response.Body) {
		return fmt.Errorf("404-Antworten verschieden: %q / %q", s.unknownBody, s.response.Body)
	}

	return nil
}

func (s *Suite) assertNotFound() error {
	if err := s.statusCode(http.StatusNotFound); err != nil {
		return err
	}

	var failure FieldError
	if err := json.Unmarshal(s.response.Body, &failure); err != nil {
		return err
	}

	if failure.Error != "not_found" || s.notFoundLeaksData() {
		return fmt.Errorf("404 mit Daten: %s", s.response.Body)
	}

	return nil
}

func (s *Suite) notFoundLeaksData() bool {
	body := string(s.response.Body)
	if s.active == "" {
		return false
	}

	return strings.Contains(body, s.active) || strings.Contains(body, s.organizations[s.active])
}

func (s *Suite) noGoal() error {
	for name := range s.organizations {
		if err := s.goalsEmpty(name); err != nil {
			return err
		}
	}

	return nil
}

func (s *Suite) noIdentity() error {
	return s.startNegative(nil)
}

func (s *Suite) startNegative(identity *FixedIdentity) error {
	s.stopNegative()
	db, err := sqlite.OpenApplication(context.Background(), s.database)
	if err != nil {
		return err
	}

	var source portorganisation.Identity
	if identity != nil {
		source = identity
	}

	organizations := apporganisation.NewService(sqlite.NewOrganizationStore(db), source)
	goals := appziel.NewService(db, organizations)
	server := httptest.NewUnstartedServer(nil)
	server.Config.Handler = webziel.NewHandler(goals, organizations, nil, server.Listener.Addr().String())
	server.Start()
	s.negative = &NegativeServer{DB: db, Server: server}
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

func (s *Suite) negativeRequest(method, path, body string) error {
	request, err := http.NewRequest(method, s.negative.Server.URL+path, strings.NewReader(body))
	if err != nil {
		return err
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Actor-ID", "local-operator")
	request.Header.Set("X-Operator-ID", "local-operator")
	response, err := s.client.Do(request)
	if err != nil {
		return err
	}

	return s.recordNegative(response)
}

func (s *Suite) recordNegative(response *http.Response) error {
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	s.response = &HTTPResponse{Status: response.StatusCode, Body: body}
	return err
}

func (s *Suite) deniedGoalList(name string) error {
	if err := s.negativeRequest("GET", s.goalPath(name)+"?actor=local-operator", ""); err != nil {
		return err
	}

	return s.assertDenied()
}

func (s *Suite) deniedGoalCreate(name string) error {
	if err := s.negativeRequest("POST", s.goalPath(name)+"?actor=local-operator", `{"name":"Verbotenes Ziel"}`); err != nil {
		return err
	}

	if err := s.assertDenied(); err != nil {
		return err
	}

	return s.goalsEmpty(name)
}

func (s *Suite) assertDenied() error {
	if err := s.statusCode(http.StatusForbidden); err != nil {
		return err
	}

	var failure FieldError
	if err := json.Unmarshal(s.response.Body, &failure); err != nil {
		return err
	}

	if failure.Error != "access_denied" || strings.Contains(string(s.response.Body), s.active) {
		return fmt.Errorf("403 mit Daten: %s", s.response.Body)
	}

	return nil
}

func (s *Suite) unassignedGoalList(name string) error {
	identity := &FixedIdentity{Actor: rechte.NewActor("other-operator", rechte.Operator)}
	if err := s.startNegative(identity); err != nil {
		return err
	}

	if err := s.negativeRequest("GET", "/api/organisationen/unbekannt/ziele", ""); err != nil {
		return err
	}

	s.unknownBody = append([]byte(nil), s.response.Body...)
	s.active = name
	return s.negativeRequest("GET", s.goalPath(name), "")
}

func (s *Suite) unassignedGoalCreate(name string) error {
	s.active = name
	return s.negativeRequest("POST", s.goalPath(name), `{"name":"Verbotenes Ziel"}`)
}

func (s *Suite) foreignOrigin(origin, goal, name string) error {
	return s.foreignWrite(goal, name, origin, "")
}

func (s *Suite) foreignHost(host, goal, name string) error {
	return s.foreignWrite(goal, name, "", host)
}

func (s *Suite) foreignWrite(goal, name, origin, host string) error {
	body, err := json.Marshal(map[string]string{"name": goal})
	if err != nil {
		return err
	}

	header := http.Header{"Content-Type": []string{"application/json"}}
	if origin != "" {
		header.Set("Origin", origin)
	}

	return s.request("POST", s.goalPath(name), string(body), header, host)
}

func (s *Suite) restartWildcard() error {
	s.stopServer()
	s.bindHost = "0.0.0.0"
	return s.startServer()
}

func (s *Suite) wildcardConfigRejected() error {
	if !s.wildcardRejected {
		return fmt.Errorf("wildcard startup lacked APP_ADDR loopback rejection")
	}
	return nil
}

func (s *Suite) noWildcardListener() error {
	response, err := s.client.Get(s.baseURL() + "/health")
	if err != nil {
		return nil
	}
	response.Body.Close()
	return fmt.Errorf("wildcard startup left HTTP listener reachable")
}

func (s *Suite) restartLoopback() error {
	s.stopServer()
	s.bindHost = "127.0.0.1"
	return s.startServer()
}
