package datenbereiche

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"agentcontrolplane/app/internal/adapter/sqlite"
	apporganisation "agentcontrolplane/app/internal/app/organisation"
	appprojekt "agentcontrolplane/app/internal/app/projekt"
)

func (s *Suite) organizationWithProject(organization, project string) error {
	if err := s.ensureOrganization(organization); err != nil {
		return err
	}

	return s.ensureProject(project, organization)
}

func (s *Suite) ensureTwoProjects(first, second, organization string) error {
	if err := s.ensureProject(first, organization); err != nil {
		return err
	}

	return s.ensureProject(second, organization)
}

func (s *Suite) ensureGoal(organization string) error {
	if s.goals[organization] != "" {
		return nil
	}

	body, _ := json.Marshal(map[string]string{"name": "Testziel " + organization})
	path := "/api/organisationen/" + url.PathEscape(s.organizations[organization]) + "/ziele"
	if err := s.request(http.MethodPost, path, string(body), "application/json"); err != nil {
		return err
	}

	return s.rememberGoal(organization)
}

func (s *Suite) rememberGoal(organization string) error {
	if s.response.Status != http.StatusCreated {
		return fmt.Errorf("Ziel von %s: HTTP %d: %s", organization, s.response.Status, s.response.Body)
	}

	var goal Goal
	if err := json.Unmarshal(s.response.Body, &goal); err != nil {
		return err
	}

	s.goals[organization] = goal.ID
	return nil
}

func (s *Suite) ensureProject(project, organization string) error {
	if s.projects[project] != "" {
		return nil
	}

	if err := s.prepareProject(organization); err != nil {
		return err
	}

	return s.createProject(project, organization)
}

func (s *Suite) prepareProject(organization string) error {
	if err := s.ensureOrganization(organization); err != nil {
		return err
	}
	if err := s.ensureGoal(organization); err != nil {
		return err
	}

	return s.openStore()
}

func (s *Suite) createProject(project, organization string) error {
	organizations := apporganisation.NewService(sqlite.NewOrganizationStore(s.dataStore), apporganisation.NewLocalIdentity())
	service := appprojekt.NewService(s.dataStore, organizations)
	created, err := service.Create(context.Background(), s.organizations[organization], project, "Beschreibung von "+project, s.goals[organization])
	if err != nil {
		return err
	}

	s.projects[project] = created.ID
	return nil
}
