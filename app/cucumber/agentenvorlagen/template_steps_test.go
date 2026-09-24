package agentenvorlagen

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

func (s *Suite) getTemplate(name, org string) error {
	if err := s.request("GET", s.orgPath(org)+"/vorlagen", "", ""); err != nil {
		return err
	}
	if err := s.statusCode(http.StatusOK); err != nil {
		return err
	}
	s.selectedAgent = name
	return nil
}

func (s *Suite) templates() ([]Template, error) {
	var result struct {
		Templates []Template `json:"templates"`
	}
	err := json.Unmarshal(s.response.Body, &result)
	return result.Templates, err
}

func (s *Suite) templateHasProfile() error {
	items, err := s.templates()
	if err != nil {
		return err
	}
	for _, item := range items {
		if item.Name == s.selectedAgent && item.Role != "" && item.Instructions != "" && len(item.Capabilities) > 0 && item.ExecutionKind == "eino" {
			return nil
		}
	}
	return fmt.Errorf("Vorlagenprofil fehlt: %s", s.response.Body)
}

func (s *Suite) getTeamTemplate(team, org string) error {
	if err := s.request("GET", s.orgPath(org)+"/teamvorlagen", "", ""); err != nil {
		return err
	}
	if err := s.statusCode(http.StatusOK); err != nil {
		return err
	}
	s.selectedTeam = team
	s.selectedOrg = org
	return nil
}

func (s *Suite) teamIncludesTemplate(name string) error {
	var result struct {
		Teams []TeamTemplate `json:"team_templates"`
	}
	if err := json.Unmarshal(s.response.Body, &result); err != nil {
		return err
	}
	for _, item := range result.Teams {
		if item.Name == s.selectedTeam && len(item.TemplateIDs) == 1 && item.TemplateIDs[0] == strings.ToLower(name) {
			s.selectedAgent = name
			_, err := s.loadTemplate(s.selectedOrg, name)
			return err
		}
	}
	return fmt.Errorf("Teamvorlage fehlt: %s", s.response.Body)
}

func (s *Suite) loadTemplate(org, name string) (Template, error) {
	path := s.orgPath(org) + "/vorlagen"
	if err := s.request("GET", path, "", ""); err != nil {
		return Template{}, err
	}
	if err := s.statusCode(http.StatusOK); err != nil {
		return Template{}, err
	}
	templates, err := s.templates()
	if err != nil {
		return Template{}, err
	}
	for _, template := range templates {
		if template.ID == strings.ToLower(name) && template.Name == name && strings.TrimSpace(template.Role) != "" && strings.TrimSpace(template.Instructions) != "" {
			return template, nil
		}
	}
	return Template{}, fmt.Errorf("Rolle oder Auftrag der Teamvorlage fehlt: %s", s.response.Body)
}

func (s *Suite) createFromTeam(name, org string) error {
	if err := s.createAgent(name, s.selectedAgent, org); err != nil {
		return err
	}
	return s.createdAgent()
}

func (s *Suite) listExactlyAgent(org, agent string) error {
	list, err := s.getAgentList(org)
	if err != nil {
		return err
	}
	if len(list.Agents) != 1 || list.Agents[0].Name != agent {
		return fmt.Errorf("Liste enthält nicht nur %s: %+v", agent, list.Agents)
	}
	return nil
}

func (s *Suite) noTeamImport() error {
	org := s.orgName(s.initialProfile.OrganizationID)
	return s.listExactlyAgent(org, s.initialProfile.Name)
}

func (s *Suite) createWithCapability(template, capability string) error {
	if err := s.ensureOrganization("Nordstern"); err != nil {
		return err
	}
	body, _ := json.Marshal(map[string]any{"name": "Gefaehrlich", "template_id": strings.ToLower(template), "execution_kind": "eino", "additional_capabilities": []string{capability}})
	return s.request("POST", s.orgPath("Nordstern"), string(body), "application/json")
}

func (s *Suite) capabilityDenied() error {
	if s.response.Status != http.StatusUnprocessableEntity && s.response.Status != http.StatusForbidden {
		return fmt.Errorf("Fähigkeit nicht abgewiesen: HTTP %d: %s", s.response.Status, s.response.Body)
	}
	if !strings.Contains(string(s.response.Body), "capabilit") && !strings.Contains(string(s.response.Body), "Fähigkeit") {
		return fmt.Errorf("Fähigkeitsfehler fehlt: %s", s.response.Body)
	}
	return nil
}

func (s *Suite) noAgentFromRequest(org string) error {
	list, err := s.getAgentList(org)
	if err != nil {
		return err
	}
	for _, item := range list.Agents {
		if item.Name == "Gefaehrlich" {
			return fmt.Errorf("unzulässiger Agent persistiert: %+v", item)
		}
	}
	return nil
}

func (s *Suite) fieldError(field string, status int) error {
	if err := s.statusCode(status); err != nil {
		return err
	}
	var failure struct {
		FieldErrors map[string]string `json:"field_errors"`
	}
	if err := json.Unmarshal(s.response.Body, &failure); err != nil {
		return err
	}
	if failure.FieldErrors[field] == "" {
		return fmt.Errorf("Feldfehler %s fehlt: %s", field, s.response.Body)
	}
	return nil
}

func (s *Suite) nameError() error { return s.fieldError("name", http.StatusUnprocessableEntity) }
func (s *Suite) templateError() error {
	return s.fieldError("template_id", http.StatusUnprocessableEntity)
}
func (s *Suite) nameConflict() error { return s.fieldError("name", http.StatusConflict) }

func (s *Suite) listEmpty(org string) error {
	list, err := s.getAgentList(org)
	if err != nil {
		return err
	}
	if list.Agents == nil || len(list.Agents) != 0 {
		return fmt.Errorf("Agentenliste nicht leer: %+v", list.Agents)
	}
	return nil
}

func (s *Suite) listAgentOnce(org, name string) error {
	list, err := s.getAgentList(org)
	if err != nil {
		return err
	}
	count := 0
	for _, item := range list.Agents {
		if item.Name == name {
			count++
		}
	}
	if count != 1 {
		return fmt.Errorf("%s erscheint %d-mal: %+v", name, count, list.Agents)
	}
	return nil
}

func (s *Suite) foreignOriginPost(path, agent, template, origin string) error {
	path = strings.Replace(path, "{id}", url.PathEscape(s.orgIDs["Nordstern"]), 1)
	body, contentType := s.originBody(path, agent, template)
	return s.requestWith("POST", path, body, contentType, origin)
}

func (s *Suite) originBody(path, agent, template string) (string, string) {
	if strings.HasPrefix(path, "/api/") {
		body, _ := json.Marshal(map[string]any{"name": agent, "template_id": strings.ToLower(template), "execution_kind": "eino"})
		return string(body), "application/json"
	}
	values := url.Values{"name": {agent}, "template_id": {strings.ToLower(template)}, "execution_kind": {"eino"}}
	return values.Encode(), "application/x-www-form-urlencoded"
}

func (s *Suite) spoofedLocalHostPost() error {
	body, _ := json.Marshal(map[string]string{"name": "Gefaelscht", "template_id": "recherche", "execution_kind": "eino"})
	req, err := http.NewRequest("POST", s.baseURL()+s.orgPath("Nordstern"), strings.NewReader(string(body)))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Host = "localhost:" + strings.Split(s.address, ":")[1]
	response, err := s.client.Do(req)
	if err != nil {
		return err
	}
	return s.recordResponse(response)
}
