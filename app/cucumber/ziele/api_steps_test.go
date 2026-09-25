package ziele

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
	s.registerBrowserSteps(sc)
	s.registerStory32API(sc)
	s.registerStory32Browser(sc)
	s.registerStory33API(sc)
	s.registerStory33Browser(sc)
}

func (s *Suite) registerAPISteps(sc *godog.ScenarioContext) {
	sc.Step(`^ich starte den lokalen Server mit einer neuen SQLite-Datenbank als Betreiberin$`, s.freshServer)
	sc.Step(`^ich lege die Organisation "([^"]*)" über HTTP an$`, s.createOrganization)
	sc.Step(`^ich lege die Organisationen "([^"]*)" und "([^"]*)" über HTTP an$`, s.createOrganizations)
	sc.Step(`^ich die Zielübersicht von "([^"]*)" über HTTP abrufe$`, s.getGoals)
	sc.Step(`^antwortet der Server erfolgreich mit einer leeren Zielliste$`, s.emptyGoals)
	sc.Step(`^ich das Stammziel "([^"]*)" in "([^"]*)" über HTTP anlege$`, s.createGoal)
	sc.Step(`^ich lege das Stammziel "([^"]*)" in "([^"]*)" über HTTP an$`, s.createGoal)
	sc.Step(`^ich ein Stammziel mit dem Namen "([^"]*)" in "([^"]*)" über HTTP anlege$`, s.createGoal)
	sc.Step(`^ich das Stammziel "([^"]*)" mit einer unbekannten Organisationskennung über HTTP anlege$`, s.createUnknownGoal)
	s.registerAPIAssertions(sc)
}

func (s *Suite) registerAPIAssertions(sc *godog.ScenarioContext) {
	sc.Step(`^antwortet der Server mit einer neuen Zielkennung$`, s.createdGoal)
	sc.Step(`^die Zielübersicht von "([^"]*)" enthält "([^"]*)" genau einmal$`, s.goalOnce)
	sc.Step(`^enthält die Zielübersicht von "([^"]*)" "([^"]*)" genau einmal$`, s.goalOnce)
	sc.Step(`^das Ziel hat keinen übergeordneten Ziel- oder Projektbezug$`, s.rootGoal)
	sc.Step(`^das Ziel hat dieselbe Kennung und keinen übergeordneten Ziel- oder Projektbezug$`, s.sameRootGoal)
	sc.Step(`^ich den Server beende und mit derselben SQLite-Datenbank erneut starte$`, s.restartServer)
	sc.Step(`^antwortet der Server mit einem Hinweis am Feld "Name"$`, s.nameError)
	sc.Step(`^die Zielübersicht von "([^"]*)" bleibt leer$`, s.goalsEmpty)
	sc.Step(`^die Organisation "([^"]*)" bleibt unverändert erhalten$`, s.organizationStillExists)
	s.registerWildcardSteps(sc)
	s.registerSecuritySteps(sc)
}

func (s *Suite) registerWildcardSteps(sc *godog.ScenarioContext) {
	sc.Step(`^ich den Server beende und einen Start mit derselben SQLite-Datenbank auf der Wildcard-Adresse "0\.0\.0\.0" versuche$`, s.restartWildcard)
	sc.Step(`^wird der Zielserver wegen der nicht lokalen APP_ADDR-Konfiguration nicht gestartet$`, s.wildcardConfigRejected)
	sc.Step(`^auf dem Wildcard-Port ist kein Zielserver erreichbar$`, s.noWildcardListener)
	sc.Step(`^ich den Server beende und mit derselben SQLite-Datenbank auf der Loopback-Adresse erneut starte$`, s.restartLoopback)
	sc.Step(`^bleibt die Zielübersicht von "([^"]*)" leer$`, s.goalsEmpty)
}

func (s *Suite) registerSecuritySteps(sc *godog.ScenarioContext) {
	sc.Step(`^antwortet der Server mit einem datenfreien HTTP-Status 404$`, s.notFound)
	sc.Step(`^antwortet der Server mit demselben datenfreien HTTP-Status 404$`, s.sameNotFound)
	sc.Step(`^es wird kein Ziel angelegt$`, s.noGoal)
	sc.Step(`^der Server für einen HTTP-Aufruf keine eindeutige Betreiberidentität ermittelt$`, s.noIdentity)
	sc.Step(`^verweigert er das Lesen der Zielübersicht von "([^"]*)" ohne Zieldaten$`, s.deniedGoalList)
	sc.Step(`^er verweigert das Anlegen eines Stammziels in "([^"]*)" ohne Teilerzeugung$`, s.deniedGoalCreate)
	sc.Step(`^ich als nicht zugeordnete Betreiberin die Zielübersicht von "([^"]*)" über HTTP abrufe$`, s.unassignedGoalList)
	sc.Step(`^ich als nicht zugeordnete Betreiberin ein Stammziel in "([^"]*)" über HTTP anlege$`, s.unassignedGoalCreate)
	sc.Step(`^ich von Origin "([^"]*)" ein Stammziel "([^"]*)" in "([^"]*)" über HTTP anlege$`, s.foreignOrigin)
	sc.Step(`^ich über Host "([^"]*)" ein Stammziel "([^"]*)" in "([^"]*)" über HTTP anlege$`, s.foreignHost)
	sc.Step(`^antwortet der Server mit dem HTTP-Status (\d+)$`, s.statusCode)
}

func (s *Suite) jsonRequest(method, path, body string) error {
	headers := http.Header{"Content-Type": []string{"application/json"}}
	return s.request(method, path, body, headers, "")
}

func (s *Suite) createOrganization(name string) error {
	body, err := json.Marshal(map[string]string{"name": name, "description": "Testorganisation"})
	if err != nil {
		return err
	}

	if err := s.jsonRequest("POST", "/api/organisationen", string(body)); err != nil {
		return err
	}

	if s.response.Status != http.StatusCreated {
		return fmt.Errorf("Organisation %q: HTTP %d: %s", name, s.response.Status, s.response.Body)
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

func (s *Suite) goalPath(name string) string {
	return "/api/organisationen/" + s.organizations[name] + "/ziele"
}

func (s *Suite) getGoals(name string) error {
	s.active = name
	return s.jsonRequest("GET", s.goalPath(name), "")
}

func (s *Suite) createGoal(goal, organization string) error {
	s.active = organization
	body, err := json.Marshal(map[string]string{"name": goal})
	if err != nil {
		return err
	}

	if err := s.jsonRequest("POST", s.goalPath(organization), string(body)); err != nil {
		return err
	}

	if s.response.Status == http.StatusCreated {
		if err := s.rememberGoal(); err != nil {
			return err
		}

		s.goals[s.goalKey(organization, goal)] = s.createdID
	}

	return nil
}

func (s *Suite) rememberGoal() error {
	var created Goal
	if err := json.Unmarshal(s.response.Body, &created); err != nil {
		return err
	}

	s.createdID = created.ID
	return nil
}

func (s *Suite) createUnknownGoal(goal string) error {
	body, err := json.Marshal(map[string]string{"name": goal})
	if err != nil {
		return err
	}

	return s.jsonRequest("POST", "/api/organisationen/unbekannt/ziele", string(body))
}

func (s *Suite) statusCode(status int) error {
	if s.response == nil || s.response.Status != status {
		return fmt.Errorf("HTTP-Status %d erwartet: %+v", status, s.response)
	}

	return nil
}

func (s *Suite) parseGoals() (GoalList, error) {
	var list GoalList
	if err := s.statusCode(http.StatusOK); err != nil {
		return list, err
	}

	err := json.Unmarshal(s.response.Body, &list)
	if err != nil || list.Goals == nil {
		return list, fmt.Errorf("ungültige Zielliste %s: %w", s.response.Body, err)
	}

	return list, nil
}

func (s *Suite) emptyGoals() error {
	list, err := s.parseGoals()
	if err != nil {
		return err
	}

	if len(list.Goals) != 0 {
		return fmt.Errorf("Zielliste nicht leer: %s", s.response.Body)
	}

	return nil
}

func (s *Suite) goalsEmpty(name string) error {
	if err := s.getGoals(name); err != nil {
		return err
	}

	return s.emptyGoals()
}

func (s *Suite) createdGoal() error {
	if err := s.statusCode(http.StatusCreated); err != nil {
		return err
	}

	var goal Goal
	if err := json.Unmarshal(s.response.Body, &goal); err != nil {
		return err
	}

	if goal.ID == "" || goal.OrganizationID != s.organizations[s.active] {
		return fmt.Errorf("Zielkennung/Organisation fehlt: %s", s.response.Body)
	}

	if s.response.Location != s.goalPath(s.active) {
		return fmt.Errorf("Location %q", s.response.Location)
	}

	s.createdID = goal.ID
	return nil
}

func (s *Suite) goalOnce(organization, goalName string) error {
	if err := s.getGoals(organization); err != nil {
		return err
	}

	list, err := s.parseGoals()
	if err != nil {
		return err
	}

	return s.countGoal(list, organization, goalName)
}

func (s *Suite) countGoal(list GoalList, organization, goalName string) error {
	count := 0
	for _, goal := range list.Goals {
		if goal.Name == goalName && goal.OrganizationID == s.organizations[organization] {
			count++
		}
	}

	if count != 1 {
		return fmt.Errorf("Ziel %q %d-mal in %s", goalName, count, s.response.Body)
	}

	return nil
}

func (s *Suite) rootGoal() error {
	list, err := s.parseGoals()
	if err != nil {
		return err
	}

	for _, goal := range list.Goals {
		if goal.ID == s.createdID {
			if goal.ParentGoalID != nil || goal.ProjectID != nil {
				return fmt.Errorf("kein Stammziel: %+v", goal)
			}

			return nil
		}
	}

	return fmt.Errorf("Ziel %q nicht in Übersicht", s.createdID)
}

func (s *Suite) sameRootGoal() error {
	return s.rootGoal()
}

func (s *Suite) nameError() error {
	if err := s.statusCode(http.StatusUnprocessableEntity); err != nil {
		return err
	}

	var failure FieldError
	if err := json.Unmarshal(s.response.Body, &failure); err != nil {
		return err
	}

	if strings.TrimSpace(failure.FieldErrors["name"]) == "" {
		return fmt.Errorf("Feldhinweis fehlt: %s", s.response.Body)
	}

	return nil
}

func (s *Suite) organizationStillExists(name string) error {
	if err := s.jsonRequest("GET", "/api/organisationen/"+s.organizations[name], ""); err != nil {
		return err
	}

	if err := s.statusCode(http.StatusOK); err != nil {
		return err
	}

	var organization Organization
	if err := json.Unmarshal(s.response.Body, &organization); err != nil {
		return err
	}

	if organization.ID != s.organizations[name] || organization.Name != name {
		return fmt.Errorf("Organisation verändert: %s", s.response.Body)
	}

	return nil
}
