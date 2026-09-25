package projektort

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/cucumber/godog"
)

func (s *suite) registerBrowserSteps(sc *godog.ScenarioContext) {
	sc.Step(`^ich öffne die Anwendung mit dem Projekt "([^"]*)" in "([^"]*)" im Browser$`, s.browserFixture)
	sc.Step(`^ich öffne die Anwendung mit Projekten in "([^"]*)" und "([^"]*)" im Browser$`, s.browserTwoOrganizations)
	sc.Step(`^ich im Projektdetail den Ausführungsort öffne$`, s.browserOpenLocation)
	sc.Step(`^sehe ich den privaten appverwalteten Projektarbeitsordner ohne frei editierbare Pfadeingabe$`, s.browserPrivateLocation)
	sc.Step(`^der Ausführungsort ist zunächst deaktiviert$`, s.browserInitiallyOff)
	sc.Step(`^ich den privaten Ausführungsort aktiviere und speichere$`, s.browserEnable)
	sc.Step(`^sehe ich den aktivierten privaten Ort im Projektdetail$`, s.browserActive)
	sc.Step(`^ich die Anwendung mit derselben SQLite-Datenbank neu starte und die Seite neu lade$`, s.browserRestart)
	sc.Step(`^sehe ich den aktivierten privaten Ort erneut$`, s.browserActive)
	sc.Step(`^sehe ich den fehlenden Ausführungsort und die Korrekturaktion zum privaten Projektarbeitsordner$`, s.browserMissingCorrection)
	sc.Step(`^ich über "([^"]*)" den Ausführungsort eines Projekts aus "([^"]*)" öffne$`, s.browserForeignLocation)
	sc.Step(`^sehe ich keine Projektdaten und keinen privaten Pfad$`, s.browserOpaque)
}

func (s *suite) browserFixture(project, org string) error {
	if err := s.fresh(); err != nil {
		return err
	}
	if err := s.createFixture(org, project); err != nil {
		return err
	}
	if err := s.startBrowser(); err != nil {
		return err
	}
	return s.browserCommand(map[string]any{"url": s.url(strings.TrimPrefix(strings.TrimSuffix(s.pagePath(org, project), "/ausfuehrungsort"), "/api")), "mode": "detail"})
}

func (s *suite) browserTwoOrganizations(first, second string) error {
	if err := s.fresh(); err != nil {
		return err
	}
	if err := s.createFixture(first, "Website"); err != nil {
		return err
	}
	if err := s.createFixture(second, "Intern"); err != nil {
		return err
	}
	if err := s.startBrowser(); err != nil {
		return err
	}
	return s.browserCommand(map[string]any{"url": s.url("/organisationen/" + s.organizations[first] + "/projekte/" + s.projects[s.key(first, "Website")]), "mode": "detail"})
}

func (s *suite) browserOpenLocation() error {
	// The project detail must provide this link; navigating directly would miss that user path.
	return s.browserCommand(map[string]any{"click": "main a[href$='/ausfuehrungsort']", "mode": "location"})
}

func (s *suite) browserPrivateLocation() error {
	if !strings.Contains(s.page.URL, "/ausfuehrungsort") || s.page.PathInput {
		return fmt.Errorf("freie Pfadeingabe oder falsche Seite: %+v", s.page)
	}
	if !strings.Contains(strings.ToLower(s.page.Text), "privat") || !strings.Contains(strings.ToLower(s.page.Text), "projekt") {
		return fmt.Errorf("privater Projektarbeitsordner nicht erklärt: %+v", s.page)
	}
	return nil
}

func (s *suite) browserInitiallyOff() error {
	if s.page.Enabled || !strings.Contains(strings.ToLower(s.page.Text), "nicht") {
		return fmt.Errorf("ungebundener Zustand fehlt: %+v", s.page)
	}
	return nil
}

func (s *suite) browserEnable() error {
	if err := s.browserCommand(map[string]any{"submit": true, "mode": "location"}); err != nil {
		return err
	}
	return s.activeLocation(s.lastProject, s.lastOrg)
}

func (s *suite) browserActive() error {
	if err := s.activeLocation(s.lastProject, s.lastOrg); err != nil {
		return err
	}
	if !strings.Contains(strings.ToLower(s.page.Text), "aktiv") && !strings.Contains(strings.ToLower(s.page.Text), "gebunden") {
		return fmt.Errorf("aktiver Ort im Browser nicht sichtbar: %+v", s.page)
	}
	return nil
}

func (s *suite) browserRestart() error {
	if err := s.restart(); err != nil {
		return err
	}
	return s.browserCommand(map[string]any{"url": s.url(s.pagePath(s.lastOrg, s.lastProject)), "mode": "location"})
}

func (s *suite) browserMissingCorrection() error {
	if !strings.Contains(strings.ToLower(s.page.Text), "nicht") || !strings.Contains(strings.ToLower(s.page.Text), "aktiv") {
		return fmt.Errorf("Korrekturaktion fehlt: %+v", s.page)
	}
	return nil
}

func (s *suite) browserForeignLocation(first, second string) error {
	path := "/organisationen/" + s.organizations[first] + "/projekte/" + s.projects[s.key(second, "Intern")] + "/ausfuehrungsort"
	if err := s.call(http.MethodGet, "/api"+path, nil); err != nil {
		return err
	}
	if err := s.status(http.StatusNotFound); err != nil {
		return err
	}
	return s.browserCommand(map[string]any{"url": s.url(path), "mode": "error"})
}

func (s *suite) browserOpaque() error {
	if strings.Contains(s.page.Text, "Intern") || strings.Contains(s.page.Text, "project-workspaces") || strings.Contains(s.page.Text, s.dbPath) {
		return fmt.Errorf("fremder Projektort sichtbar: %+v", s.page)
	}
	return nil
}
