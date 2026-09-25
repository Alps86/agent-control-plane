package main

import (
	"context"
	"fmt"
	"os"

	"agentcontrolplane/app/internal/adapter/sqlite"
	"agentcontrolplane/app/internal/adapter/web"
	webmodellwahl "agentcontrolplane/app/internal/adapter/web/modellwahl"
	appfreigabe "agentcontrolplane/app/internal/app/modellfreigabe"
	appmodellwahl "agentcontrolplane/app/internal/app/modellwahl"
	apporganisation "agentcontrolplane/app/internal/app/organisation"
	domainfreigabe "agentcontrolplane/app/internal/domain/modellfreigabe"
	domainmodellwahl "agentcontrolplane/app/internal/domain/modellwahl"
	portmodellwahl "agentcontrolplane/app/internal/port/modellwahl"
	"agentcontrolplane/ui/bridge"
)

// mountModelChoice is called after migration 14 and Story-24 access wiring.
func (b *Bootstrap) mountModelChoice(server *web.Server, db *sqlite.Database, ui *bridge.Bridge, grants *appfreigabe.Service) (domainmodellwahl.Catalog, error) {
	catalog, err := appmodellwahl.NewCatalogLoader().Load(os.Getenv("APP_MODEL_CATALOG_PATH"))
	if err != nil {
		return domainmodellwahl.Catalog{}, fmt.Errorf("Modellwahl startup: %w", err)
	}

	service := appmodellwahl.NewService(db, apporganisation.NewLocalIdentity(), modelChoiceStore{db}, modelGrantAccess{grants}, catalog)
	handler := webmodellwahl.NewHandler(service, ui, b.address())
	server.Handle("/api/organisationen/{id}/agenten/{agentID}/modellwahl", handler)
	server.Handle("/organisationen/{id}/agenten/{agentID}/modellwahl", handler)
	return catalog, nil
}

func (s modelChoiceStore) Get(ctx context.Context, organizationID, agentID, operatorID string) (domainmodellwahl.Selection, error) {
	return s.db.GetModelChoice(ctx, organizationID, agentID, operatorID)
}

func (s modelChoiceStore) Replace(ctx context.Context, organizationID, agentID, operatorID string, selection domainmodellwahl.Selection) error {
	return s.db.Replace(ctx, organizationID, agentID, operatorID, selection)
}

func (a modelGrantAccess) Allowed(ctx context.Context, organizationID, agentID, reference string) error {
	decision, err := a.grants.Authorize(ctx, organizationID, agentID, reference)
	if err != nil {
		return err
	}

	if decision.Allowed {
		return nil
	}

	if decision.Reason == domainfreigabe.ReasonOrganization {
		return portmodellwahl.ErrOrganizationGrant
	}

	if decision.Reason == domainfreigabe.ReasonAgent {
		return portmodellwahl.ErrAgentGrant
	}

	return portmodellwahl.ErrConnectionUnavailable
}
