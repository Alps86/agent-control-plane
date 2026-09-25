package berichtsweg

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"

	"agentcontrolplane/app/internal/adapter/sqlite"
	webagent "agentcontrolplane/app/internal/adapter/web/agent"
	webbericht "agentcontrolplane/app/internal/adapter/web/berichtsweg"
	appagent "agentcontrolplane/app/internal/app/agent"
	appbericht "agentcontrolplane/app/internal/app/berichtsweg"
	"agentcontrolplane/app/internal/domain/rechte"
	"agentcontrolplane/ui/bridge"
)

type FixedIdentity struct{ Actor rechte.Actor }

func (i *FixedIdentity) Actors() []rechte.Actor { return []rechte.Actor{i.Actor} }

func (s *Suite) foreignIdentity(string) error {
	db, err := sqlite.OpenApplication(context.Background(), s.database)
	if err != nil {
		return err
	}
	ui, err := bridge.New()
	if err != nil {
		_ = db.Close()
		return err
	}
	identity := &FixedIdentity{Actor: rechte.NewActor("other-operator", rechte.Operator)}
	organizations := sqlite.NewOrganizationStore(db)
	service := appbericht.NewService(db, db, organizations, identity)
	agents := appagent.NewService(db, identity, organizations)
	server := httptest.NewUnstartedServer(nil)
	mux := http.NewServeMux()
	report := webbericht.NewHandler(service, ui, server.Listener.Addr().String())
	agent := webagent.NewHandler(agents, ui, server.Listener.Addr().String(), service)
	for _, pattern := range []string{"GET /api/organisationen/{id}/berichtswege", "PUT /api/organisationen/{id}/agenten/{agentID}/berichtsweg", "GET /organisationen/{id}/berichtswege"} {
		mux.Handle(pattern, report)
	}
	for _, pattern := range []string{"/api/organisationen/{id}/agenten", "/api/organisationen/{id}/agenten/"} {
		mux.Handle(pattern, agent)
	}
	server.Config.Handler = mux
	server.Start()
	s.foreignURL = server.URL
	s.foreignClose = func() { server.Close(); _ = db.Close() }
	return nil
}

func (s *Suite) foreignCall(method, path string, body string) (Response, error) {
	request, err := http.NewRequest(method, s.foreignURL+path, strings.NewReader(body))
	if err != nil {
		return Response{}, err
	}
	if body != "" {
		request.Header.Set("Content-Type", "application/json")
	}
	response, err := s.client.Do(request)
	if err != nil {
		return Response{}, err
	}
	defer response.Body.Close()
	data, err := io.ReadAll(response.Body)
	return Response{Status: response.StatusCode, Body: data, Header: response.Header.Clone()}, err
}

func (s *Suite) compareForeign(method, path, body string) error {
	foreign, err := s.foreignCall(method, path, body)
	if err != nil {
		return err
	}
	unknown := strings.Replace(path, s.orgID("Südstern"), "unbekannte-kennung", 1)
	reference, err := s.foreignCall(method, unknown, body)
	if err != nil {
		return err
	}
	s.foreignReplies = append(s.foreignReplies, foreign)
	s.unknownReplies = append(s.unknownReplies, reference)
	return nil
}

func (s *Suite) foreignReads(org string) error {
	id := s.orgID(org)
	s.foreignReplies, s.unknownReplies = nil, nil
	if err := s.compareForeign("GET", "/api/organisationen/"+id+"/berichtswege", ""); err != nil {
		return err
	}
	return s.compareForeign("GET", "/api/organisationen/"+id+"/agenten/"+s.agentID(org, "Fremd"), "")
}

func (s *Suite) noForeignData() error {
	if len(s.foreignReplies) == 0 || len(s.foreignReplies) != len(s.unknownReplies) {
		return fmt.Errorf("Fremdantworten fehlen")
	}
	for index, reply := range s.foreignReplies {
		unknown := s.unknownReplies[index]
		if reply.Status != 404 || reply.Status != unknown.Status || string(reply.Body) != string(unknown.Body) {
			return fmt.Errorf("Fremdantwort %d unterscheidbar: %d %s / %d %s", index, reply.Status, reply.Body, unknown.Status, unknown.Body)
		}
		for _, secret := range []string{"Fremd", "Südstern", s.orgID("Südstern"), s.agentID("Südstern", "Fremd")} {
			if strings.Contains(string(reply.Body), secret) {
				return fmt.Errorf("Fremddaten: %s", reply.Body)
			}
		}
	}
	return nil
}

func (s *Suite) foreignWrite(agent string) error {
	s.foreignReplies, s.unknownReplies = nil, nil
	path := "/api/organisationen/" + s.orgID("Südstern") + "/agenten/" + s.agentID("Südstern", agent) + "/berichtsweg"
	return s.compareForeign("PUT", path, `{"parent_id":"unbekannt"}`)
}

func (s *Suite) foreignDenied(org string) error {
	if err := s.noForeignData(); err != nil {
		return err
	}
	return s.chartUnchanged(org)
}
