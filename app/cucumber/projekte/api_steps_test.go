package projekte

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/cucumber/godog"
)

func (s *Suite) initializeScenario(sc *godog.ScenarioContext) {
	sc.After(s.afterScenario)
	s.registerAPISteps(sc)
	s.registerSecuritySteps(sc)
	s.registerStartupSteps(sc)
	s.registerRegressionSteps(sc)
	s.registerBrowserSteps(sc)
}

func (s *Suite) registerAPISteps(sc *godog.ScenarioContext) {
	s.registerAPISetup(sc)
	s.registerAPIProjectActions(sc)
	s.registerAPIAssertions(sc)
}

func (s *Suite) registerAPISetup(sc *godog.ScenarioContext) {
	sc.Step(`^ich starte den lokalen Server mit einer neuen SQLite-Datenbank als Betreiberin$`, s.freshServer)
	sc.Step(`^ich lege die Organisation "([^"]*)" über HTTP an$`, s.createOrganization)
	sc.Step(`^ich lege die Organisationen "([^"]*)" und "([^"]*)" über HTTP an$`, s.createOrganizations)
	sc.Step(`^ich lege das Stammziel "([^"]*)" in "([^"]*)" über HTTP an$`, s.createGoal)
}

func (s *Suite) registerAPIProjectActions(sc *godog.ScenarioContext) {
	sc.Step(`^ich die Projektübersicht von "([^"]*)" über HTTP abrufe$`, s.getProjects)
	sc.Step(`^antwortet der Server erfolgreich mit einer leeren Projektliste$`, s.emptyProjects)
	sc.Step(`^ich das Projektformular von "([^"]*)" über HTTP abrufe$`, s.getProjectForm)
	sc.Step(`^bietet es das Ziel "([^"]*)" zur Auswahl an$`, s.formOffersGoal)
	sc.Step(`^ich das Projekt "([^"]*)" mit der Beschreibung "([^"]*)" und dem Ziel "([^"]*)" in "([^"]*)" über HTTP anlege$`, s.createProject)
	sc.Step(`^ich lege das Projekt "([^"]*)" mit der Beschreibung "([^"]*)" und dem Ziel "([^"]*)" in "([^"]*)" über HTTP an$`, s.createProject)
	sc.Step(`^ich ein Projekt mit dem Namen "([^"]*)" und dem Ziel "([^"]*)" in "([^"]*)" über HTTP anlege$`, s.createProjectWithName)
	sc.Step(`^ich das Projekt "([^"]*)" ohne Zielzuordnung in "([^"]*)" über HTTP anlege$`, s.createWithoutGoal)
	sc.Step(`^ich das Projekt "([^"]*)" mit einer unbekannten Zielkennung in "([^"]*)" über HTTP anlege$`, s.createWithUnknownGoal)
	sc.Step(`^ich das Projekt "([^"]*)" mit dem Ziel "([^"]*)" aus "([^"]*)" in "([^"]*)" über HTTP anlege$`, s.createWithForeignGoal)
	sc.Step(`^ich das Projekt "([^"]*)" mit einer unbekannten Organisationskennung über HTTP anlege$`, s.createInUnknownOrganization)
}

func (s *Suite) registerAPIAssertions(sc *godog.ScenarioContext) {
	sc.Step(`^antwortet der Server mit einer neuen Projektkennung$`, s.createdProject)
	sc.Step(`^die Projektübersicht von "([^"]*)" enthält "([^"]*)" mit der Beschreibung "([^"]*)" genau einmal$`, s.projectWithDescriptionOnce)
	sc.Step(`^enthält die Projektübersicht von "([^"]*)" "([^"]*)" mit der Beschreibung "([^"]*)" genau einmal$`, s.projectWithDescriptionOnce)
	sc.Step(`^die Projektübersicht von "([^"]*)" enthält "([^"]*)" genau einmal$`, s.projectOnce)
	sc.Step(`^das Projektdetail von "([^"]*)" zeigt das Ziel "([^"]*)" aus "([^"]*)"$`, s.detailGoal)
	sc.Step(`^das Projektdetail von "([^"]*)" zeigt dieselbe Projektkennung und dasselbe Ziel "([^"]*)"$`, s.sameDetail)
	sc.Step(`^das Projektdetail von "([^"]*)" zeigt eine leere Aufgabenliste$`, s.emptyTasks)
	sc.Step(`^ich den Server beende und mit derselben SQLite-Datenbank erneut starte$`, s.restartServer)
	sc.Step(`^antwortet der Server mit einem Hinweis am Projektfeld "([^"]*)"$`, s.projectFieldError)
	sc.Step(`^die Projektübersicht von "([^"]*)" bleibt leer$`, s.projectsEmpty)
	sc.Step(`^das Ziel "([^"]*)" bleibt unverändert erhalten$`, s.goalStillExists)
	sc.Step(`^das Ziel "([^"]*)" bleibt unverändert in "([^"]*)" erhalten$`, s.goalStillInOrganization)
	sc.Step(`^es wird kein Projekt angelegt$`, s.noProject)
}

func (s *Suite) jsonRequest(method, path, body string) error {
	return s.request(method, path, body, http.Header{"Content-Type": []string{"application/json"}}, "")
}

func (s *Suite) createOrganization(name string) error {
	data, _ := json.Marshal(map[string]string{"name": name, "description": "Testorganisation"})
	if err := s.jsonRequest("POST", "/api/organisationen", string(data)); err != nil {
		return err
	}

	if err := s.statusCode(http.StatusCreated); err != nil {
		return err
	}

	return s.rememberOrganization(name)
}

func (s *Suite) rememberOrganization(name string) error {
	var organization Organization
	if err := json.Unmarshal(s.response.Body, &organization); err != nil {
		return err
	}

	if organization.ID == "" || organization.Name != name {
		return fmt.Errorf("Organisation nicht angelegt: %s", s.response.Body)
	}

	s.organizations[name], s.active = organization.ID, name
	return nil
}

func (s *Suite) createOrganizations(first, second string) error {
	if err := s.createOrganization(first); err != nil {
		return err
	}

	return s.createOrganization(second)
}

func (s *Suite) goalKey(organization, name string) string    { return organization + "\x00" + name }
func (s *Suite) projectKey(organization, name string) string { return organization + "\x00" + name }
func (s *Suite) goalPath(name string) string {
	return "/api/organisationen/" + s.organizations[name] + "/ziele"
}
func (s *Suite) projectPath(name string) string {
	return "/api/organisationen/" + s.organizations[name] + "/projekte"
}
func (s *Suite) detailPath(organization, project string) string {
	return s.projectPath(organization) + "/" + s.projects[s.projectKey(organization, project)]
}

func (s *Suite) createGoal(name, organization string) error {
	data, _ := json.Marshal(map[string]string{"name": name})
	if err := s.jsonRequest("POST", s.goalPath(organization), string(data)); err != nil {
		return err
	}

	if err := s.statusCode(http.StatusCreated); err != nil {
		return err
	}

	return s.rememberGoal(name, organization)
}

func (s *Suite) rememberGoal(name, organization string) error {
	var goal Goal
	if err := json.Unmarshal(s.response.Body, &goal); err != nil {
		return err
	}

	if goal.ID == "" || goal.OrganizationID != s.organizations[organization] || goal.Name != name {
		return fmt.Errorf("Ziel nicht angelegt: %s", s.response.Body)
	}

	s.goals[s.goalKey(organization, name)] = goal.ID
	return nil
}

func (s *Suite) createProject(name, description, goal, organization string) error {
	return s.postProject(organization, name, description, s.goals[s.goalKey(organization, goal)])
}

func (s *Suite) postProject(organization, name, description, goalID string) error {
	s.active = organization
	data, _ := json.Marshal(map[string]string{"name": name, "description": description, "goal_id": goalID})
	if err := s.jsonRequest("POST", s.projectPath(organization), string(data)); err != nil {
		return err
	}

	if s.response.Status == http.StatusCreated {
		var project Project
		if err := json.Unmarshal(s.response.Body, &project); err != nil {
			return err
		}

		s.projects[s.projectKey(organization, name)] = project.ID
		s.createdID = project.ID
	}

	return nil
}

func (s *Suite) createProjectWithName(name, goal, organization string) error {
	return s.createProject(name, "Nur Versuch", goal, organization)
}

func (s *Suite) createWithoutGoal(project, organization string) error {
	return s.postProject(organization, project, "Nur Versuch", "")
}

func (s *Suite) createWithUnknownGoal(project, organization string) error {
	return s.postProject(organization, project, "Nur Versuch", "unbekannt")
}

func (s *Suite) createWithForeignGoal(project, goal, goalOrganization, projectOrganization string) error {
	return s.postProject(projectOrganization, project, "Nur Versuch", s.goals[s.goalKey(goalOrganization, goal)])
}

func (s *Suite) createInUnknownOrganization(name string) error {
	data, _ := json.Marshal(map[string]string{"name": name, "description": "Nur Versuch", "goal_id": "unbekannt"})
	return s.jsonRequest("POST", "/api/organisationen/unbekannt/projekte", string(data))
}

func (s *Suite) getProjects(organization string) error {
	s.active = organization
	return s.jsonRequest("GET", s.projectPath(organization), "")
}

func (s *Suite) getProjectForm(organization string) error {
	s.active = organization
	return s.request("GET", "/organisationen/"+s.organizations[organization]+"/projekte/neu", "", nil, "")
}

func (s *Suite) formOffersGoal(goal string) error {
	if err := s.statusCode(http.StatusOK); err != nil {
		return err
	}

	body := string(s.response.Body)
	if !strings.Contains(body, goal) || !strings.Contains(body, s.goals[s.goalKey(s.active, goal)]) {
		return fmt.Errorf("Zieloption fehlt: %s", body)
	}

	return nil
}

func (s *Suite) statusCode(status int) error {
	if s.response == nil || s.response.Status != status {
		return fmt.Errorf("HTTP-Status %d erwartet: %+v", status, s.response)
	}

	return nil
}

func (s *Suite) parseProjects() (ProjectList, error) {
	var list ProjectList
	if err := s.statusCode(http.StatusOK); err != nil {
		return list, err
	}

	if err := json.Unmarshal(s.response.Body, &list); err != nil || list.Projects == nil {
		return list, fmt.Errorf("ungültige Projektliste: %s: %v", s.response.Body, err)
	}

	return list, nil
}

func (s *Suite) emptyProjects() error {
	list, err := s.parseProjects()
	if err != nil {
		return err
	}

	if len(list.Projects) != 0 {
		return fmt.Errorf("Projektliste nicht leer: %s", s.response.Body)
	}

	return nil
}

func (s *Suite) projectsEmpty(organization string) error {
	if err := s.getProjects(organization); err != nil {
		return err
	}

	return s.emptyProjects()
}

func (s *Suite) createdProject() error {
	if err := s.statusCode(http.StatusCreated); err != nil {
		return err
	}

	var project Project
	if err := json.Unmarshal(s.response.Body, &project); err != nil {
		return err
	}

	if project.ID == "" || project.ID != s.createdID || project.OrganizationID != s.organizations[s.active] || project.GoalID == "" {
		return fmt.Errorf("Projektkennung/Zielbezug fehlt: %s", s.response.Body)
	}

	return nil
}

func (s *Suite) projectOnce(organization, name string) error {
	return s.projectWithDescriptionOnce(organization, name, "")
}

func (s *Suite) projectWithDescriptionOnce(organization, name, description string) error {
	if err := s.getProjects(organization); err != nil {
		return err
	}

	list, err := s.parseProjects()
	if err != nil {
		return err
	}

	return s.countProject(list, organization, name, description)
}

func (s *Suite) countProject(list ProjectList, organization, name, description string) error {
	count := 0
	for _, project := range list.Projects {
		if project.Name == name {
			if project.OrganizationID != s.organizations[organization] || description != "" && project.Description != description {
				return fmt.Errorf("Projektfelder falsch: %+v", project)
			}

			count++
		}
	}

	if count != 1 {
		return fmt.Errorf("Projekt %q %d-mal: %s", name, count, s.response.Body)
	}

	return nil
}

func (s *Suite) getProjectDetail(name string) (Project, error) {
	var project Project
	organization := s.active
	if err := s.jsonRequest("GET", s.detailPath(organization, name), ""); err != nil {
		return project, err
	}

	if err := s.statusCode(http.StatusOK); err != nil {
		return project, err
	}

	err := json.Unmarshal(s.response.Body, &project)
	return project, err
}

func (s *Suite) detailGoal(projectName, goalName, organization string) error {
	s.active = organization
	project, err := s.getProjectDetail(projectName)
	if err != nil {
		return err
	}

	if project.GoalID != s.goals[s.goalKey(organization, goalName)] || project.OrganizationID != s.organizations[organization] || project.Name != projectName {
		return fmt.Errorf("Projektdetail/Zielbezug falsch: %+v", project)
	}

	return nil
}

func (s *Suite) sameDetail(projectName, goalName string) error {
	project, err := s.getProjectDetail(projectName)
	if err != nil {
		return err
	}

	if project.ID != s.createdID || project.GoalID != s.goals[s.goalKey(s.active, goalName)] {
		return fmt.Errorf("Projektkennung/Zielbezug nach Neustart falsch: %+v", project)
	}

	return nil
}

func (s *Suite) emptyTasks(projectName string) error {
	project, err := s.getProjectDetail(projectName)
	if err != nil {
		return err
	}

	if project.Tasks == nil || len(project.Tasks) != 0 {
		return fmt.Errorf("anfänglich leere Aufgabenliste fehlt: %s", s.response.Body)
	}

	return nil
}

func (s *Suite) projectFieldError(field string) error {
	if err := s.statusCode(http.StatusUnprocessableEntity); err != nil {
		return err
	}

	var failure FieldError
	if err := json.Unmarshal(s.response.Body, &failure); err != nil {
		return err
	}

	key := map[string]string{"Name": "name", "Ziel": "goal_id"}[field]
	if strings.TrimSpace(failure.FieldErrors[key]) == "" {
		return fmt.Errorf("Feldhinweis %q fehlt: %s", key, s.response.Body)
	}

	return nil
}

func (s *Suite) goalStillExists(goalName string) error {
	return s.goalStillInOrganization(goalName, s.active)
}

func (s *Suite) goalStillInOrganization(goalName, organization string) error {
	if err := s.jsonRequest("GET", s.goalPath(organization), ""); err != nil {
		return err
	}

	if err := s.statusCode(http.StatusOK); err != nil {
		return err
	}

	var list GoalList
	if err := json.Unmarshal(s.response.Body, &list); err != nil {
		return err
	}

	return s.assertGoalInList(list, goalName, organization)
}

func (s *Suite) assertGoalInList(list GoalList, goalName, organization string) error {
	for _, goal := range list.Goals {
		if goal.ID == s.goals[s.goalKey(organization, goalName)] && goal.Name == goalName && goal.OrganizationID == s.organizations[organization] {
			return nil
		}
	}

	return fmt.Errorf("Ziel verändert: %s", s.response.Body)
}

func (s *Suite) noProject() error {
	for name := range s.organizations {
		if err := s.projectsEmpty(name); err != nil {
			return err
		}
	}

	return nil
}
