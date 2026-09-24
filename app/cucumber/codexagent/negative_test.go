package codexagent

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"

	"agentcontrolplane/app/internal/adapter/agent/codexcli"
	"agentcontrolplane/app/internal/adapter/sqlite"
	webagent "agentcontrolplane/app/internal/adapter/web/agent"
	appagent "agentcontrolplane/app/internal/app/agent"
)

func (s *Suite) startNegative() error {
	db, err := sqlite.OpenWithMigrations(context.Background(), s.dbPath, sqlite.RunMigration(2), sqlite.OrganizationMigration(), sqlite.GoalMigration(), sqlite.AgentMigration(), sqlite.ProjectMigration(), sqlite.DataScopeMigration())
	if err != nil {
		return err
	}
	s.negativeDB = db
	service := appagent.NewService(db, UnassignedIdentity{}, sqlite.NewOrganizationStore(db))
	if !service.RegisterTemplate(codexcli.NewTemplate()) {
		return fmt.Errorf("Codex-Vorlage nicht registriert")
	}
	s.negative = httptest.NewUnstartedServer(http.NotFoundHandler())
	s.negative.Config.Handler = webagent.NewHandler(service, nil, s.negative.Listener.Addr().String())
	s.negative.Start()
	return nil
}

func (s *Suite) postUnassigned(name string) error {
	body, err := json.Marshal(map[string]string{"name": name, "template_id": "codex-cli", "execution_kind": "codex_cli"})
	if err != nil {
		return err
	}
	request, err := http.NewRequest("POST", s.negative.URL+s.agentsPath(), strings.NewReader(string(body)))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := s.client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	data, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}
	s.response = Response{Status: response.StatusCode, Body: data, Location: response.Header.Get("Location")}
	return nil
}
