package ziele

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/cucumber/godog"
)

func (s *Suite) registerStory32API(sc *godog.ScenarioContext) {
	sc.Step(`^ich lege das Teilziel "([^"]*)" unter "([^"]*)" in "([^"]*)" über HTTP an$`, s.createChild)
	sc.Step(`^ich das Teilziel "([^"]*)" unter "([^"]*)" in "([^"]*)" über HTTP anlege$`, s.createChild)
	sc.Step(`^ich das Teilziel mit dem Namen "([^"]*)" unter "([^"]*)" in "([^"]*)" über HTTP anlege$`, s.createChild)
	sc.Step(`^ich das Projekt "([^"]*)" mit dem Ziel "([^"]*)" in "([^"]*)" über HTTP anlege$`, s.createProject32)
	sc.Step(`^ich lege das Projekt "([^"]*)" mit dem Ziel "([^"]*)" in "([^"]*)" über HTTP an$`, s.createProject32)
	sc.Step(`^ich den Zielbaum von "([^"]*)" über HTTP abrufe$`, s.getTree)
	sc.Step(`^enthält er den Pfad "([^"]*)" für das Projekt "([^"]*)"$`, s.treeProjectPath)
	sc.Step(`^er enthält den Pfad "([^"]*)" für das Projekt "([^"]*)"$`, s.treeProjectPath)
	sc.Step(`^beide Teilziele haben "([^"]*)" als übergeordnetes Ziel$`, s.childrenHaveParent)
	sc.Step(`^ich den Status von "([^"]*)" in "([^"]*)" über HTTP ausdrücklich auf "([^"]*)" setze$`, s.setGoalStatus)
	sc.Step(`^zeigt der Zielbaum "([^"]*)" mit dem Status "([^"]*)"$`, s.goalHasStatus)
	sc.Step(`^der Zielbaum zeigt "([^"]*)" weiterhin mit dem Status "([^"]*)"$`, s.goalHasStatus)
	sc.Step(`^zeigt der Zielbaum "([^"]*)" weiterhin mit dem Status "([^"]*)"$`, s.goalHasStatus)
	sc.Step(`^zeigt der Zielbaum auch "([^"]*)" mit dem Status "([^"]*)"$`, s.goalHasStatus)
	sc.Step(`^behalten beide Teilziele den Status "([^"]*)"$`, s.childrenHaveStatus)
	sc.Step(`^"([^"]*)" und "([^"]*)" behalten ihren bisherigen Status$`, s.goalsKeepStatus)
	sc.Step(`^"([^"]*)" und "([^"]*)" haben weiterhin ihren bisherigen Status$`, s.goalsKeepStatus)
	sc.Step(`^"([^"]*)" behält seinen bisherigen Status$`, s.goalKeepsStatus)
	s.registerStory32APIErrors(sc)
}

func (s *Suite) registerStory32APIErrors(sc *godog.ScenarioContext) {
	sc.Step(`^antwortet der Server mit einem Hinweis am Zielfeld "([^"]*)"$`, s.goalFieldError)
	sc.Step(`^der Zielbaum von "([^"]*)" bleibt leer$`, s.treeEmpty)
	sc.Step(`^der Zielbaum von "([^"]*)" enthält nur das unveränderte Stammziel "([^"]*)"$`, s.onlyRoot)
	sc.Step(`^er enthält weder "([^"]*)" noch "([^"]*)"$`, s.treeExcludes)
	sc.Step(`^enthält er weder "([^"]*)" noch "([^"]*)"$`, s.treeExcludes)
	sc.Step(`^der Zielbaum von "([^"]*)" zeigt "([^"]*)" nur am Ziel "([^"]*)"$`, s.treeProjectAtGoal)
	sc.Step(`^ich als nicht zugeordnete Betreiberin den Zielbaum von "([^"]*)" über HTTP abrufe$`, s.unassignedTree)
	sc.Step(`^ich als nicht zugeordnete Betreiberin den Status von "([^"]*)" in "([^"]*)" über HTTP ändere$`, s.unassignedStatus)
	sc.Step(`^der Status von "([^"]*)" bleibt unverändert$`, s.goalKeepsStatus)
}

func (s *Suite) goalKey(organization, name string) string { return organization + "\x00" + name }
func (s *Suite) treePath(name string) string {
	return "/api/organisationen/" + s.organizations[name] + "/zielbaum"
}
func (s *Suite) childPath(organization, parent string) string {
	return s.goalPath(organization) + "/" + s.goals[s.goalKey(organization, parent)] + "/kinder"
}

func (s *Suite) statusPath(organization, goal string) string {
	return s.goalPath(organization) + "/" + s.goals[s.goalKey(organization, goal)] + "/status"
}

func (s *Suite) createChild(name, parent, organization string) error {
	s.active = organization
	parentID := s.goals[s.goalKey(organization, parent)]
	if parentID == "" {
		parentID = s.foreignGoalID(parent)
	}
	data, _ := json.Marshal(map[string]string{"name": name})
	path := s.goalPath(organization) + "/" + parentID + "/kinder"
	if err := s.jsonRequest("POST", path, string(data)); err != nil {
		return err
	}
	if s.response.Status != http.StatusCreated {
		return nil
	}
	return s.rememberChild(name, organization, parentID)
}

func (s *Suite) foreignGoalID(name string) string {
	for key, id := range s.goals {
		if strings.HasSuffix(key, "\x00"+name) {
			return id
		}
	}
	return "unbekannt"
}

func (s *Suite) rememberChild(name, organization, parentID string) error {
	var goal Goal
	if err := json.Unmarshal(s.response.Body, &goal); err != nil {
		return err
	}
	if goal.ID == "" || goal.Name != name || goal.OrganizationID != s.organizations[organization] || goal.ParentGoalID == nil || *goal.ParentGoalID != parentID {
		return fmt.Errorf("Teilzielantwort falsch: %s", s.response.Body)
	}
	s.goals[s.goalKey(organization, name)] = goal.ID
	return nil
}

func (s *Suite) createProject32(name, goal, organization string) error {
	data, _ := json.Marshal(map[string]string{"name": name, "description": "Projektzielpfad", "goal_id": s.goals[s.goalKey(organization, goal)]})
	path := "/api/organisationen/" + s.organizations[organization] + "/projekte"
	if err := s.jsonRequest("POST", path, string(data)); err != nil {
		return err
	}
	if err := s.statusCode(http.StatusCreated); err != nil {
		return err
	}
	return s.rememberProject32(name, goal, organization)
}

func (s *Suite) rememberProject32(name, goal, organization string) error {
	var project struct {
		ID     string `json:"id"`
		GoalID string `json:"goal_id"`
	}
	if err := json.Unmarshal(s.response.Body, &project); err != nil {
		return err
	}
	if project.ID == "" || project.GoalID != s.goals[s.goalKey(organization, goal)] {
		return fmt.Errorf("Projektantwort falsch: %s", s.response.Body)
	}
	s.projects[s.goalKey(organization, name)] = project.ID
	return nil
}

func (s *Suite) getTree(name string) error {
	s.active = name
	return s.jsonRequest("GET", s.treePath(name), "")
}

func (s *Suite) parseTree() (GoalTree, error) {
	var tree GoalTree
	if err := s.statusCode(http.StatusOK); err != nil {
		return tree, err
	}
	if err := json.Unmarshal(s.response.Body, &tree); err != nil {
		return tree, err
	}
	if tree.Goals == nil || tree.ProjectPaths == nil {
		return tree, fmt.Errorf("Zielbaum-Felder fehlen: %s", s.response.Body)
	}
	return tree, nil
}

func (s *Suite) treeProjectPath(path, projectName string) error {
	tree, err := s.parseTree()
	if err != nil {
		return err
	}
	names := strings.Split(path, " > ")
	for _, item := range tree.ProjectPaths {
		if item.ProjectName == projectName && item.ProjectID == s.projects[s.goalKey(s.active, projectName)] && reflect.DeepEqual(item.GoalNames, names) {
			return nil
		}
	}
	return fmt.Errorf("Projektpfad %q für %q fehlt: %s", path, projectName, s.response.Body)
}

func (s *Suite) childrenHaveParent(parent string) error {
	tree, err := s.parseTree()
	if err != nil {
		return err
	}
	parentID, count := s.goals[s.goalKey(s.active, parent)], 0
	for _, goal := range tree.Goals {
		if goal.ParentGoalID == nil {
			continue
		}
		if *goal.ParentGoalID != parentID {
			return fmt.Errorf("falscher Elternknoten: %+v", goal)
		}
		count++
	}
	if count != 2 {
		return fmt.Errorf("%d Teilziele statt zwei: %s", count, s.response.Body)
	}
	return nil
}

func (s *Suite) setGoalStatus(goal, organization, label string) error {
	status := s.goalStatus(label)
	if status == "" {
		return fmt.Errorf("unbekannter Teststatus %q", label)
	}
	data, _ := json.Marshal(map[string]string{"status": status})
	if err := s.jsonRequest("POST", s.statusPath(organization, goal), string(data)); err != nil {
		return err
	}
	if err := s.statusCode(http.StatusOK); err != nil {
		return err
	}
	var updated Goal
	if err := json.Unmarshal(s.response.Body, &updated); err != nil {
		return err
	}
	if updated.Status != status || updated.ID != s.goals[s.goalKey(organization, goal)] {
		return fmt.Errorf("Statusantwort falsch: %s", s.response.Body)
	}
	return nil
}

func (s *Suite) goalStatus(label string) string {
	if label == "erreicht" {
		return "achieved"
	}
	return ""
}

func (s *Suite) treeGoal(name string) (Goal, error) {
	if err := s.getTree(s.active); err != nil {
		return Goal{}, err
	}
	tree, err := s.parseTree()
	if err != nil {
		return Goal{}, err
	}
	for _, goal := range tree.Goals {
		if goal.ID == s.goals[s.goalKey(s.active, name)] {
			return goal, nil
		}
	}
	return Goal{}, fmt.Errorf("Ziel %q fehlt im Baum: %s", name, s.response.Body)
}

func (s *Suite) goalHasStatus(name, label string) error {
	goal, err := s.treeGoal(name)
	if err != nil {
		return err
	}
	if goal.Status != s.goalStatus(label) {
		return fmt.Errorf("Status von %q ist %q", name, goal.Status)
	}
	return nil
}

func (s *Suite) goalKeepsStatus(name string) error {
	goal, err := s.treeGoal(name)
	if err != nil {
		return err
	}
	if goal.Status != "planned" {
		return fmt.Errorf("Status von %q änderte sich auf %q", name, goal.Status)
	}
	return nil
}

func (s *Suite) goalsKeepStatus(first, second string) error {
	if err := s.goalKeepsStatus(first); err != nil {
		return err
	}
	return s.goalKeepsStatus(second)
}

func (s *Suite) childrenHaveStatus(label string) error {
	if err := s.goalHasStatus("Redaktion", label); err != nil {
		return err
	}
	return s.goalHasStatus("Technik", label)
}

func (s *Suite) goalFieldError(field string) error {
	if err := s.statusCode(http.StatusUnprocessableEntity); err != nil {
		return err
	}
	var failure FieldError
	if err := json.Unmarshal(s.response.Body, &failure); err != nil {
		return err
	}
	key := map[string]string{"Name": "name", "Übergeordnetes Ziel": "parent_goal_id"}[field]
	if strings.TrimSpace(failure.FieldErrors[key]) == "" {
		return fmt.Errorf("Feldhinweis %q fehlt: %s", key, s.response.Body)
	}
	return nil
}

func (s *Suite) onlyRoot(organization, root string) error {
	if err := s.getTree(organization); err != nil {
		return err
	}
	tree, err := s.parseTree()
	if err != nil {
		return err
	}
	if len(tree.Goals) != 1 || tree.Goals[0].Name != root || tree.Goals[0].ParentGoalID != nil {
		return fmt.Errorf("Zielbaum verändert: %s", s.response.Body)
	}
	return nil
}

func (s *Suite) treeEmpty(organization string) error {
	if err := s.getTree(organization); err != nil {
		return err
	}
	tree, err := s.parseTree()
	if err != nil {
		return err
	}
	if len(tree.Goals) != 0 || len(tree.ProjectPaths) != 0 {
		return fmt.Errorf("Zielbaum nicht leer: %s", s.response.Body)
	}
	return nil
}

func (s *Suite) treeExcludes(goalName, projectName string) error {
	tree, err := s.parseTree()
	if err != nil {
		return err
	}
	for _, goal := range tree.Goals {
		if goal.Name == goalName {
			return fmt.Errorf("fremdes Ziel sichtbar: %s", s.response.Body)
		}
	}
	for _, project := range tree.ProjectPaths {
		if project.ProjectName == projectName {
			return fmt.Errorf("fremdes Projekt sichtbar: %s", s.response.Body)
		}
	}
	return nil
}

func (s *Suite) treeProjectAtGoal(organization, projectName, goalName string) error {
	if err := s.getTree(organization); err != nil {
		return err
	}
	return s.treeProjectPath(goalName, projectName)
}

func (s *Suite) unassignedTree(organization string) error {
	if err := s.unassignedGoalList(organization); err != nil {
		return err
	}
	return s.negativeRequest("GET", s.treePath(organization), "")
}

func (s *Suite) unassignedStatus(goal, organization string) error {
	s.active = organization
	return s.negativeRequest("POST", s.statusPath(organization, goal), `{"status":"achieved"}`)
}
