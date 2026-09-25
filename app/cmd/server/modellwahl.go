package main

import (
	"fmt"
	"os"

	"agentcontrolplane/app/internal/adapter/sqlite"
	"agentcontrolplane/app/internal/adapter/web"
	webmodellwahl "agentcontrolplane/app/internal/adapter/web/modellwahl"
	appmodellwahl "agentcontrolplane/app/internal/app/modellwahl"
	apporganisation "agentcontrolplane/app/internal/app/organisation"
	portmodellwahl "agentcontrolplane/app/internal/port/modellwahl"
	"agentcontrolplane/ui/bridge"
)

// mountModelChoice is called after migration 14 and Story-24 access wiring.
func (b *Bootstrap) mountModelChoice(server *web.Server, db *sqlite.Database, ui *bridge.Bridge, access portmodellwahl.OpenRouterAccess) error {
	catalog, err := appmodellwahl.NewCatalogLoader().Load(os.Getenv("APP_MODEL_CATALOG_PATH"))
	if err != nil {
		return fmt.Errorf("Modellwahl startup: %w", err)
	}

	service := appmodellwahl.NewService(db, apporganisation.NewLocalIdentity(), db, access, catalog)
	handler := webmodellwahl.NewHandler(service, ui, b.address())
	server.Handle("/api/organisationen/{id}/agenten/{agentID}/modellwahl", handler)
	server.Handle("/organisationen/{id}/agenten/{agentID}/modellwahl", handler)
	return nil
}
