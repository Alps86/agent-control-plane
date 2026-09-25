package organisationswechsel

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"

	"agentcontrolplane/app/internal/adapter/sqlite"
	weborganisation "agentcontrolplane/app/internal/adapter/web/organisation"
	webwechsel "agentcontrolplane/app/internal/adapter/web/organisationswechsel"
	apporganisation "agentcontrolplane/app/internal/app/organisation"
	appregel "agentcontrolplane/app/internal/app/organisationsregel"
	"agentcontrolplane/app/internal/domain/rechte"
	"agentcontrolplane/ui/bridge"
)

func (i *FixedIdentity) Actors() []rechte.Actor { return []rechte.Actor{i.Actor} }

func (s *Suite) foreignIdentity(_ string) error {
	if err := s.switchTo("Nordstern"); err != nil {
		return err
	}
	if s.foreignClose != nil {
		s.foreignClose()
		s.foreignClose = nil
	}
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
	organizations := apporganisation.NewService(sqlite.NewOrganizationStore(db), identity)
	rules := appregel.NewService(sqlite.NewOrganizationStore(db), identity, db)
	server := httptest.NewUnstartedServer(nil)
	mux := http.NewServeMux()
	old := weborganisation.NewHandler(organizations, ui, server.Listener.Addr().String())
	next := webwechsel.NewHandler(organizations, rules, ui, server.Listener.Addr().String())
	for _, pattern := range []string{"/api/organisationen", "/api/organisationen/", "/organisationen", "/organisationen/"} {
		mux.Handle(pattern, old)
	}
	for _, pattern := range []string{"GET /api/organisationswechsel", "GET /organisationswechsel", "POST /organisationswechsel/", "POST /api/organisationen/{id}/wechsel", "GET /api/organisationen/{id}/arbeitsregeln", "PUT /api/organisationen/{id}/arbeitsregeln", "POST /api/organisationen/{id}/arbeitsregeln/vorschau", "GET /organisationen/{id}/arbeitsregeln", "POST /organisationen/{id}/arbeitsregeln", "POST /organisationen/{id}/arbeitsregeln/vorschau"} {
		mux.Handle(pattern, next)
	}
	server.Config.Handler = mux
	server.Start()
	s.foreignServer = server.URL
	s.foreignClose = func() { server.Close(); _ = db.Close() }
	return nil
}

func (s *Suite) foreignCall(method, path string, body string) (Response, error) {
	req, err := http.NewRequest(method, s.foreignServer+path, strings.NewReader(body))
	if err != nil {
		return Response{}, err
	}
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	r, err := s.client.Do(req)
	if err != nil {
		return Response{}, err
	}
	defer r.Body.Close()
	data, err := io.ReadAll(r.Body)
	return Response{Status: r.StatusCode, Body: data, Header: r.Header.Clone()}, err
}

func (s *Suite) foreignRoutes() error {
	org, err := s.organization("Südwind")
	if err != nil {
		return err
	}
	paths := []string{"/api/organisationen/" + url.PathEscape(org.ID), "/organisationen/" + url.PathEscape(org.ID), "/api/organisationen/" + url.PathEscape(org.ID) + "/wechsel"}
	methods := []string{"GET", "GET", "POST"}
	s.foreignReplies = nil
	s.unknownReplies = nil
	for index, path := range paths {
		foreign, err := s.foreignCall(methods[index], path, "")
		if err != nil {
			return err
		}
		s.foreignReplies = append(s.foreignReplies, foreign)
		unknownPath := strings.Replace(path, org.ID, "unbekannte-kennung", 1)
		unknown, err := s.foreignCall(methods[index], unknownPath, "")
		if err != nil {
			return err
		}
		s.unknownReplies = append(s.unknownReplies, unknown)
	}
	return nil
}

func (s *Suite) foreignChange(name string) error {
	org, err := s.organization(name)
	if err != nil {
		return err
	}
	path := "/api/organisationen/" + url.PathEscape(org.ID) + "/arbeitsregeln"
	unknown := "/api/organisationen/unbekannte-kennung/arbeitsregeln"
	body := `{"area":"delegation","from":"requested","to":"approved","approver":"betreiber","revision":0}`
	s.foreignReplies = nil
	s.unknownReplies = nil
	for _, attempt := range []struct{ method, suffix, body string }{{"POST", "/vorschau", body}, {"PUT", "", body}} {
		foreign, err := s.foreignCall(attempt.method, path+attempt.suffix, attempt.body)
		if err != nil {
			return err
		}
		unk, err := s.foreignCall(attempt.method, unknown+attempt.suffix, attempt.body)
		if err != nil {
			return err
		}
		s.foreignReplies = append(s.foreignReplies, foreign)
		s.unknownReplies = append(s.unknownReplies, unk)
	}
	return nil
}

func (s *Suite) noForeignData(name string) error {
	org, err := s.organization(name)
	if err != nil {
		return err
	}
	for _, reply := range s.foreignReplies {
		if reply.Status != 404 {
			return fmt.Errorf("Fremdreferenz Status %d: %s", reply.Status, reply.Body)
		}
		for _, secret := range []string{org.ID, org.Name, org.Description} {
			if strings.Contains(string(reply.Body), secret) {
				return fmt.Errorf("Fremddaten in 404: %s", reply.Body)
			}
		}
	}
	return nil
}

func (s *Suite) sameUnknown() error {
	if len(s.foreignReplies) == 0 || len(s.foreignReplies) != len(s.unknownReplies) {
		return fmt.Errorf("Vergleichsantworten fehlen")
	}
	for index, foreign := range s.foreignReplies {
		unknown := s.unknownReplies[index]
		if foreign.Status != 404 || foreign.Status != unknown.Status || string(foreign.Body) != string(unknown.Body) {
			return fmt.Errorf("Fremd/unbekannt verschieden: %d %q / %d %q", foreign.Status, foreign.Body, unknown.Status, unknown.Body)
		}
	}
	return nil
}
