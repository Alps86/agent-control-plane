package uibridge

import (
	"bytes"
	"fmt"
	"net/http"
	"os/exec"
	"regexp"
	"strings"

	"github.com/cucumber/godog"
)

func (s *Suite) initializeScenario(sc *godog.ScenarioContext) {
	sc.Step(`^die UI ist aus den Vite-Quellen frisch gebaut und in die Bridge eingebettet$`, s.builtUI)
	s.registerBuildSteps(sc)
	s.registerPageSteps(sc)
}

func (s *Suite) registerBuildSteps(sc *godog.ScenarioContext) {
	sc.Step(`^der frische Build enthält die neutralen Templates "bridge-check/page.html" und "bridge-check/content.html"$`, s.neutralTemplatesBuilt)
	sc.Step(`^der Anwendungsserver verwendet für alle Templates dieselbe Ansichts-Map$`, s.sameViewMap)
	sc.Step(`^er "bridge-check/page" und "bridge-check/content" über die öffentliche Bridge rendert$`, s.renderNeutral)
	sc.Step(`^enthalten beide Antworten die Werte dieser Ansichts-Map$`, s.neutralValues)
	sc.Step(`^die benannten Root-Templates "page" und "content" bleiben mit derselben Map renderbar$`, s.rootTemplatesStillWork)
}

func (s *Suite) registerPageSteps(sc *godog.ScenarioContext) {
	sc.Step(`^der Anwendungsserver die Organisationsansicht mit der Map der JSON-Fixture "organization.json" als ganze Seite rendert$`, s.renderOrganization)
	sc.Step(`^enthält die Antwort ein vollständiges HTML-Dokument mit dem Namen der Beispielorganisation$`, s.completeOrganization)
	sc.Step(`^die referenzierten CSS- und HTMX-Assets sind über die öffentliche Bridge erreichbar$`, s.assetsReachable)
	sc.Step(`^die ausgelieferten Assets stammen aus dem eingebetteten Build$`, s.embeddedAssets)
	sc.Step(`^der Anwendungsserver verwendet die Map der JSON-Fixture "projects.json"$`, s.useProjects)
	sc.Step(`^er daraus die Projektansicht als ganze Seite und als HTMX-Fragment rendert$`, s.renderProjects)
	sc.Step(`^zeigen beide Antworten denselben Projekttitel und dieselben Projekte$`, s.sameProjects)
	sc.Step(`^nur die ganze Seite enthält Dokumentrahmen und Navigation$`, s.onlyPageHasFrame)
	sc.Step(`^das Fragment enthält ausschließlich den austauschbaren Inhaltsbereich$`, s.onlyContent)
	s.registerFixtureSteps(sc)
}

func (s *Suite) registerFixtureSteps(sc *godog.ScenarioContext) {
	sc.Step(`^die Vorschau über die öffentliche Bridge "organization-alternate.json" lädt und damit die Organisationsansicht rendert$`, s.alternateOrganization)
	sc.Step(`^sehe ich den Organisationsnamen der alternativen JSON-Fixture$`, s.seeAlternate)
	sc.Step(`^ich sehe dort nicht den Organisationsnamen der zuerst ausgelieferten Fixture$`, s.notOriginal)
	sc.Step(`^die Ansichts-Map enthält als Organisationsnamen "<script>alert\(1\)</script>"$`, s.scriptName)
	sc.Step(`^der Anwendungsserver die Organisationsansicht als ganze Seite rendert$`, s.renderMap)
	sc.Step(`^steht der Organisationsname nur als maskierter Text im HTML$`, s.escapedText)
	sc.Step(`^kein ausführbares Script aus dem Organisationsnamen erscheint im Dokument$`, s.noScript)
	s.registerAccessSteps(sc)
}

func (s *Suite) registerAccessSteps(sc *godog.ScenarioContext) {
	sc.Step(`^eine unbekannte Fixture oder ein Pfad außerhalb des Fixture-Verzeichnisses angefordert wird$`, s.invalidFixtures)
	sc.Step(`^verweigert die öffentliche Bridge den Zugriff ohne Dateiinhalte preiszugeben$`, s.fixtureDenied)
	sc.Step(`^ein unbekanntes Asset angefordert wird$`, s.unknownAsset)
	sc.Step(`^liefert die öffentliche Bridge eine Nicht-gefunden-Antwort$`, s.assetNotFound)
}

func (s *Suite) builtUI() error {
	if !s.built {
		if err := s.buildFresh(); err != nil {
			return err
		}

		s.built = true
	}

	s.data = nil
	s.page = ""
	s.part = ""
	s.asset = ""
	s.denied = nil
	s.status = 0
	if err := s.verifyCopyTree("templates"); err != nil {
		return err
	}

	if err := s.verifyCopyTree("fixtures"); err != nil {
		return err
	}

	return s.verifyEmbeddedAssets()
}

func (s *Suite) buildFresh() error {
	build := exec.Command("npm", "run", "build")
	build.Dir = "../../../ui/web"
	output, err := build.CombinedOutput()
	if err != nil {
		return fmt.Errorf("fresh Vite build: %w: %s", err, output)
	}

	return nil
}

func (s *Suite) renderOrganization() error {
	s.fixture = "organization.json"
	return s.loadAndRender()
}

func (s *Suite) loadAndRender() error {
	var err error
	s.data, err = s.bridge.LoadFixture(s.fixture)
	if err != nil {
		return err
	}

	return s.renderMap()
}

func (s *Suite) renderMap() error {
	var output bytes.Buffer
	if err := s.bridge.Render(&output, "page", s.data); err != nil {
		return err
	}

	s.page = output.String()
	return nil
}

func (s *Suite) completeOrganization() error {
	if !strings.Contains(s.page, "<!doctype html>") || !strings.Contains(s.page, "Atelier Nord") {
		return fmt.Errorf("organization page incomplete")
	}

	page, err := s.browserPage(s.server.URL + "/?view=organization")
	if err != nil {
		return err
	}

	if !strings.Contains(page.Heading, "Atelier Nord") || !strings.Contains(page.Text, "Auf einen Blick") {
		return fmt.Errorf("Chrome did not display organization: %+v", page)
	}

	help, err := s.browserCommand(map[string]any{"help": true})
	if err != nil || !strings.Contains(help.Help, "Vorschau") {
		return fmt.Errorf("Chrome HTMX help missing: %+v: %v", help, err)
	}

	return nil
}

func (s *Suite) assetsReachable() error {
	paths := regexp.MustCompile(`(?:href|src)="(/assets/[^"]+)"`).FindAllStringSubmatch(s.page, -1)
	if len(paths) < 2 {
		return fmt.Errorf("CSS and HTMX references missing")
	}

	for _, path := range paths {
		body, status, err := s.fetch(path[1])
		if err != nil || status != http.StatusOK || len(body) < 50 {
			return fmt.Errorf("asset %s: status %d, bytes %d: %v", path[1], status, len(body), err)
		}
	}

	return nil
}

func (s *Suite) embeddedAssets() error {
	css, status, err := s.fetch("/assets/index.css")
	if err != nil || status != http.StatusOK || !strings.Contains(css, "app-shell") {
		return fmt.Errorf("embedded CSS missing: status %d: %v", status, err)
	}

	htmx, status, err := s.fetch("/assets/htmx.min.js")
	if err != nil || status != http.StatusOK || !strings.Contains(htmx, "htmx") {
		return fmt.Errorf("embedded HTMX missing: status %d: %v", status, err)
	}

	return nil
}

func (s *Suite) useProjects() error {
	s.fixture = "projects.json"
	return s.loadAndRender()
}

func (s *Suite) renderProjects() error {
	page, status, err := s.fetch("/?view=projects")
	if err != nil || status != http.StatusOK {
		return fmt.Errorf("project page HTTP %d: %v", status, err)
	}

	part, status, err := s.fetch("/fragment?view=projects")
	if err != nil || status != http.StatusOK {
		return fmt.Errorf("project fragment HTTP %d: %v", status, err)
	}

	s.page, s.part = page, part
	return nil
}

func (s *Suite) sameProjects() error {
	for _, expected := range []string{"Projekte", "Kundenportal", "Wissensbasis", "Markenauftritt"} {
		if !strings.Contains(s.page, expected) || !strings.Contains(s.part, expected) {
			return fmt.Errorf("project %q absent from page or fragment", expected)
		}
	}

	return nil
}

func (s *Suite) onlyPageHasFrame() error {
	if !strings.Contains(s.page, "<!doctype html>") || !strings.Contains(s.page, "<nav>") {
		return fmt.Errorf("full page lacks document frame")
	}

	if strings.Contains(s.part, "<!doctype html>") || strings.Contains(s.part, "<nav>") {
		return fmt.Errorf("fragment contains document frame")
	}

	return nil
}

func (s *Suite) onlyContent() error {
	if !strings.Contains(s.part, "project-grid") || strings.Contains(s.part, "app-shell") {
		return fmt.Errorf("fragment is not isolated content")
	}

	return nil
}

func (s *Suite) alternateOrganization() error {
	s.fixture = "organization-alternate.json"
	return s.loadAndRender()
}

func (s *Suite) seeAlternate() error {
	if !strings.Contains(s.page, "Küstenwerk") {
		return fmt.Errorf("alternate fixture organization absent")
	}

	return nil
}

func (s *Suite) notOriginal() error {
	if strings.Contains(s.page, "Atelier Nord") {
		return fmt.Errorf("original fixture organization remains")
	}

	return nil
}

func (s *Suite) scriptName() error {
	s.fixture = "organization.json"
	data, err := s.bridge.LoadFixture(s.fixture)
	if err != nil {
		return err
	}

	s.data = data
	view := data["View"].(map[string]any)
	organization := view["Organization"].(map[string]any)
	organization["Name"] = "<script>alert(1)</script>"
	return nil
}

func (s *Suite) escapedText() error {
	if !strings.Contains(s.page, "&lt;script&gt;alert(1)&lt;/script&gt;") {
		return fmt.Errorf("script name missing as escaped text")
	}

	return nil
}

func (s *Suite) noScript() error {
	if strings.Contains(s.page, "<script>alert(1)</script>") {
		return fmt.Errorf("script name is executable")
	}

	return nil
}

func (s *Suite) invalidFixtures() error {
	for _, name := range []string{"unknown.json", "../go.mod", "../../app/go.mod"} {
		data, err := s.bridge.LoadFixture(name)
		if err == nil || data != nil {
			return fmt.Errorf("fixture %q was accepted", name)
		}

		s.denied = append(s.denied, err.Error())
	}

	return nil
}

func (s *Suite) fixtureDenied() error {
	if len(s.denied) != 3 {
		return fmt.Errorf("fixture denial count %d", len(s.denied))
	}

	for _, failure := range s.denied {
		if strings.Contains(failure, "module agentcontrolplane") || strings.Contains(failure, "require (") {
			return fmt.Errorf("fixture rejection disclosed file contents")
		}
	}

	return nil
}

func (s *Suite) unknownAsset() error {
	_, status, err := s.fetch("/assets/unknown.css")
	if err != nil {
		return err
	}

	s.status = status
	return nil
}

func (s *Suite) assetNotFound() error {
	if s.status != http.StatusNotFound {
		return fmt.Errorf("unknown asset status %d", s.status)
	}

	return nil
}
