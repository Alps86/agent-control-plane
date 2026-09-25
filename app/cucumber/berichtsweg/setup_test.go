package berichtsweg

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func (s *Suite) setupAPI() error { return s.fresh() }

func (s *Suite) ensureOrganizations(a, b string) error {
	if err := s.ensureOrganization(a); err != nil {
		return err
	}
	return s.ensureOrganization(b)
}

func (s *Suite) ensureOrganization(name string) error {
	if s.orgID(name) != "" {
		return nil
	}
	if err := s.call("POST", "/api/organisationen", map[string]string{"name": name, "description": "Testorganisation " + name}); err != nil {
		return err
	}
	if err := s.status(http.StatusCreated); err != nil {
		return err
	}
	var item Organization
	if err := json.Unmarshal(s.response.Body, &item); err != nil {
		return err
	}
	if item.ID == "" || item.Name != name {
		return fmt.Errorf("Organisation %s: %s", name, s.response.Body)
	}
	s.organizations[name] = item.ID
	return nil
}

func (s *Suite) ensureThreeAgents(a, b, c, org string) error {
	for _, name := range []string{a, b, c} {
		if err := s.ensureAgent(name, org); err != nil {
			return err
		}
	}
	return nil
}

func (s *Suite) ensureAgent(name, org string) error {
	if s.agentID(org, name) != "" {
		return nil
	}
	path := "/api/organisationen/" + s.orgID(org) + "/agenten"
	body := map[string]string{"name": name, "template_id": "recherche", "execution_kind": "eino"}
	if err := s.call("POST", path, body); err != nil {
		return err
	}
	if err := s.status(http.StatusCreated); err != nil {
		return err
	}
	var item Agent
	if err := json.Unmarshal(s.response.Body, &item); err != nil {
		return err
	}
	if item.ID == "" || item.Name != name {
		return fmt.Errorf("Agent %s: %s", name, s.response.Body)
	}
	s.agents[s.key(org, name)] = item.ID
	return nil
}

func (s *Suite) ensureProject(org string) error {
	if s.projects[org] != "" {
		return nil
	}
	if err := s.call("POST", "/api/organisationen/"+s.orgID(org)+"/ziele", map[string]string{"name": "Testziel"}); err != nil {
		return err
	}
	if err := s.status(http.StatusCreated); err != nil {
		return err
	}
	var goal struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(s.response.Body, &goal); err != nil {
		return err
	}
	body := map[string]string{"name": "Testprojekt", "description": "Testprojekt", "goal_id": goal.ID}
	if err := s.call("POST", "/api/organisationen/"+s.orgID(org)+"/projekte", body); err != nil {
		return err
	}
	if err := s.status(http.StatusCreated); err != nil {
		return err
	}
	var project Project
	if err := json.Unmarshal(s.response.Body, &project); err != nil {
		return err
	}
	s.projects[org] = project.ID
	return nil
}

func (s *Suite) ensureTask(title, agent, org string) error {
	if err := s.ensureProject(org); err != nil {
		return err
	}
	path := "/api/organisationen/" + s.orgID(org) + "/projekte/" + s.projects[org] + "/aufgaben"
	body := map[string]any{"title": title, "description": "Testaufgabe", "priority": "normal", "assignee_id": s.agentID(org, agent)}
	if err := s.call("POST", path, body); err != nil {
		return err
	}
	if err := s.status(http.StatusCreated); err != nil {
		return err
	}
	var task Task
	if err := json.Unmarshal(s.response.Body, &task); err != nil {
		return err
	}
	if task.ID == "" {
		return fmt.Errorf("Aufgabenkennung fehlt: %s", s.response.Body)
	}
	s.tasks[s.key(org, title)] = task.ID
	return nil
}

func (s *Suite) taskAssigned(title, agent string) error {
	org := "Nordstern"
	path := "/api/organisationen/" + s.orgID(org) + "/projekte/" + s.projects[org] + "/aufgaben/" + s.tasks[s.key(org, title)]
	if err := s.call("GET", path, nil); err != nil {
		return err
	}
	if err := s.status(http.StatusOK); err != nil {
		return err
	}
	var task Task
	if err := json.Unmarshal(s.response.Body, &task); err != nil {
		return err
	}
	if task.AssigneeID != s.agentID(org, agent) {
		return fmt.Errorf("Aufgabenzuweisung geändert: %+v", task)
	}
	return nil
}
