package main

import (
	"fmt"
	"os"
	"strings"

	"agentcontrolplane/app/internal/adapter/web"
	webprobe "agentcontrolplane/app/internal/adapter/web/modellpruefung"
	appprobe "agentcontrolplane/app/internal/app/modellpruefung"
	"agentcontrolplane/app/internal/app/modellverbindung"
)

func (b *Bootstrap) mountModelProbe(server *web.Server, flow *modellverbindung.Service) error {
	modelID := strings.TrimSpace(os.Getenv("APP_CODEX_MODEL"))
	status := appprobe.NewVerbindungspruefung(flow)
	handler, err := webprobe.NewHandler(flow, status, modelID, b.address(), nil)
	if err != nil {
		return fmt.Errorf("Codex model probe startup: %w", err)
	}

	server.Handle("POST /api/settings/modelle/codex/probe", handler.Handler())
	return nil
}
