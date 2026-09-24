package uibootstrap

import (
	"fmt"
	"strings"

	"github.com/cucumber/godog"
)

func (s *Suite) initializeScenario(sc *godog.ScenarioContext) {
	sc.Step(`^die statische Vorschau ist mit den ausgelieferten JSON-Beispieldaten geöffnet$`, s.defaultPreview)
	sc.Step(`^ich die Organisationsübersicht öffne$`, s.openOrganization)
	sc.Step(`^sehe ich den Namen der Beispielorganisation und ihre Übersicht$`, s.seeOrganization)
	sc.Step(`^ich kann über die Navigation die Projekte, Agenten und Aufgaben öffnen$`, s.seeNavigation)
	sc.Step(`^ich nacheinander die Projekt-, Agenten- und Aufgabenansicht öffne$`, s.openAllViews)
	sc.Step(`^zeigt jede Ansicht ihren eigenen Titel und die zugehörigen Beispieldaten$`, s.seeViews)
	sc.Step(`^die Navigation zeigt mir die aktuell geöffnete Ansicht an$`, s.seeActive)
	s.registerStates(sc)
}

func (s *Suite) registerStates(sc *godog.ScenarioContext) {
	sc.Step(`^ich wähle die Beispieldaten für eine Organisation ohne Projekte, Agenten und Aufgaben$`, s.chooseEmpty)
	sc.Step(`^zeigt jede Ansicht einen passenden Leerzustand statt einer leeren Fläche$`, s.seeEmpty)
	sc.Step(`^die Navigation zur Organisation bleibt benutzbar$`, s.returnOrganization)
	sc.Step(`^ich wähle die Beispieldaten für eine fehlerhafte Aufgabenansicht$`, s.chooseError)
	sc.Step(`^ich die Aufgabenansicht öffne$`, s.openTasks)
	sc.Step(`^sehe ich einen verständlichen Fehlerhinweis ohne rohe technische Fehlermeldung$`, s.seeError)
	sc.Step(`^ich kann über die Navigation zur Organisationsübersicht zurückkehren$`, s.returnOrganization)
	sc.Step(`^ich wähle eine zweite JSON-Fixture mit einem anderen Organisationsnamen$`, s.chooseAlternate)
	sc.Step(`^sehe ich den Organisationsnamen dieser zweiten Fixture$`, s.seeAlternate)
	sc.Step(`^ich sehe dort nicht den Organisationsnamen der zuerst ausgelieferten Fixture$`, s.notOriginal)
	sc.Step(`^ich betrachte die statische Vorschau auf einem schmalen Bildschirm$`, s.chooseNarrow)
	sc.Step(`^kann ich die Navigation zu Projekten, Agenten und Aufgaben erreichen$`, s.seeNavigation)
	sc.Step(`^die Inhalte sind ohne horizontales Scrollen lesbar$`, s.noHorizontalScroll)
}

func (s *Suite) defaultPreview() error {
	s.fixture = ""
	s.view = ""
	s.narrow = false
	return s.ask(map[string]any{"url": s.baseURL + "/?view=organization"})
}

func (s *Suite) openOrganization() error {
	s.view = "organization"
	return s.open("organization")
}

func (s *Suite) openTasks() error {
	s.view = "tasks"
	return s.open("tasks")
}

func (s *Suite) open(view string) error {
	url := s.baseURL + "/?view=" + view
	if s.fixture != "" {
		url += "&fixture=" + s.fixture
	}

	return s.ask(map[string]any{"url": url})
}

func (s *Suite) click(view string) error {
	selector := fmt.Sprintf(`nav a[href^="/?view=%s"]`, view)
	s.view = view
	return s.ask(map[string]any{"click": selector})
}

func (s *Suite) chooseEmpty() error {
	s.fixture = "empty"
	return nil
}

func (s *Suite) chooseError() error {
	s.fixture = "error"
	return nil
}

func (s *Suite) chooseAlternate() error {
	s.fixture = "alternate"
	return nil
}

func (s *Suite) chooseNarrow() error {
	s.narrow = true
	return s.ask(map[string]any{"width": 375, "url": s.baseURL + "/?view=organization"})
}

func (s *Suite) seeOrganization() error {
	return s.contains("Atelier Nord", "Auf einen Blick")
}

func (s *Suite) seeNavigation() error {
	for _, view := range []string{"projects", "agents", "tasks"} {
		if err := s.scrollLink(view); err != nil {
			return err
		}

		if !s.hasVisibleLink(view) {
			return fmt.Errorf("navigation link %s missing: %+v", view, s.page.Links)
		}
	}

	if s.narrow {
		return nil
	}

	return s.checkHelp()
}

func (s *Suite) scrollLink(view string) error {
	selector := fmt.Sprintf(`nav a[href^="/?view=%s"]`, view)
	return s.ask(map[string]any{"scroll": selector})
}

func (s *Suite) hasVisibleLink(view string) bool {
	for _, link := range s.page.Links {
		if strings.Contains(link.Href, "view="+view) && link.Visible {
			return true
		}
	}

	return false
}

func (s *Suite) checkHelp() error {
	if err := s.ask(map[string]any{"help": true}); err != nil {
		return err
	}

	if s.page.Help == "" {
		return fmt.Errorf("HTMX preview help did not load")
	}

	return nil
}

func (s *Suite) openAllViews() error {
	if err := s.open("organization"); err != nil {
		return err
	}

	return s.visitAllViews()
}

func (s *Suite) visitAllViews() error {
	for _, view := range []string{"projects", "agents", "tasks"} {
		if err := s.click(view); err != nil {
			return err
		}

		if err := s.checkView(view); err != nil {
			return err
		}

		if err := s.checkActiveView(view); err != nil {
			return err
		}
	}

	return nil
}

func (s *Suite) checkActiveView(view string) error {
	labels := map[string]string{"projects": "Projekte", "agents": "Agenten", "tasks": "Aufgaben"}
	if !strings.Contains(s.page.Active, labels[view]) {
		return fmt.Errorf("active navigation for %s: %q", view, s.page.Active)
	}

	return nil
}

func (s *Suite) checkView(view string) error {
	titles := map[string]string{"projects": "Projekte", "agents": "Agenten", "tasks": "Aufgaben"}
	if s.page.Title != titles[view]+" · Agent Control Plane" {
		return fmt.Errorf("document title for %s: %q", view, s.page.Title)
	}

	if s.fixture == "empty" {
		return s.checkEmptyView(view)
	}

	examples := map[string]string{"projects": "Kundenportal", "agents": "Mira", "tasks": "Onboarding-Ablauf vereinfachen"}
	if s.page.Heading != titles[view] {
		return fmt.Errorf("heading for %s: %q", view, s.page.Heading)
	}

	return s.contains(examples[view])
}

func (s *Suite) checkEmptyView(view string) error {
	expected := map[string]string{"projects": "Noch keine Projekte", "agents": "Noch keine Agenten", "tasks": "Noch keine Aufgaben"}
	return s.contains(expected[view], "Zur Übersicht")
}

func (s *Suite) seeViews() error {
	return s.checkView(s.view)
}

func (s *Suite) seeActive() error {
	if !strings.Contains(s.page.Active, "Aufgaben") {
		return fmt.Errorf("active navigation is %q", s.page.Active)
	}

	return nil
}

func (s *Suite) seeEmpty() error {
	return s.checkEmptyView(s.view)
}

func (s *Suite) returnOrganization() error {
	if err := s.click("organization"); err != nil {
		return err
	}

	if !strings.Contains(s.page.Heading, "Guten Tag") {
		return fmt.Errorf("organization page missing: %q", s.page.Heading)
	}

	return nil
}

func (s *Suite) seeError() error {
	if err := s.contains("Aufgaben nicht verfügbar", "Bitte versuchen Sie es erneut"); err != nil {
		return err
	}

	if strings.Contains(s.page.Alert, "Error:") || strings.Contains(s.page.Alert, "stack") {
		return fmt.Errorf("raw technical error shown: %q", s.page.Alert)
	}

	return nil
}

func (s *Suite) seeAlternate() error {
	return s.contains("Küstenwerk")
}

func (s *Suite) notOriginal() error {
	if strings.Contains(s.page.Text, "Atelier Nord") {
		return fmt.Errorf("original organization still visible")
	}

	return nil
}

func (s *Suite) noHorizontalScroll() error {
	if s.page.Horizontal {
		return fmt.Errorf("page overflows viewport horizontally")
	}

	return nil
}

func (s *Suite) contains(values ...string) error {
	for _, value := range values {
		if !strings.Contains(s.page.Text, value) {
			return fmt.Errorf("visible content lacks %q: %s", value, s.page.Text)
		}
	}

	return nil
}
