package modellfreigabe

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	appfreigabe "agentcontrolplane/app/internal/app/modellfreigabe"
	domainfreigabe "agentcontrolplane/app/internal/domain/modellfreigabe"
	portfreigabe "agentcontrolplane/app/internal/port/modellfreigabe"
)

type controlledCaller struct{ suite *suite }

func (c controlledCaller) Call(ctx context.Context, _, _ string, access portfreigabe.Access) error {
	if c.suite.callerEntered != nil {
		release := c.suite.callerRelease
		c.suite.callerEntered <- struct{}{}
		select {
		case <-release:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	key, err := access.Resolve(ctx)
	if err != nil {
		return err
	}
	if key != firstKey && key != secondKey {
		return fmt.Errorf("kontrollierter Anbieter erhielt unbekannten Schlüssel")
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, c.suite.provider.URL+"/api/v1/chat/completions", strings.NewReader(`{"model":"synthetic-test-model","messages":[{"role":"user","content":"ping"}]}`))
	if err != nil {
		return err
	}
	request.Header.Set("Authorization", "Bearer "+key)
	request.Header.Set("Content-Type", "application/json")
	response, err := c.suite.client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(response.Body, 1024))
		return fmt.Errorf("kontrollierter Modellanbieter: HTTP %d: %s", response.StatusCode, body)
	}
	return nil
}

func (s *suite) organizations(first, second string) error {
	if err := s.createOrganization(first); err != nil {
		return err
	}
	return s.createOrganization(second)
}

func (s *suite) agents(first, second, org, third, otherOrg string) error {
	for _, pair := range [][2]string{{first, org}, {second, org}, {third, otherOrg}} {
		if err := s.createAgent(pair[0], pair[1]); err != nil {
			return err
		}
	}
	return nil
}

func (s *suite) browserEntities(firstOrg, secondOrg, firstAgent, secondAgent, agentOrg string) error {
	if err := s.organizations(firstOrg, secondOrg); err != nil {
		return err
	}
	if err := s.createAgent(firstAgent, agentOrg); err != nil {
		return err
	}
	return s.createAgent(secondAgent, agentOrg)
}

func (s *suite) providerReady() error {
	if s.provider == nil || s.service == nil || s.reference == "" {
		return fmt.Errorf("kontrollierter Modellzugang fehlt")
	}
	return nil
}

func (s *suite) apiPath(org string) string {
	return "/api/organisationen/" + url.PathEscape(s.orgIDs[org]) + "/modellfreigabe/openrouter"
}

func (s *suite) pagePath(org string) string {
	return "/organisationen/" + url.PathEscape(s.orgIDs[org]) + "/modellfreigabe/openrouter"
}

func (s *suite) grantOrganization(org string) error {
	if err := s.request(http.MethodPost, s.apiPath(org)+"/organisation", nil); err != nil {
		return err
	}
	return s.expectStatus(http.StatusOK)
}

func (s *suite) grantAgent(agent string) error {
	org := s.agentOrgs[agent]
	if err := s.request(http.MethodPost, s.apiPath(org)+"/agenten/"+url.PathEscape(s.agentIDs[agent]), nil); err != nil {
		return err
	}
	return s.expectStatus(http.StatusOK)
}

func (s *suite) grantedLevel(org, level string) error {
	if level == "keiner" {
		return nil
	}
	if level == "Organisation allein" {
		return s.grantOrganization(org)
	}
	return fmt.Errorf("unbekannte Freigabeebene %q", level)
}

func (s *suite) grantedOrganizationAgent(org, agent string) error {
	if err := s.grantOrganization(org); err != nil {
		return err
	}
	return s.grantAgent(agent)
}

func (s *suite) grantedOrganizationAgents(org, first, second string) error {
	if err := s.grantedOrganizationAgent(org, first); err != nil {
		return err
	}
	return s.grantAgent(second)
}

func (s *suite) revoke(name string) error {
	if id := s.orgIDs[name]; id != "" {
		if err := s.request(http.MethodDelete, s.apiPath(name)+"/organisation", nil); err != nil {
			return err
		}
		return s.expectStatus(http.StatusOK)
	}
	org := s.agentOrgs[name]
	if org == "" {
		return fmt.Errorf("unbekanntes Freigabeziel %q", name)
	}
	if err := s.request(http.MethodDelete, s.apiPath(org)+"/agenten/"+url.PathEscape(s.agentIDs[name]), nil); err != nil {
		return err
	}
	return s.expectStatus(http.StatusOK)
}

func (s *suite) expectStatus(code int) error {
	if s.status != code {
		return fmt.Errorf("HTTP %d statt %d: %s", s.status, code, s.body)
	}
	return nil
}

func (s *suite) overview(org string) (appfreigabe.Overview, error) {
	if err := s.request(http.MethodGet, s.apiPath(org), nil); err != nil {
		return appfreigabe.Overview{}, err
	}
	if err := s.expectStatus(http.StatusOK); err != nil {
		return appfreigabe.Overview{}, err
	}
	var view appfreigabe.Overview
	return view, json.Unmarshal(s.body, &view)
}

func (s *suite) agentStatus(agent string) (domainfreigabe.Status, error) {
	org := s.agentOrgs[agent]
	path := s.apiPath(org) + "/agenten/" + url.PathEscape(s.agentIDs[agent])
	if err := s.request(http.MethodGet, path, nil); err != nil {
		return domainfreigabe.Status{}, err
	}
	if err := s.expectStatus(http.StatusOK); err != nil {
		return domainfreigabe.Status{}, err
	}
	var status domainfreigabe.Status
	return status, json.Unmarshal(s.body, &status)
}

func (s *suite) authorizeAgent(agent string) error {
	decision, err := s.service.Authorize(context.Background(), s.orgIDs[s.agentOrgs[agent]], s.agentIDs[agent], s.reference)
	if err != nil {
		return err
	}
	s.lastDecision = decision.Reason
	if decision.Allowed {
		s.lastDecision = "allowed"
	}
	return nil
}

func (s *suite) invokeAgent(agent string) error {
	s.lastError = s.service.Invoke(context.Background(), s.orgIDs[s.agentOrgs[agent]], s.agentIDs[agent], s.reference, controlledCaller{s})
	return nil
}

func (s *suite) crossAgentAttempts(first, firstOrg, second, secondOrg string) error {
	if s.agentOrgs[first] != firstOrg || s.agentOrgs[second] != secondOrg {
		return fmt.Errorf("falscher Agentenkontext")
	}
	for _, agent := range []string{first, second} {
		if err := s.invokeAgent(agent); err != nil {
			return err
		}
		if s.lastError == nil {
			return fmt.Errorf("%s erreichte trotz fehlender Freigabe den Anbieter", agent)
		}
	}
	return nil
}

func (s *suite) authorizeThenWait(agent string) error {
	s.waitingAgent = agent
	return s.startHeldInvocation(agent)
}

func (s *suite) resumeInvocation() error { return s.releaseAuthorizedBinding() }

func (s *suite) createAdditionalOrganization() error {
	if err := s.createOrganization("West"); err != nil {
		return err
	}
	return s.createAgent("Wera", "West")
}

func (s *suite) createAdditionalAgent(org string) error { return s.createAgent("Nela", org) }

func (s *suite) delegatedAttempt(delegator, executor string) error {
	if s.agentIDs[delegator] == "" || s.agentIDs[executor] == "" {
		return fmt.Errorf("delegierter Agent fehlt")
	}
	return s.invokeAgent(executor)
}

func (s *suite) rotateKey() error {
	if err := s.request(http.MethodPost, "/api/settings/modellanbieter/openrouter", map[string]string{"key": secondKey}); err != nil {
		return err
	}
	if err := s.expectStatus(http.StatusOK); err != nil {
		return err
	}
	if err := s.request(http.MethodPost, "/api/settings/modellanbieter/openrouter/pruefen", nil); err != nil {
		return err
	}
	return s.expectStatus(http.StatusOK)
}

func (s *suite) fetchViews() error {
	if err := s.request(http.MethodGet, s.pagePath("Nord"), nil); err != nil {
		return err
	}
	if err := s.expectStatus(http.StatusOK); err != nil {
		return err
	}
	s.pageHTML = append([]byte(nil), s.body...)
	if _, err := s.overview("Nord"); err != nil {
		return err
	}
	s.jsonBody = append([]byte(nil), s.body...)
	_, err := s.agentStatus("Mira")
	if err == nil {
		s.jsonBody = append(s.jsonBody, s.body...)
	}
	return err
}

func (s *suite) organizationGrantedAgentNot(org, agent string) error {
	view, err := s.overview(org)
	if err != nil {
		return err
	}
	if !view.Status.OrganizationGranted {
		return fmt.Errorf("Organisationsfreigabe fehlt")
	}
	for _, item := range view.Agents {
		if item.Name == agent && !item.Status.AgentGranted {
			return nil
		}
	}
	return fmt.Errorf("Agent %s nicht explizit ungefreigegeben: %+v", agent, view.Agents)
}

func (s *suite) agentDenied(agent string) error {
	before := s.modelCalls()
	if err := s.invokeAgent(agent); err != nil {
		return err
	}
	if s.lastError == nil || !strings.Contains(s.lastError.Error(), domainfreigabe.ReasonAgent) || s.modelCalls() != before {
		return fmt.Errorf("Agentenfreigabe nicht vor Modellaufruf erzwungen: err=%v, calls=%d", s.lastError, s.modelCalls()-before)
	}
	return nil
}

func (s *suite) agentAllowed(agent string) error {
	decision, err := s.service.Authorize(context.Background(), s.orgIDs[s.agentOrgs[agent]], s.agentIDs[agent], s.reference)
	if err != nil {
		return err
	}
	if !decision.Allowed {
		return fmt.Errorf("Agent %s unzulässig: %s", agent, decision.Reason)
	}
	return nil
}

func (s *suite) agentCalledOnce(agent string) error {
	if err := s.invokeAgent(agent); err != nil {
		return err
	}
	if s.lastError != nil || s.modelCalls() != 1 {
		return fmt.Errorf("kontrollierter Modellzugang: calls=%d err=%v", s.modelCalls(), s.lastError)
	}
	return nil
}

func (s *suite) agentNotCalled(agent string) error {
	before := s.modelCalls()
	if err := s.invokeAgent(agent); err != nil {
		return err
	}
	if s.lastError == nil || s.modelCalls() != before {
		return fmt.Errorf("Agent %s erreichte Modellzugang: calls=%d err=%v", agent, s.modelCalls()-before, s.lastError)
	}
	return nil
}

func (s *suite) twoAgentsDenied(first, second string) error {
	if err := s.agentNotCalled(first); err != nil {
		return err
	}
	return s.agentNotCalled(second)
}

func (s *suite) decisionDeniedForReason(reason string) error {
	if s.lastDecision != reason {
		return fmt.Errorf("Zuordnung: %q statt %q", s.lastDecision, reason)
	}
	return nil
}

func (s *suite) bothDenied() error {
	if s.modelCalls() != 0 {
		return fmt.Errorf("kontrollierter Aufruf trotz Ablehnung: %d", s.modelCalls())
	}
	return nil
}

func (s *suite) noForeignData() error {
	text := string(s.body) + fmt.Sprint(s.lastError)
	for _, forbidden := range []string{firstKey, secondKey, s.orgIDs["Süd"], "Süd"} {
		if forbidden != "" && strings.Contains(text, forbidden) {
			return fmt.Errorf("fremde Daten in öffentlicher Antwort")
		}
	}
	return nil
}

func (s *suite) agentStatusNotGranted(agent string) error {
	status, err := s.agentStatus(agent)
	if err != nil {
		return err
	}
	if status.AgentGranted {
		return fmt.Errorf("Agent %s bleibt freigegeben", agent)
	}
	return nil
}

func (s *suite) agentCalled(agent string) error {
	before := s.modelCalls()
	if err := s.invokeAgent(agent); err != nil {
		return err
	}
	if s.lastError != nil || s.modelCalls() != before+1 {
		return fmt.Errorf("Agent %s: calls=%d err=%v", agent, s.modelCalls()-before, s.lastError)
	}
	return nil
}

func (s *suite) organizationReason() error {
	status, err := s.agentStatus("Mira")
	if err != nil {
		return err
	}
	if status.Reason != domainfreigabe.ReasonOrganization {
		return fmt.Errorf("falscher Organisationsgrund: %q", status.Reason)
	}
	return nil
}

func (s *suite) invocationDeniedForReason(reason string) error {
	if s.lastError == nil || s.lastError.Error() != reason {
		return fmt.Errorf("Aufrufgrund %v statt %q", s.lastError, reason)
	}
	return nil
}

func (s *suite) noCalls() error {
	if s.modelCalls() != 0 {
		return fmt.Errorf("kontrollierter Anbieter wurde %d-mal aufgerufen", s.modelCalls())
	}
	return nil
}

func (s *suite) newEntitiesNotGranted() error {
	for _, agent := range []string{"Wera", "Nela"} {
		status, err := s.agentStatus(agent)
		if err != nil {
			return err
		}
		if status.AgentGranted || status.Allowed {
			return fmt.Errorf("neuer Agent %s erbte Freigabe", agent)
		}
	}
	view, err := s.overview("West")
	if err != nil {
		return err
	}
	if view.Status.OrganizationGranted {
		return fmt.Errorf("neue Organisation erbte Freigabe")
	}
	return nil
}

func (s *suite) newEntitiesDenied() error { return s.twoAgentsDenied("Wera", "Nela") }

func (s *suite) delegatedDenied(agent string) error {
	if s.lastError == nil || !strings.Contains(s.lastError.Error(), domainfreigabe.ReasonAgent) || s.modelCalls() != 0 {
		return fmt.Errorf("delegierter Agent %s: %v, calls=%d", agent, s.lastError, s.modelCalls())
	}
	return nil
}

func (s *suite) noInheritedGrant(_, executor string) error {
	status, err := s.agentStatus(executor)
	if err != nil {
		return err
	}
	if status.AgentGranted {
		return fmt.Errorf("delegierter Agent erbte Freigabe")
	}
	return nil
}

func (s *suite) onlyOriginalGrants(org, agent string) error {
	view, err := s.overview(org)
	if err != nil {
		return err
	}
	if !view.Status.OrganizationGranted {
		return fmt.Errorf("Organisationsfreigabe nach Rotation verloren")
	}
	for _, item := range view.Agents {
		if item.Status.AgentGranted != (item.Name == agent) {
			return fmt.Errorf("Rotation änderte Agentenfreigabe: %+v", item)
		}
	}
	other, err := s.overview("Süd")
	if err != nil {
		return err
	}
	if other.Status.OrganizationGranted {
		return fmt.Errorf("Rotation gab fremde Organisation frei")
	}
	return nil
}

func (s *suite) publicViewFields() error {
	if !strings.Contains(string(s.pageHTML), s.reference) || !strings.Contains(string(s.jsonBody), s.reference) ||
		!strings.Contains(string(s.jsonBody), `"organization_granted"`) || !strings.Contains(string(s.jsonBody), `"agent_granted"`) {
		return fmt.Errorf("Referenz/Freigabestatus fehlt in öffentlichen Ansichten")
	}
	return nil
}

func (s *suite) noSecretsInViews() error {
	for _, body := range [][]byte{s.pageHTML, s.jsonBody} {
		for _, fragment := range []string{firstKey, secondKey, "SYNTHETIC-Alpha", "SYNTHETIC-Beta"} {
			if strings.Contains(string(body), fragment) {
				return fmt.Errorf("Schlüsselteil in öffentlicher Antwort")
			}
		}
	}
	return nil
}

func (s *suite) noSecretsInAgentOrError() error {
	if err := s.request(http.MethodGet, "/api/organisationen/"+url.PathEscape(s.orgIDs["Nord"])+"/agenten/"+url.PathEscape(s.agentIDs["Mira"]), nil); err != nil {
		return err
	}
	if err := s.expectStatus(http.StatusOK); err != nil {
		return err
	}
	agentBody := append([]byte(nil), s.body...)
	if err := s.agentNotCalled("Nora"); err != nil {
		return err
	}
	for _, part := range []string{firstKey, secondKey, "SYNTHETIC-Alpha", "SYNTHETIC-Beta"} {
		if strings.Contains(string(agentBody)+fmt.Sprint(s.lastError), part) {
			return fmt.Errorf("Schlüsselteil in Agent oder Fehler")
		}
	}
	return nil
}

func (s *suite) disconnectSettings() error {
	if err := s.request(http.MethodDelete, "/api/settings/modellanbieter/openrouter", nil); err != nil {
		return err
	}
	return s.expectStatus(http.StatusOK)
}

func (s *suite) reconnectSettings() error {
	if err := s.request(http.MethodPost, "/api/settings/modellanbieter/openrouter", map[string]string{"key": secondKey}); err != nil {
		return err
	}
	if err := s.expectStatus(http.StatusOK); err != nil {
		return err
	}
	if err := s.request(http.MethodPost, "/api/settings/modellanbieter/openrouter/pruefen", nil); err != nil {
		return err
	}
	return s.expectStatus(http.StatusOK)
}

func (s *suite) reconnectedReference() error {
	if err := s.request(http.MethodGet, "/api/settings/modellanbieter/openrouter", nil); err != nil {
		return err
	}
	if err := s.expectStatus(http.StatusOK); err != nil {
		return err
	}
	var status struct {
		Reference string `json:"reference"`
	}
	if err := json.Unmarshal(s.body, &status); err != nil {
		return err
	}
	if status.Reference != s.reference || strings.Contains(string(s.body), secondKey) {
		return fmt.Errorf("erneute Verbindung verriet Schlüssel oder wechselte Referenz")
	}
	return nil
}

func (s *suite) oldGrantsInvalid(org, agent string) error {
	view, err := s.overview(org)
	if err != nil {
		return err
	}
	if view.Status.OrganizationGranted {
		return fmt.Errorf("alte Organisationsfreigabe nach Reconnect wirksam")
	}
	status, err := s.agentStatus(agent)
	if err != nil {
		return err
	}
	if status.AgentGranted || status.Allowed {
		return fmt.Errorf("alte Agentenfreigabe nach Reconnect wirksam")
	}
	return nil
}

func (s *suite) regrantAfterReconnect(org, agent string) error {
	if err := s.grantOrganization(org); err != nil {
		return err
	}
	if err := s.grantAgent(agent); err != nil {
		return err
	}
	s.reconnectBaseline = s.modelCalls()
	return nil
}

func (s *suite) reconnectedCallOnce(agent string) error {
	if err := s.agentCalled(agent); err != nil {
		return err
	}
	if s.modelCalls() != s.reconnectBaseline+1 {
		return fmt.Errorf("nach Reconnect %d statt genau einem neuen Aufruf", s.modelCalls()-s.reconnectBaseline)
	}
	return nil
}

func (s *suite) reconnectNoSecrets() error {
	if err := s.fetchViews(); err != nil {
		return err
	}
	if err := s.noSecretsInViews(); err != nil {
		return err
	}
	return s.reconnectedReference()
}

func (s *suite) positivelyAuthorized(agent string) error { return s.agentAllowed(agent) }

func (s *suite) holdAuthorizedBinding() error { return s.startHeldInvocation("Mira") }

func (s *suite) startHeldInvocation(agent string) error {
	s.callerEntered = make(chan struct{}, 1)
	s.callerRelease = make(chan struct{})
	s.invokeDone = make(chan error, 1)
	go func() {
		s.invokeDone <- s.service.Invoke(context.Background(), s.orgIDs[s.agentOrgs[agent]], s.agentIDs[agent], s.reference, controlledCaller{s})
	}()
	select {
	case <-s.callerEntered:
		return nil
	case err := <-s.invokeDone:
		return fmt.Errorf("Aufruf erreichte Caller nicht: %v", err)
	case <-time.After(3 * time.Second):
		return fmt.Errorf("autorisierter Caller wurde nicht erreicht")
	}
}

func (s *suite) reconnectSameReference() error {
	if err := s.reconnectSettings(); err != nil {
		return err
	}
	return s.reconnectedReference()
}

func (s *suite) releaseAuthorizedBinding() error {
	close(s.callerRelease)
	s.callerRelease = nil
	select {
	case s.lastError = <-s.invokeDone:
		return nil
	case <-time.After(3 * time.Second):
		return fmt.Errorf("Resolver für alte Bindung blockiert")
	}
}

func (s *suite) oldBindingRevoked() error {
	if s.lastError == nil || s.lastError.Error() != domainfreigabe.ReasonConnection {
		return fmt.Errorf("alte Bindung nicht widerrufen: %v", s.lastError)
	}
	return nil
}

func (s *suite) noCallFromBinding() error { return s.noCalls() }
