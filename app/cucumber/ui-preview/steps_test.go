package uipreview

import (
	"fmt"
	"strings"

	"github.com/cucumber/godog"
)

func (s *Suite) initializeScenario(sc *godog.ScenarioContext) {
	sc.Step(`^die UI wurde frisch gebaut und die Go-Vorschau wird mit einem freien Vorschauport gestartet$`, s.ready)
	s.registerNetwork(sc)
	s.registerPages(sc)
	s.registerFixtures(sc)
	s.registerAssets(sc)
	s.registerDenials(sc)
}

func (s *Suite) ready() error {
	return nil
}

func (s *Suite) registerNetwork(sc *godog.ScenarioContext) {
	sc.Step(`^lauscht die Vorschau ausschließlich auf einer Loopback-Adresse$`, s.loopbackOnly)
	sc.Step(`^der konfigurierte Vorschauport ist über diese Adresse erreichbar$`, s.reachable)
	sc.Step(`^eine Verbindung über eine andere lokale Netzwerkschnittstelle wird nicht angenommen$`, s.nonLoopbackDenied)
	sc.Step(`^die Vorschau mit einem ungültigen oder bereits belegten Vorschauport gestartet wird$`, s.invalidPort)
	sc.Step(`^endet der Start mit einer verständlichen Fehlermeldung$`, s.startError)
	sc.Step(`^auf diesem Port wird kein weiterer Vorschauprozess gestartet$`, s.noSecondServer)
}

func (s *Suite) registerPages(sc *godog.ScenarioContext) {
	sc.Step(`^ich im Browser "([^"]*)" der Go-Vorschau öffne$`, s.browse)
	sc.Step(`^ich im Browser "([^"]*)" öffne$`, s.browse)
	sc.Step(`^ich "([^"]*)" öffne$`, s.browse)
	sc.Step(`^ich die Organisationsseite der Go-Vorschau öffne$`, s.browseOrganization)
	sc.Step(`^sehe ich ein vollständiges HTML-Dokument mit der Beispielorganisation "([^"]*)"$`, s.fullOrganization)
	sc.Step(`^ich kann über die Navigation Projekte, Agenten und Aufgaben öffnen$`, s.navigation)
	sc.Step(`^ich "([^"]*)" als normale Browseranfrage und als HTMX-Anfrage öffne$`, s.fullAndFragment)
	sc.Step(`^zeigen beide Antworten dieselben Projekte aus "([^"]*)"$`, s.sameProjects)
	sc.Step(`^nur die normale Antwort enthält Dokumentrahmen und Navigation$`, s.onlyFullShell)
	sc.Step(`^die HTMX-Antwort enthält nur den austauschbaren Inhaltsbereich$`, s.fragmentOnly)
}

func (s *Suite) registerFixtures(sc *godog.ScenarioContext) {
	sc.Step(`^sehe ich "([^"]*)" und nicht "([^"]*)"$`, s.alternateName)
	sc.Step(`^ich von dort über die Navigation die Projektansicht öffne$`, s.clickProjects)
	sc.Step(`^zeigt die Adresse weiterhin die Fixture-Auswahl "([^"]*)"$`, s.fixtureAddress)
	sc.Step(`^die Projektansicht verwendet die Daten aus "([^"]*)"$`, s.alternateProjects)
	sc.Step(`^eine Kopie der Organisations-Fixture enthält einen neuen eindeutigen Namen$`, s.copyFixture)
	sc.Step(`^ich die UI aus dieser Kopie ohne Go-Codeänderung neu baue und die Go-Vorschau starte$`, s.rebuildFixture)
	sc.Step(`^zeigt die Organisationsansicht den neuen Namen aus der JSON-Fixture$`, s.changedFixture)
	sc.Step(`^sehe ich den verständlichen Leerzustand der Projektansicht$`, s.emptyProjects)
	sc.Step(`^sehe ich den verständlichen Aufgabenfehler mit einem Rückweg zur Übersicht$`, s.tasksError)
}

func (s *Suite) registerAssets(sc *godog.ScenarioContext) {
	sc.Step(`^laden die referenzierten CSS- und HTMX-Dateien erfolgreich$`, s.assets)
	sc.Step(`^ich die Vorschauhilfe im Browser öffne$`, s.openHelp)
	sc.Step(`^erscheint das eingebettete Hilfe-Fragment ohne vollständigen Seitenrahmen$`, s.helpFragment)
}

func (s *Suite) registerDenials(sc *godog.ScenarioContext) {
	sc.Step(`^ich eine unbekannte Ansicht oder Fixture über die Vorschauadresse anfordere$`, s.unknown)
	sc.Step(`^erhalte ich eine Nicht-gefunden-Antwort ohne Beispieldaten einer anderen Ansicht$`, s.unknownDenied)
	sc.Step(`^ich einen Pfad außerhalb der erlaubten Ansichts-, Asset- und Fragmentrouten anfordere$`, s.otherPath)
	sc.Step(`^erhalte ich eine Nicht-gefunden-Antwort ohne lokale Dateiinhalte$`, s.otherPathDenied)
	sc.Step(`^ich einen Pfadversuch im Fixture-Namen oder Asset-Pfad anfordere$`, s.traversal)
	sc.Step(`^erhalte ich keine lokalen Dateiinhalte und keine Fixture-JSON-Antwort$`, s.traversalDenied)
}

func (s *Suite) browseOrganization() error {
	return s.browse("/?view=organization")
}

func (s *Suite) fullOrganization(name string) error {
	if s.page.Title == "" || s.page.Heading == "" {
		return fmt.Errorf("browser did not render a full document: %+v", s.page)
	}

	return s.contains(name)
}

func (s *Suite) navigation() error {
	for _, item := range []string{"projects", "agents", "tasks"} {
		page, err := s.browserCommand(map[string]any{"click": `nav a[href*="view=` + item + `"]`})
		if err != nil || !strings.Contains(page.URL, "view="+item) {
			return fmt.Errorf("navigation to %s: %v, %+v", item, err, page)
		}
	}

	return nil
}

func (s *Suite) fullAndFragment(path string) error {
	full, err := s.fetch(path, false)
	if err != nil {
		return err
	}

	part, err := s.fetch(path, true)
	s.full, s.part = full.Body, part.Body
	return err
}

func (s *Suite) sameProjects(name string) error {
	if name != "projects.json" || !strings.Contains(s.full, "Kundenportal") || !strings.Contains(s.part, "Kundenportal") {
		return fmt.Errorf("project data differs across full and HTMX responses")
	}

	return nil
}

func (s *Suite) onlyFullShell() error {
	if !strings.Contains(s.full, "<!doctype html>") || !strings.Contains(s.full, "<nav>") {
		return fmt.Errorf("full page misses document shell")
	}

	if strings.Contains(s.part, "<!doctype html>") || strings.Contains(s.part, "<nav>") {
		return fmt.Errorf("fragment includes document shell")
	}

	return nil
}

func (s *Suite) fragmentOnly() error {
	if !strings.Contains(s.part, "project-grid") || strings.Contains(s.part, "<html") {
		return fmt.Errorf("HTMX response is not a project fragment")
	}

	return nil
}
