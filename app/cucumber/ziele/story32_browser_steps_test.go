package ziele

import (
	"fmt"
	"strings"

	"github.com/cucumber/godog"
)

func (s *Suite) registerStory32Browser(sc *godog.ScenarioContext) {
	sc.Step(`^ich das Teilziel "([^"]*)" unter "([^"]*)" anlege$`, s.browserCreateChild)
	sc.Step(`^ich lege das Teilziel "([^"]*)" unter "([^"]*)" an$`, s.browserCreateChild)
	sc.Step(`^ich das Projekt "([^"]*)" mit dem Ziel "([^"]*)" in "([^"]*)" anlege$`, s.browserCreateProject32)
	sc.Step(`^ich lege das Projekt "([^"]*)" mit dem Ziel "([^"]*)" in "([^"]*)" an$`, s.browserCreateProject32)
	sc.Step(`^sehe ich den Zielpfad "([^"]*)" mit dem Projekt "([^"]*)"$`, s.browserProjectPath)
	sc.Step(`^ich sehe den Zielpfad "([^"]*)" mit dem Projekt "([^"]*)"$`, s.browserProjectPath)
	sc.Step(`^sehe ich beide Zielpfade und ihre Projektzuordnungen erneut$`, s.browserBothPaths)
	sc.Step(`^sehe ich den bisherigen Status von "([^"]*)" und "([^"]*)"$`, s.browserRememberStatuses)
	sc.Step(`^ich bei "([^"]*)" die Statusaktion öffne und "([^"]*)" speichere$`, s.browserSetStatus)
	sc.Step(`^sehe ich "([^"]*)" mit dem Status "([^"]*)"$`, s.browserGoalStatus)
	sc.Step(`^im Browser behält "([^"]*)" seinen bisherigen Status$`, s.browserGoalKeepsStatus)
	sc.Step(`^ich die Seite neu lade$`, s.browserReload)
	sc.Step(`^sehe ich denselben Status beider Ziele$`, s.browserSameStatuses)
	s.registerStory32BrowserErrors(sc)
}

func (s *Suite) registerStory32BrowserErrors(sc *godog.ScenarioContext) {
	sc.Step(`^ich das Formular für ein Teilziel unter "([^"]*)" öffne$`, s.browserChildForm)
	sc.Step(`^ich als Teilzielname "([^"]*)" eingebe$`, s.browserFillChild)
	sc.Step(`^ich das Teilzielformular speichere$`, s.browserSubmitChild)
	sc.Step(`^sehe ich einen Hinweis am Zielfeld "Name"$`, s.browserChildNameError)
	sc.Step(`^unter "([^"]*)" erscheint kein Teilziel$`, s.browserNoChild)
	sc.Step(`^sehe ich weder "([^"]*)" noch "([^"]*)"$`, s.browserNeither)
	sc.Step(`^ich sehe, dass der Zielbaum zu "([^"]*)" gehört$`, s.browserGoalsBelongTo)
}

func (s *Suite) browserGoalID(name string) (string, error) {
	for _, card := range s.page.GoalCards {
		if card.Name == name && card.ID != "" {
			return card.ID, nil
		}
	}
	return "", fmt.Errorf("Zielkarte %q fehlt: %+v", name, s.page.GoalCards)
}

func (s *Suite) rememberBrowserGoals() {
	for _, card := range s.page.GoalCards {
		if card.ID != "" {
			s.goals[s.goalKey(s.active, card.Name)] = card.ID
		}
	}
}

func (s *Suite) browserCreateChild(name, parent string) error {
	if err := s.browserOpenGoalsFor(s.active); err != nil {
		return err
	}
	s.rememberBrowserGoals()
	if err := s.browserChildForm(parent); err != nil {
		return err
	}
	if err := s.browserFillChild(name); err != nil {
		return err
	}
	if err := s.browserSubmitChild(); err != nil {
		return err
	}
	s.rememberBrowserGoals()
	if _, err := s.browserGoalID(name); err != nil {
		return err
	}
	return nil
}

func (s *Suite) browserChildSelector() string {
	return `form[data-action="child"][data-goal-id="` + s.createdID + `"]`
}

func (s *Suite) browserChildForm(parent string) error {
	id, err := s.browserGoalID(parent)
	if err != nil {
		return err
	}
	s.createdID = id
	return s.browserCommand(map[string]any{"check": s.browserChildSelector(), "mode": "goals"})
}

func (s *Suite) browserFillChild(name string) error {
	s.childName = name
	field := s.browserChildSelector() + ` input[name=name]`
	return s.browserCommand(map[string]any{"fill": name, "field": field, "mode": "goals"})
}

func (s *Suite) browserSubmitChild() error {
	mode := "goals"
	if strings.TrimSpace(s.childName) == "" {
		mode = "error"
	}
	return s.browserCommand(map[string]any{"submit": true, "form": s.browserChildSelector(), "mode": mode})
}

func (s *Suite) browserCreateProject32(name, goal, organization string) error {
	s.active = organization
	path := "/organisationen/" + s.organizations[organization] + "/projekte/neu"
	if err := s.browserNavigate(path, "project-form"); err != nil {
		return err
	}
	if err := s.browserCommand(map[string]any{"fill": name, "mode": "project-form"}); err != nil {
		return err
	}
	if err := s.browserCommand(map[string]any{"select": goal, "mode": "project-form"}); err != nil {
		return err
	}
	if err := s.browserCommand(map[string]any{"submit": true, "mode": "projects"}); err != nil {
		return err
	}
	return s.browserOpenGoalsFor(organization)
}

func (s *Suite) browserProjectPath(path, project string) error {
	path = strings.ReplaceAll(path, " > ", " → ")
	for _, item := range s.page.ProjectPaths {
		if strings.Contains(item, path) && strings.Contains(item, project) {
			return nil
		}
	}
	return fmt.Errorf("Projektpfad %q für %q fehlt im Browser: %+v", path, project, s.page.ProjectPaths)
}

func (s *Suite) browserBothPaths() error {
	if err := s.browserProjectPath("Wissensportal > Redaktion", "Artikel"); err != nil {
		return err
	}
	return s.browserProjectPath("Wissensportal > Technik", "Website")
}

func (s *Suite) browserStatus(name string) (string, error) {
	for _, card := range s.page.GoalCards {
		if card.Name == name {
			return card.Status, nil
		}
	}
	return "", fmt.Errorf("Zielstatus %q fehlt: %+v", name, s.page.GoalCards)
}

func (s *Suite) browserRememberStatuses(first, second string) error {
	for _, name := range []string{first, second} {
		status, err := s.browserStatus(name)
		if err != nil {
			return err
		}
		if status == "" {
			return fmt.Errorf("Status von %q leer", name)
		}
		s.baseline[s.goalKey(s.active, name)] = status
	}
	return nil
}

func (s *Suite) browserSetStatus(name, label string) error {
	id, err := s.browserGoalID(name)
	if err != nil {
		return err
	}
	form := `form[data-action="status"][data-goal-id="` + id + `"]`
	if err := s.browserCommand(map[string]any{"check": form, "mode": "goals"}); err != nil {
		return err
	}
	if err := s.browserCommand(map[string]any{"select": "Erreicht", "field": form + ` select[name=status]`, "mode": "goals"}); err != nil {
		return err
	}
	if label != "erreicht" {
		return fmt.Errorf("unbekanntes Statuslabel %q", label)
	}
	return s.browserCommand(map[string]any{"submit": true, "form": form, "mode": "goals"})
}

func (s *Suite) browserGoalStatus(name, label string) error {
	status, err := s.browserStatus(name)
	if err != nil {
		return err
	}
	if status != s.goalStatus(label) {
		return fmt.Errorf("Browserstatus %q für %q", status, name)
	}
	return nil
}

func (s *Suite) browserGoalKeepsStatus(name string) error {
	status, err := s.browserStatus(name)
	if err != nil {
		return err
	}
	if status != s.baseline[s.goalKey(s.active, name)] {
		return fmt.Errorf("Status von %q änderte sich zu %q", name, status)
	}
	return nil
}

func (s *Suite) browserReload() error {
	return s.browserNavigate("/organisationen/"+s.organizations[s.active]+"/ziele", "goals")
}

func (s *Suite) browserSameStatuses() error {
	if err := s.browserGoalStatus("Redaktion", "erreicht"); err != nil {
		return err
	}
	return s.browserGoalKeepsStatus("Wissensportal")
}

func (s *Suite) browserChildNameError() error {
	for _, item := range s.page.ChildErrors {
		if item.ID == s.createdID && strings.TrimSpace(item.Message) != "" {
			return nil
		}
	}
	return fmt.Errorf("Teilziel-Feldfehler fehlt: %+v", s.page.ChildErrors)
}

func (s *Suite) browserNoChild(parent string) error {
	if err := s.browserReload(); err != nil {
		return err
	}
	if _, err := s.browserGoalID(parent); err != nil {
		return err
	}
	if len(s.page.GoalCards) != 1 {
		return fmt.Errorf("unerwartetes Teilziel: %+v", s.page.GoalCards)
	}
	return nil
}

func (s *Suite) browserNeither(goal, project string) error {
	if strings.Contains(s.page.Text, goal) || strings.Contains(s.page.Text, project) {
		return fmt.Errorf("fremdes Ziel/Projekt im Browser: %+v", s.page)
	}
	return nil
}
