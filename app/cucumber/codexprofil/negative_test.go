package codexprofil

import (
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"

	"agentcontrolplane/app/internal/adapter/runtime/codexprofil"
	"agentcontrolplane/app/internal/adapter/sqlite"
	webprofil "agentcontrolplane/app/internal/adapter/web/codexprofil"
	appagent "agentcontrolplane/app/internal/app/agent"
	appcodexprofil "agentcontrolplane/app/internal/app/codexprofil"
	apporganisation "agentcontrolplane/app/internal/app/organisation"
	"agentcontrolplane/app/internal/domain/rechte"
)

type UnassignedIdentity struct{}

func (UnassignedIdentity) Actors() []rechte.Actor { return nil }

func (s *Suite) unassignedService() (*appcodexprofil.Service, *sqlite.Database, error) {
	db, err := sqlite.OpenApplication(context.Background(), s.dbPath)
	if err != nil {
		return nil, nil, err
	}
	service := appcodexprofil.NewService(db, db, UnassignedIdentity{},
		sqlite.NewOrganizationStore(db), codexprofil.NewLocator(s.dbPath))
	return service, db, nil
}

func (s *Suite) ambiguousIdentity() error {
	service, db, err := s.unassignedService()
	if err != nil {
		return err
	}
	s.negativeDB = db
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		return err
	}
	s.negative = &httptest.Server{Listener: listener, Config: &http.Server{Handler: http.NotFoundHandler()}}
	address := s.negative.Listener.Addr().String()
	agents := appagent.NewService(db, apporganisation.NewLocalIdentity(), sqlite.NewOrganizationStore(db))
	s.negative.Config.Handler = webprofil.NewHandler(service, agents, nil, address, nil)
	s.negative.Start()
	return s.negativeWrite()
}

func (s *Suite) negativeWrite() error {
	input, _ := json.Marshal(map[string]bool{"workspace_enabled": true, "write_enabled": true})
	request, err := http.NewRequest("PUT", s.negative.URL+s.profilePath(), strings.NewReader(string(input)))
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
	s.response = Response{Method: "PUT", Status: response.StatusCode, Body: data}
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
