package projekte

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/cucumber/godog"
)

func (s *Suite) registerRegressionSteps(sc *godog.ScenarioContext) {
	sc.Step(`^ich die öffentlichen Organisations-, Ziel-, Agenten-, Projekt- und Settings-Seiten von "([^"]*)" über HTTP abrufe$`, s.canonicalPages)
	sc.Step(`^antwortet jede dieser Seiten mit HTTP 200$`, s.allCanonicalPagesOK)
	sc.Step(`^die Organisationsübersicht verlinkt nach "/settings"$`, s.listLinksToSettings)
	sc.Step(`^die Organisationsdetails verlinken nach "/settings"$`, s.detailLinksToSettings)
}

func (s *Suite) canonicalPages(organization string) error {
	id := s.organizations[organization]
	s.routePages = map[string]HTTPResponse{}
	paths := map[string]string{
		"list": "/organisationen", "detail": "/organisationen/" + id,
		"goals": "/organisationen/" + id + "/ziele", "agents": "/organisationen/" + id + "/agenten",
		"projects": "/organisationen/" + id + "/projekte", "settings": "/settings",
	}
	for name, path := range paths {
		if err := s.request("GET", path, "", nil, ""); err != nil {
			return err
		}

		s.routePages[name] = *s.response
	}

	return nil
}

func (s *Suite) allCanonicalPagesOK() error {
	if len(s.routePages) != 6 {
		return fmt.Errorf("kanonische Seiten fehlen: %+v", s.routePages)
	}

	for name, response := range s.routePages {
		if response.Status != http.StatusOK || !strings.Contains(response.Header.Get("Content-Type"), "text/html") {
			return fmt.Errorf("%s nicht als HTML verfügbar: HTTP %d: %s", name, response.Status, response.Body)
		}
	}

	settings := string(s.routePages["settings"].Body)
	if !strings.Contains(settings, "/settings/modelle/codex") || !strings.Contains(settings, "/settings/modellanbieter/openrouter") {
		return fmt.Errorf("globale Settings ohne Anbieterlinks: %s", settings)
	}

	return nil
}

func (s *Suite) listLinksToSettings() error   { return s.pageLinksToSettings("list") }
func (s *Suite) detailLinksToSettings() error { return s.pageLinksToSettings("detail") }

func (s *Suite) pageLinksToSettings(name string) error {
	if !strings.Contains(string(s.routePages[name].Body), `href="/settings"`) {
		return fmt.Errorf("%s ohne Settings-Link: %s", name, s.routePages[name].Body)
	}

	return nil
}
