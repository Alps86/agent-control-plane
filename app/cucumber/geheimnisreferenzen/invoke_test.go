package geheimnisreferenzen

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	domainfreigabe "agentcontrolplane/app/internal/domain/modellfreigabe"
	grantport "agentcontrolplane/app/internal/port/modellfreigabe"
)

func (s *Suite) setupInvoke() error {
	if err := s.connectRouter(); err != nil {
		return err
	}
	if err := s.createOrganization(); err != nil {
		return err
	}
	if err := s.createAgent(); err != nil {
		return err
	}
	return s.grantAgent()
}

func (s *Suite) createOrganization() error {
	if err := s.request(http.MethodPost, "/api/organisationen", map[string]string{"name": "Nord"}); err != nil {
		return err
	}
	return s.createdID(&s.orgID)
}

func (s *Suite) createAgent() error {
	path := "/api/organisationen/" + s.orgID + "/agenten"
	if err := s.request(http.MethodPost, path, map[string]string{"name": "Mira", "template_id": "recherche", "execution_kind": "eino"}); err != nil {
		return err
	}
	return s.createdID(&s.agentID)
}

func (s *Suite) createdID(target *string) error {
	if s.status != http.StatusCreated {
		return fmt.Errorf("Anlage HTTP %d: %s", s.status, s.last)
	}
	var result struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(s.last, &result); err != nil {
		return err
	}
	if result.ID == "" {
		return fmt.Errorf("öffentliche Anlage ohne ID")
	}
	*target = result.ID
	return nil
}

func (s *Suite) grantAgent() error {
	base := "/api/organisationen/" + s.orgID + "/modellfreigabe/openrouter"
	if err := s.request(http.MethodPost, base+"/organisation", nil); err != nil {
		return err
	}
	if s.status != http.StatusOK {
		return fmt.Errorf("Organisationsfreigabe HTTP %d: %s", s.status, s.last)
	}
	if err := s.request(http.MethodPost, base+"/agenten/"+s.agentID, nil); err != nil {
		return err
	}
	if s.status != http.StatusOK {
		return fmt.Errorf("Agentenfreigabe HTTP %d: %s", s.status, s.last)
	}
	return nil
}

func (s *Suite) invoke() error {
	s.invokeErr = s.grants.Invoke(context.Background(), s.orgID, s.agentID, "openrouter-central", Caller{s})
	return nil
}

func (c Caller) Call(ctx context.Context, _, _ string, access grantport.Access) error {
	key, err := access.Resolve(ctx)
	if err != nil {
		return err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, c.suite.provider.URL+"/api/v1/chat/completions", strings.NewReader(`{"model":"synthetic","messages":[]}`))
	if err != nil {
		return err
	}
	request.Header.Set("Authorization", "Bearer "+key)
	response, err := c.suite.provider.Client().Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("kontrollierter Anbieter HTTP %d", response.StatusCode)
	}
	return nil
}

func (s *Suite) oneCall() error {
	if count := s.providerCalls.Load(); count != 1 {
		return fmt.Errorf("Anbieter erhielt %d statt einen Request", count)
	}
	return nil
}

func (s *Suite) invokeDenied() error {
	if s.invokeErr == nil || s.invokeErr.Error() != domainfreigabe.ReasonConnection {
		return fmt.Errorf("Verbindungsfehler fehlt: %v", s.invokeErr)
	}
	return s.oneCall()
}

func (s *Suite) noInvokeSecrets() error {
	if s.invokeErr == nil {
		return fmt.Errorf("Verbindungsfehler fehlt")
	}
	if strings.Contains(s.invokeErr.Error(), testKey) || strings.Contains(s.invokeErr.Error(), testAccess) {
		return fmt.Errorf("Verbindungsfehler enthält Secret")
	}
	return nil
}
