package modellfreigabe

import (
	"encoding/json"
	"fmt"
	"strings"
)

func (s *suite) browserCommand(input map[string]any) error {
	if s.browser.process == nil {
		if err := s.browser.start(); err != nil {
			return err
		}
	}
	if err := s.browser.command(input); err != nil {
		return err
	}
	s.browserHTML = append(s.browserHTML, s.browser.page.HTML)
	return nil
}

func (s *suite) openSettingsBrowser() error {
	if err := s.browserCommand(map[string]any{"navigate": s.app.URL + "/settings/modellanbieter/openrouter", "wait": map[string]string{"heading": "OpenRouter"}}); err != nil {
		return err
	}
	if err := s.browserCommand(map[string]any{"click": "Organisationen auswählen →", "wait": map[string]string{"heading": "Organisationsübersicht"}}); err != nil {
		return err
	}
	return s.browserOpenGrantPage("Nord")
}

func (s *suite) browserOpenGrantPage(org string) error {
	// Navigation bleibt im echten Chrome. Der freigegebene Link wird zuvor von Settings aus bedient.
	return s.browserCommand(map[string]any{"navigate": s.app.URL + s.pagePath(org), "wait": map[string]string{"heading": "OpenRouter für " + org}})
}

func (s *suite) browserReference() error {
	if !strings.Contains(s.browser.page.HTML, s.reference) || strings.Contains(s.browser.page.HTML, firstKey) {
		return fmt.Errorf("Browser zeigt Referenz nicht schlüsselfrei")
	}
	return nil
}

func (s *suite) browserOrganizationsNotGranted(first, second string) error {
	for _, org := range []string{first, second} {
		if err := s.browserOpenGrantPage(org); err != nil {
			return err
		}
		if err := s.browserCheckOrganization(false); err != nil {
			return err
		}
	}
	return s.browserOpenGrantPage(first)
}

func (s *suite) browserCheckOrganization(granted bool) error {
	return s.browserCheck(`document.querySelector('#organisation-heading')?.closest('section')?.textContent`, granted)
}

func (s *suite) browserCheckAgent(agent string, granted bool) error {
	name, _ := json.Marshal(agent)
	expression := `([...document.querySelectorAll('#agents-heading ~ .grant-row')].find(row => row.querySelector('strong')?.textContent?.trim() === ` + string(name) + `)?.textContent)`
	return s.browserCheck(expression, granted)
}

func (s *suite) browserCheck(expression string, granted bool) error {
	state := "Status: nicht freigegeben"
	if granted {
		state = "Status: freigegeben"
	}
	value, _ := json.Marshal(state)
	return s.browserCommand(map[string]any{"evaluate": `(() => {
		const text = (` + expression + ` || '').replace(/\s+/g, ' ');
		if (!text.includes(` + string(value) + `)) throw new Error('Freigabestatus fehlt: ' + text)
	})()`, "wait": map[string]string{"heading": s.browser.page.Heading}})
}

func (s *suite) browserGrant(name string) error {
	org := s.agentOrgs[name]
	button := name + " ausdrücklich freigeben"
	statusExpression := s.browserAgentStatusExpression(name, true)
	if org == "" {
		org = name
		button = "Organisation ausdrücklich freigeben"
		statusExpression = s.browserOrganizationStatusExpression(true)
	}
	if !strings.HasSuffix(s.browser.page.URL, s.pagePath(org)) {
		if err := s.browserOpenGrantPage(org); err != nil {
			return err
		}
	}
	return s.browserCommand(map[string]any{"click": button, "wait": map[string]string{"heading": "OpenRouter für " + org, "expression": statusExpression}})
}

func (s *suite) browserOrganizationStatusExpression(granted bool) string {
	status := "Status: nicht freigegeben"
	if granted {
		status = "Status: freigegeben"
	}
	quoted, _ := json.Marshal(status)
	return `((document.querySelector('#organisation-heading')?.closest('section')?.textContent || '').replace(/\s+/g, ' ').includes(` + string(quoted) + `))`
}

func (s *suite) browserAgentStatusExpression(agent string, granted bool) string {
	status := "Status: nicht freigegeben"
	if granted {
		status = "Status: freigegeben"
	}
	nameJSON, _ := json.Marshal(agent)
	statusJSON, _ := json.Marshal(status)
	return `((([...document.querySelectorAll('#agents-heading ~ .grant-row')].find(row => row.querySelector('strong')?.textContent?.trim() === ` + string(nameJSON) + `)?.textContent) || '').replace(/\s+/g, ' ').includes(` + string(statusJSON) + `))`
}

func (s *suite) browserOrganizationGranted(org, firstAgent, secondAgent string) error {
	if err := s.browserCheckOrganization(true); err != nil {
		return err
	}
	if err := s.browserCheckAgent(firstAgent, false); err != nil {
		return err
	}
	return s.browserCheckAgent(secondAgent, false)
}

func (s *suite) browserAgentGranted(agent, other string) error {
	if err := s.browserCheckAgent(agent, true); err != nil {
		return err
	}
	return s.browserCheckAgent(other, false)
}

func (s *suite) browserRevoke(name string) error {
	org := s.agentOrgs[name]
	button := "Freigabe für " + name + " widerrufen"
	statusExpression := s.browserAgentStatusExpression(name, false)
	if s.orgIDs[name] != "" {
		org = name
		button = "Organisationsfreigabe widerrufen"
		statusExpression = s.browserOrganizationStatusExpression(false)
	}
	if !strings.HasSuffix(s.browser.page.URL, s.pagePath(org)) {
		if err := s.browserOpenGrantPage(org); err != nil {
			return err
		}
	}
	return s.browserCommand(map[string]any{"click": button, "wait": map[string]string{"expression": statusExpression}})
}

func (s *suite) browserAgentRevoked(agent, other string) error {
	if err := s.browserCheckAgent(agent, false); err != nil {
		return err
	}
	return s.browserCheckAgent(other, true)
}

func (s *suite) browserOrganizationReason() error {
	if !strings.Contains(s.browser.page.Text, "Organisationsfreigabe fehlt; die Verbindung ist nicht nutzbar.") {
		return fmt.Errorf("fehlender Organisationsgrund in Browserseite")
	}
	return nil
}

func (s *suite) browserNoUsableConnection() error {
	for _, agent := range []string{"Mira", "Nora"} {
		status, err := s.agentStatus(agent)
		if err != nil {
			return err
		}
		if status.Allowed {
			return fmt.Errorf("%s ist nach Organisationsentzug nutzbar", agent)
		}
	}
	return nil
}

func (s *suite) browserOpenSettingsAndAgent(agent string) error {
	if err := s.browserCommand(map[string]any{"navigate": s.app.URL + "/settings/modellanbieter/openrouter", "wait": map[string]string{"heading": "OpenRouter"}, "responses": true}); err != nil {
		return err
	}
	if err := s.browserOpenGrantPage(s.agentOrgs[agent]); err != nil {
		return err
	}
	return s.browserCommand(map[string]any{"navigate": s.app.URL + "/organisationen/" + s.orgIDs[s.agentOrgs[agent]] + "/agenten/" + s.agentIDs[agent], "responses": true})
}

func (s *suite) browserPublicFields() error {
	if !strings.Contains(s.browser.page.HTML, "Mira") {
		return fmt.Errorf("Agentenansicht fehlt")
	}
	if !strings.Contains(s.browser.page.HTML, "openrouter-central") && !strings.Contains(s.browser.page.HTML, "Modell") {
		return fmt.Errorf("Agentenansicht zeigt keinen Modellkontext")
	}
	return nil
}

func (s *suite) browserNoSecrets() error {
	content := s.browser.page.HTML
	for _, html := range s.browserHTML {
		content += html
	}
	for _, response := range s.browser.replies {
		content += response.Body
	}
	for _, field := range s.browser.page.Fields {
		content += field.Value
	}
	for _, part := range []string{firstKey, secondKey, "SYNTHETIC-Alpha", "SYNTHETIC-Beta"} {
		if strings.Contains(content, part) {
			return fmt.Errorf("Schlüsselteil im Browser oder einer Antwort")
		}
	}
	return nil
}
