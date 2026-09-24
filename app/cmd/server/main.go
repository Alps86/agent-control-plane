package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"agentcontrolplane/app/internal/adapter/sqlite"
	"agentcontrolplane/app/internal/adapter/web"
	"agentcontrolplane/app/internal/adapter/web/experimente"
	weborganisation "agentcontrolplane/app/internal/adapter/web/organisation"
	apporganisation "agentcontrolplane/app/internal/app/organisation"
	"agentcontrolplane/app/internal/app/system"
	"agentcontrolplane/ui/bridge"
)

func main() {
	if err := (&Bootstrap{}).Run(); err != nil {
		log.Fatal(err)
	}
}

func (b *Bootstrap) Run() error {
	db, err := sqlite.OpenWithMigrations(context.Background(), b.path(), sqlite.RunMigration(2), sqlite.OrganizationMigration(), sqlite.GoalMigration(), sqlite.AgentMigration())
	if err != nil {
		return fmt.Errorf("database startup: %w", err)
	}

	defer db.Close()
	ui, err := bridge.New()
	if err != nil {
		return fmt.Errorf("UI startup: %w", err)
	}

	service := apporganisation.NewService(sqlite.NewOrganizationStore(db), apporganisation.NewLocalIdentity())
	organizations := weborganisation.NewHandler(service, ui)
	server := web.NewServer(system.NewProbe(), db)
	b.mount(server, organizations, ui)
	b.mountGoals(server, db, service, ui)
	b.mountAgents(server, db, ui)
	return http.ListenAndServe(b.address(), server.Handler())
}

func (b *Bootstrap) mount(server *web.Server, organizations *weborganisation.Handler, ui *bridge.Bridge) {
	for _, pattern := range []string{"/api/organisationen", "/api/organisationen/", "/organisationen", "/organisationen/"} {
		server.Handle(pattern, organizations)
	}

	server.Handle("GET /experimente", experimente.NewHandler(ui))
	server.Handle("GET /{$}", http.RedirectHandler("/organisationen", http.StatusSeeOther))
	server.Handle("/assets/", ui.Assets())
	server.Handle("/fragments/", ui.Assets())
}

func (b *Bootstrap) address() string {
	address := os.Getenv("APP_ADDR")
	if address == "" {
		return "127.0.0.1:8080"
	}

	return address
}

func (b *Bootstrap) path() string {
	path := os.Getenv("APP_DB_PATH")
	if path == "" {
		return "agent-control-plane.db"
	}

	return path
}
