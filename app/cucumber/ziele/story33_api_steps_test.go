package ziele

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/cucumber/godog"
)

func (s *Suite) registerStory33API(sc *godog.ScenarioContext) {
	sc.Step(`^ich "([^"]*)" in "([^"]*)" über HTTP unter "([^"]*)" verschiebe$`, s.moveGoal33)
	sc.Step(`^ich "([^"]*)" in "([^"]*)" über HTTP unter "([^"]*)" verknüpfe$`, s.moveGoal33)
	sc.Step(`^ich die Elternziel-Verknüpfung von "([^"]*)" in "([^"]*)" über HTTP löse$`, s.detachGoal33)
	sc.Step(`^ich "([^"]*)" über die Organisation "([^"]*)" unter "([^"]*)" verschiebe$`, s.moveForeignGoal33)
	sc.Step(`^ich als nicht zugeordnete Betreiberin "([^"]*)" in "([^"]*)" unter "([^"]*)" verschiebe$`, s.unassignedMove33)
	s.registerStory33APIAssertions(sc)
}

func (s *Suite) registerStory33APIAssertions(sc *godog.ScenarioContext) {
	sc.Step(`^enthält der Zielbaum von "([^"]*)" den Pfad "([^"]*)" für das Projekt "([^"]*)"$`, s.treePath33)
	sc.Step(`^enthält der Zielbaum von "([^"]*)" weiterhin den Pfad "([^"]*)" für das Projekt "([^"]*)"$`, s.treePath33)
	sc.Step(`^der Zielbaum von "([^"]*)" enthält weiterhin den Pfad "([^"]*)" für das Projekt "([^"]*)"$`, s.treePath33)
	sc.Step(`^"([^"]*)" hat "([^"]*)" als übergeordnetes Ziel$`, s.goalParent33)
	sc.Step(`^"([^"]*)" hat weiterhin "([^"]*)" als übergeordnetes Ziel$`, s.goalParent33)
	sc.Step(`^"([^"]*)" hat das übergeordnete Ziel "([^"]*)"$`, s.goalParent33)
	sc.Step(`^hat "([^"]*)" das übergeordnete Ziel "([^"]*)"$`, s.goalParent33)
	sc.Step(`^"([^"]*)" ist wieder ein Stammziel$`, s.goalRoot33)
	sc.Step(`^ist "([^"]*)" wieder ein Stammziel$`, s.goalRoot33)
	sc.Step(`^"([^"]*)" bleibt ein Stammziel in "([^"]*)"$`, s.goalRootIn33)
	sc.Step(`^"([^"]*)" hat weiterhin den Status "([^"]*)"$`, s.goalHasStatus)
	sc.Step(`^der Zielbaum von "([^"]*)" enthält weiterhin alle vier Ziele genau einmal$`, s.fourGoals33)
	sc.Step(`^antwortet der Server mit einem Hinweis an der Kante "Übergeordnetes Ziel"$`, s.goalFieldError33)
	sc.Step(`^antwortet der Server ohne fremde Zieldaten mit HTTP-Status 404$`, s.notFound)
}

func (s *Suite) movePath33(organization, goal string) string {
	return s.goalPath(organization) + "/" + s.goals[s.goalKey(organization, goal)] + "/verschieben"
}

func (s *Suite) moveGoal33(goal, organization, parent string) error {
	s.active = organization
	parentID := s.goals[s.goalKey(organization, parent)]
	if parentID == "" {
		parentID = s.foreignGoalID(parent)
	}

	return s.sendMove33(s.movePath33(organization, goal), parentID)
}

func (s *Suite) sendMove33(path, parentID string) error {
	body, err := json.Marshal(map[string]string{"parent_goal_id": parentID})
	if err != nil {
		return err
	}

	return s.jsonRequest("POST", path, string(body))
}

func (s *Suite) detachGoal33(goal, organization string) error {
	s.active = organization
	return s.sendMove33(s.movePath33(organization, goal), "")
}

func (s *Suite) moveForeignGoal33(goal, organization, parent string) error {
	s.active = organization
	path := s.goalPath(organization) + "/" + s.foreignGoalID(goal) + "/verschieben"
	return s.sendMove33(path, s.goals[s.goalKey(organization, parent)])
}

func (s *Suite) unassignedMove33(goal, organization, parent string) error {
	if err := s.unassignedGoalList(organization); err != nil {
		return err
	}

	s.active = organization
	body := `{"parent_goal_id":"` + s.goals[s.goalKey(organization, parent)] + `"}`
	return s.negativeRequest("POST", s.movePath33(organization, goal), body)
}

func (s *Suite) treePath33(organization, path, project string) error {
	if err := s.getTree(organization); err != nil {
		return err
	}

	return s.treeProjectPath(path, project)
}

func (s *Suite) goalParent33(name, parent string) error {
	goal, err := s.treeGoal(name)
	if err != nil {
		return err
	}

	if goal.ParentGoalID == nil || *goal.ParentGoalID != s.goals[s.goalKey(s.active, parent)] {
		return fmt.Errorf("Elternziel von %q ist nicht %q: %+v", name, parent, goal)
	}

	return nil
}

func (s *Suite) goalRoot33(name string) error {
	goal, err := s.treeGoal(name)
	if err != nil {
		return err
	}

	if goal.ParentGoalID != nil {
		return fmt.Errorf("%q ist kein Stammziel: %+v", name, goal)
	}

	return nil
}

func (s *Suite) goalRootIn33(name, organization string) error {
	s.active = organization
	return s.goalRoot33(name)
}

func (s *Suite) fourGoals33(organization string) error {
	if err := s.getTree(organization); err != nil {
		return err
	}

	tree, err := s.parseTree()
	if err != nil {
		return err
	}

	seen := map[string]bool{}
	for _, goal := range tree.Goals {
		seen[goal.ID] = true
	}
	if len(tree.Goals) != 4 || len(seen) != 4 {
		return fmt.Errorf("Ziele fehlen oder doppelt: %s", s.response.Body)
	}

	return nil
}

func (s *Suite) goalFieldError33() error {
	if err := s.goalFieldError("Übergeordnetes Ziel"); err != nil {
		return err
	}

	if !strings.Contains(string(s.response.Body), "parent_goal_id") || s.response.Status != http.StatusUnprocessableEntity {
		return fmt.Errorf("benannte Elternziel-Kante fehlt: %s", s.response.Body)
	}

	return nil
}
