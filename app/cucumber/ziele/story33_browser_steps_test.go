package ziele

import (
	"fmt"
	"strings"

	"github.com/cucumber/godog"
)

func (s *Suite) registerStory33Browser(sc *godog.ScenarioContext) {
	sc.Step(`^ich bei "([^"]*)" das übergeordnete Ziel "([^"]*)" auswähle und die Verschiebung speichere$`, s.browserMove33)
	sc.Step(`^ich bei "([^"]*)" die Auswahl "Stammziel" als übergeordnetes Ziel speichere$`, s.browserDetach33)
	sc.Step(`^sehe ich "([^"]*)" als übergeordnetes Ziel von "([^"]*)"$`, s.browserParent33)
	sc.Step(`^ich sehe "([^"]*)" als übergeordnetes Ziel von "([^"]*)"$`, s.browserParent33)
	sc.Step(`^ich sehe "([^"]*)" weiterhin als übergeordnetes Ziel von "([^"]*)"$`, s.browserParent33)
	sc.Step(`^ich sehe "([^"]*)" als Stammziel$`, s.browserRoot33)
	sc.Step(`^sehe ich "([^"]*)" als Stammziel$`, s.browserRoot33)
	sc.Step(`^ich sehe "([^"]*)" weiterhin als Stammziel$`, s.browserRoot33)
	sc.Step(`^sehe ich "([^"]*)" weiterhin als Stammziel$`, s.browserRoot33)
	sc.Step(`^sehe ich einen Hinweis am Zielfeld "Übergeordnetes Ziel"$`, s.browserMoveError33)
}

func (s *Suite) browserMoveForm33(name string) (string, error) {
	id, err := s.browserGoalID(name)
	if err != nil {
		return "", err
	}

	return `form[data-action="reparent"][data-goal-id="` + id + `"]`, nil
}

func (s *Suite) browserMove33(name, parent string) error {
	s.rememberBrowserGoals()
	form, err := s.browserMoveForm33(name)
	if err != nil {
		return err
	}

	s.createdID, s.childName = s.goals[s.goalKey(s.active, name)], parent
	if err := s.browserCommand(map[string]any{"select": parent, "field": form + ` select[name=parent_goal_id]`, "mode": "goals"}); err != nil {
		return err
	}

	mode := "goals"
	if name == parent || s.browserDescendant33(name, parent) {
		mode = "error"
	}
	return s.browserCommand(map[string]any{"submit": true, "form": form, "mode": mode})
}

func (s *Suite) browserDescendant33(name, parent string) bool {
	lookup := map[string]string{}
	for _, card := range s.page.GoalCards {
		lookup[card.ID] = card.ParentID
	}
	ancestor := s.goals[s.goalKey(s.active, name)]
	for id := s.goals[s.goalKey(s.active, parent)]; id != ""; id = lookup[id] {
		if id == ancestor {
			return true
		}
	}
	return false
}

func (s *Suite) browserDetach33(name string) error {
	return s.browserMove33(name, "Stammziel")
}

func (s *Suite) browserParent33(parent, name string) error {
	goal, err := s.browserGoalCard33(name)
	if err != nil {
		return err
	}

	if goal.ParentID != s.goals[s.goalKey(s.active, parent)] {
		return fmt.Errorf("Browser-Elternziel von %q ist nicht %q: %+v", name, parent, goal)
	}

	return nil
}

func (s *Suite) browserGoalCard33(name string) (BrowserGoal, error) {
	for _, card := range s.page.GoalCards {
		if card.Name == name {
			return card, nil
		}
	}

	return BrowserGoal{}, fmt.Errorf("Browser-Ziel %q fehlt: %+v", name, s.page.GoalCards)
}

func (s *Suite) browserRoot33(name string) error {
	goal, err := s.browserGoalCard33(name)
	if err != nil {
		return err
	}

	if goal.ParentID != "" {
		return fmt.Errorf("%q ist im Browser kein Stammziel: %+v", name, goal)
	}

	return nil
}

func (s *Suite) browserMoveError33() error {
	for _, item := range s.page.MoveErrors {
		if item.ID == s.createdID && strings.TrimSpace(item.Message) != "" {
			return nil
		}
	}

	return fmt.Errorf("Elternziel-Feldhinweis fehlt: %+v", s.page.MoveErrors)
}
