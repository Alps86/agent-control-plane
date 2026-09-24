package rechte

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"

	webrechte "agentcontrolplane/app/internal/adapter/web/rechte"
	domainrechte "agentcontrolplane/app/internal/domain/rechte"
	"github.com/cucumber/godog"
)

// NewSuite erzeugt einen leeren Blackbox-Szenariokontext.
func NewSuite() *Suite {
	return &Suite{}
}

func (s *Suite) InitializeScenario(sc *godog.ScenarioContext) {
	sc.Before(s.before)
	sc.After(s.after)
	sc.Step(`^der Go-Server kennt die Beispielressource "([^"]*)" mit dem Namen "([^"]*)" in Organisation "([^"]*)"$`, s.resource)
	sc.Step(`^der Server ermittelt eindeutig den Betreiber "([^"]*)" für Organisation "([^"]*)"$`, s.operator)
	sc.Step(`^der Server ermittelt eindeutig den Agenten "([^"]*)" für Organisation "([^"]*)" mit Zuordnung zur Ressource "([^"]*)"$`, s.assignedAgent)
	sc.Step(`^der Server ermittelt eindeutig den Agenten "([^"]*)" für Organisation "([^"]*)" ohne Ressourcenzuordnung$`, s.unassignedAgent)
	sc.Step(`^der Server hat (.*)$`, s.context)
	sc.Step(`^ich über HTTP GET "([^"]*)" an den Server sende$`, s.get)
	sc.Step(`^ich über HTTP GET "([^"]*)" mit diesen Headern an den Server sende:$`, s.getWithHeaders)
	sc.Step(`^antwortet der Server mit dem HTTP-Status (\d+)$`, s.expectStatus)
	sc.Step(`^die JSON-Antwort enthält die Ressource "([^"]*)" mit dem Namen "([^"]*)" und der Organisation "([^"]*)"$`, s.expectResource)
	sc.Step(`^der JSON-Fehler ist genau "([^"]*)" ohne Ressourcendaten$`, s.expectError)
	sc.Step(`^Status und Antwortkörper stimmen mit der vorherigen Antwort überein$`, s.expectSame)
}

func (s *Suite) before(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
	s.directory = &Directory{
		organizations: map[string][]string{}, resources: map[string]domainrechte.Resource{},
		assignments: map[string]map[string]bool{},
	}
	s.server, s.body, s.oldBody = nil, nil, nil
	s.status, s.oldStatus = 0, 0
	return ctx, nil
}

func (s *Suite) after(ctx context.Context, _ *godog.Scenario, _ error) (context.Context, error) {
	if s.server != nil {
		s.server.Close()
	}

	return ctx, nil
}

func (d *Directory) Actors() []domainrechte.Actor {
	return d.actors
}

func (d *Directory) OrganizationIDs(id string) []string {
	return d.organizations[id]
}

func (d *Directory) Resource(id string) (domainrechte.Resource, bool) {
	resource, found := d.resources[id]
	return resource, found
}

func (d *Directory) Assigned(agentID, resourceID string) bool {
	return d.assignments[agentID][resourceID]
}

func (s *Suite) resource(id, name, organization string) error {
	s.directory.resources[id] = domainrechte.NewResource(id, name, organization)
	return nil
}

func (s *Suite) operator(id, organization string) error {
	s.directory.actors = []domainrechte.Actor{domainrechte.NewActor(id, domainrechte.Operator)}
	s.directory.organizations[id] = []string{organization}
	return nil
}

func (s *Suite) assignedAgent(id, organization, resourceID string) error {
	s.directory.actors = []domainrechte.Actor{domainrechte.NewActor(id, domainrechte.Agent)}
	s.directory.organizations[id] = []string{organization}
	s.directory.assignments[id] = map[string]bool{resourceID: true}
	return nil
}

func (s *Suite) unassignedAgent(id, organization string) error {
	s.directory.actors = []domainrechte.Actor{domainrechte.NewActor(id, domainrechte.Agent)}
	s.directory.organizations[id] = []string{organization}
	return nil
}

func (s *Suite) context(value string) error {
	if value == "keine Akteuridentität" {
		return nil
	}

	if value == "zwei Akteuridentitäten" {
		return s.twoActors()
	}

	if value == "einen eindeutigen Betreiber ohne Organisationszuordnung" {
		s.directory.actors = []domainrechte.Actor{domainrechte.NewActor("betreiber-a", domainrechte.Operator)}
		return nil
	}

	if value == "einen eindeutigen Betreiber mit zwei Organisationszuordnungen" {
		return s.twoOrganizations()
	}

	return fmt.Errorf("unbekannter Kontext: %q", value)
}

func (s *Suite) twoActors() error {
	s.directory.actors = []domainrechte.Actor{
		domainrechte.NewActor("betreiber-a", domainrechte.Operator),
		domainrechte.NewActor("betreiber-b", domainrechte.Operator),
	}
	return nil
}

func (s *Suite) twoOrganizations() error {
	s.operator("betreiber-a", "org-eigen")
	s.directory.organizations["betreiber-a"] = []string{"org-eigen", "org-fremd"}
	return nil
}

func (s *Suite) get(path string) error {
	return s.send(path, nil)
}

func (s *Suite) getWithHeaders(path string, table *godog.Table) error {
	headers := http.Header{}
	for _, row := range table.Rows {
		if len(row.Cells) != 2 {
			return fmt.Errorf("Header-Zeile benötigt Name und Wert")
		}

		headers.Set(row.Cells[0].Value, row.Cells[1].Value)
	}

	return s.send(path, headers)
}

func (s *Suite) send(path string, headers http.Header) error {
	s.ensureServer()
	request, err := http.NewRequest(http.MethodGet, s.server.URL+path, nil)
	if err != nil {
		return err
	}

	request.Header = headers
	response, err := s.server.Client().Do(request)
	if err != nil {
		return err
	}

	defer response.Body.Close()
	s.oldStatus, s.oldBody = s.status, bytes.Clone(s.body)
	s.status = response.StatusCode
	s.body, err = io.ReadAll(response.Body)
	return err
}

func (s *Suite) ensureServer() {
	if s.server == nil {
		s.server = httptest.NewServer(webrechte.NewHandler(s.directory))
	}
}

func (s *Suite) expectStatus(expected int) error {
	if s.status != expected {
		return fmt.Errorf("HTTP-Status: erhalten %d, erwartet %d", s.status, expected)
	}

	return nil
}

func (s *Suite) expectResource(id, name, organization string) error {
	var actual map[string]any
	if err := json.Unmarshal(s.body, &actual); err != nil {
		return err
	}

	expected := map[string]any{"id": id, "name": name, "organization_id": organization}
	if !reflect.DeepEqual(actual, expected) {
		return fmt.Errorf("Ressource: erhalten %s, erwartet %v", s.body, expected)
	}

	return nil
}

func (s *Suite) expectError(code string) error {
	expected, err := json.Marshal(map[string]string{"error": code})
	if err != nil {
		return err
	}

	if !bytes.Equal(bytes.TrimSpace(s.body), expected) {
		return fmt.Errorf("Fehlerkörper: erhalten %q, erwartet %q", s.body, expected)
	}

	return nil
}

func (s *Suite) expectSame() error {
	if s.status != s.oldStatus || !bytes.Equal(s.body, s.oldBody) {
		return fmt.Errorf("Antworten unterscheiden sich: %d %q gegen %d %q", s.oldStatus, s.oldBody, s.status, s.body)
	}

	return nil
}
