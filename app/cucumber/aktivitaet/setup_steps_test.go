package aktivitaet

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func (s *Suite) ensureOrganization(name string) error {
	if s.organizations[name] != "" {
		return nil
	}

	data, _ := json.Marshal(map[string]string{"name": name, "description": "Testorganisation"})
	if err := s.request("POST", "/api/organisationen", string(data)); err != nil {
		return err
	}

	var value struct {
		ID string `json:"id"`
	}
	if err := s.created(&value); err != nil {
		return err
	}

	s.organizations[name] = value.ID
	return nil
}

func (s *Suite) ensureProjectAndAgent(org, project, agent string) error {
	if err := s.ensureOrganization(org); err != nil {
		return err
	}

	if err := s.ensureProject(org, project); err != nil {
		return err
	}

	return s.ensureAgent(org, agent)
}

func (s *Suite) ensureProject(org, name string) error {
	goal, err := s.createGoal(org, name)
	if err != nil {
		return err
	}

	data, _ := json.Marshal(map[string]string{"name": name, "description": "Testprojekt", "goal_id": goal})
	path := "/api/organisationen/" + s.organizations[org] + "/projekte"
	if err := s.request("POST", path, string(data)); err != nil {
		return err
	}

	var value struct {
		ID string `json:"id"`
	}
	if err := s.created(&value); err != nil {
		return err
	}

	s.projects[s.key(org, name)] = value.ID
	return nil
}

func (s *Suite) createGoal(org, project string) (string, error) {
	data, _ := json.Marshal(map[string]string{"name": "Ziel " + project})
	path := "/api/organisationen/" + s.organizations[org] + "/ziele"
	if err := s.request("POST", path, string(data)); err != nil {
		return "", err
	}

	var value struct {
		ID string `json:"id"`
	}
	if err := s.created(&value); err != nil {
		return "", err
	}

	return value.ID, nil
}

func (s *Suite) ensureAgent(org, name string) error {
	data, _ := json.Marshal(map[string]string{"name": name, "template_id": "recherche", "execution_kind": "eino"})
	path := "/api/organisationen/" + s.organizations[org] + "/agenten"
	if err := s.request("POST", path, string(data)); err != nil {
		return err
	}

	var value struct {
		ID string `json:"id"`
	}
	if err := s.created(&value); err != nil {
		return err
	}

	s.agents[s.key(org, name)] = value.ID
	return nil
}

func (s *Suite) created(value any) error {
	if s.status != http.StatusCreated {
		return fmt.Errorf("HTTP %d statt 201: %s", s.status, s.body)
	}

	if err := json.Unmarshal(s.body, value); err != nil {
		return err
	}

	return nil
}

func (s *Suite) key(org, name string) string { return org + "\x00" + name }

func (s *Suite) taskPath(org, project string) string {
	return "/api/organisationen/" + s.organizations[org] + "/projekte/" + s.projects[s.key(org, project)] + "/aufgaben"
}
