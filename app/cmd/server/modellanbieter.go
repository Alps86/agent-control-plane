package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	credentialsadapter "agentcontrolplane/app/internal/adapter/credentials"
	"agentcontrolplane/app/internal/adapter/model/codexauth"
	"agentcontrolplane/app/internal/adapter/sqlite"
	"agentcontrolplane/app/internal/adapter/web"
	webcodex "agentcontrolplane/app/internal/adapter/web/modellanbieter"
	webopenrouter "agentcontrolplane/app/internal/adapter/web/openrouter"
	"agentcontrolplane/app/internal/app/modellverbindung"
	"agentcontrolplane/app/internal/app/openrouterverbindung"
	"agentcontrolplane/ui/bridge"
)

func (b *Bootstrap) mountModelProviders(server *web.Server, db *sqlite.Database, ui *bridge.Bridge) error {
	secretPath, keyPath := b.credentialPaths()
	store, err := credentialsadapter.NewStore(secretPath, keyPath)
	if err != nil {
		return fmt.Errorf("credential startup: %w", err)
	}

	auth, err := codexauth.NewDirect(codexauth.Config{Issuer: "https://auth.openai.com", ClientID: codexauth.CodexCLIClientID})
	if err != nil {
		return fmt.Errorf("Codex provider startup: %w", err)
	}

	flow := modellverbindung.New(auth, store)
	if err := b.mountCodex(server, ui, flow); err != nil {
		return err
	}

	return b.mountOpenRouter(server, db, ui, store)
}

func (b *Bootstrap) credentialPaths() (string, string) {
	directory := filepath.Join(filepath.Dir(b.path()), ".agent-control-plane-secrets")
	secretPath := os.Getenv("APP_CREDENTIALS_PATH")
	if secretPath == "" {
		secretPath = filepath.Join(directory, "credentials.enc")
	}

	keyPath := os.Getenv("APP_CREDENTIAL_KEY_PATH")
	if keyPath == "" {
		keyPath = filepath.Join(directory, "master.key")
	}

	return secretPath, keyPath
}

func (b *Bootstrap) mountCodex(server *web.Server, ui *bridge.Bridge, flow *modellverbindung.Service) error {
	handler, err := webcodex.New(flow, b.address())
	if err != nil {
		return fmt.Errorf("Codex settings startup: %w", err)
	}

	server.Handle("/settings/modelle/codex/device/", handler.Handler())
	server.Handle("GET /settings/modelle/codex", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b.codexPage(w, r, ui, flow)
	}))
	return b.mountModelProbe(server, flow)
}

func (b *Bootstrap) mountOpenRouter(server *web.Server, db *sqlite.Database, ui *bridge.Bridge, store *credentialsadapter.Store) error {
	probe, err := b.openRouterProbe()
	if err != nil {
		return err
	}
	service := openrouterverbindung.NewService(store, probe)
	handler := webopenrouter.NewHandler(service, ui, b.address())
	server.Handle("/settings/modellanbieter/openrouter", handler)
	server.Handle("/settings/modellanbieter/openrouter/", handler)
	server.Handle("/api/settings/modellanbieter/openrouter", handler)
	server.Handle("/api/settings/modellanbieter/openrouter/", handler)
	grants := b.mountModelGrants(server, db, ui, service)
	return b.mountModelChoice(server, db, ui, grants)
}

func (b *Bootstrap) codexPage(w http.ResponseWriter, r *http.Request, ui *bridge.Bridge, flow *modellverbindung.Service) {
	if !b.localSettingsRequest(r) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	status, err := flow.Status(r.Context())
	code := http.StatusOK
	if err != nil {
		status = modellverbindung.Status{State: "unavailable", Reason: "transport_unavailable"}
		code = http.StatusBadGateway
	}

	b.renderPage(w, ui, b.codexTemplate(r), b.codexData(status), code)
}

func (b *Bootstrap) codexTemplate(r *http.Request) string {
	if strings.EqualFold(r.Header.Get("HX-Request"), "true") {
		return "modelle/codex/content"
	}

	return "modelle/codex/page"
}

func (b *Bootstrap) codexData(status modellverbindung.Status) map[string]any {
	return map[string]any{"state": status.State, "reason": status.Reason, "verificationUrl": status.VerificationURL, "userCode": status.UserCode}
}
