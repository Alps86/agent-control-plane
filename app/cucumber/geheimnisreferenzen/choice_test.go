package geheimnisreferenzen

import (
	"context"
	"fmt"
	"net/http"

	domainfreigabe "agentcontrolplane/app/internal/domain/modellfreigabe"
	domainchoice "agentcontrolplane/app/internal/domain/modellwahl"
	portchoice "agentcontrolplane/app/internal/port/modellwahl"
)

func (s ChoiceStore) Get(ctx context.Context, org, agent, operator string) (domainchoice.Selection, error) {
	return s.db.GetModelChoice(ctx, org, agent, operator)
}

func (s ChoiceStore) Replace(ctx context.Context, org, agent, operator string, selection domainchoice.Selection) error {
	return s.db.Replace(ctx, org, agent, operator, selection)
}

func (g ChoiceGrant) Allowed(ctx context.Context, org, agent, reference string) error {
	decision, err := g.grants.Authorize(ctx, org, agent, reference)
	if err != nil {
		return err
	}
	if decision.Allowed {
		return nil
	}
	if decision.Reason == domainfreigabe.ReasonOrganization {
		return portchoice.ErrOrganizationGrant
	}
	if decision.Reason == domainfreigabe.ReasonAgent {
		return portchoice.ErrAgentGrant
	}
	return portchoice.ErrConnectionUnavailable
}

func (s *Suite) noFallbackValue() error {
	path := "/api/organisationen/" + s.orgID + "/agenten/" + s.agentID + "/modellwahl"
	if err := s.request(http.MethodGet, path, nil); err != nil {
		return err
	}
	if s.status != http.StatusOK {
		return fmt.Errorf("Modellwahl HTTP %d: %s", s.status, s.last)
	}
	return s.noStatusSecrets()
}
