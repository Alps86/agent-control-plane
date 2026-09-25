package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"

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
	if err := b.configureAddress(); err != nil {
		return err
	}

	db, err := b.openDatabase()
	if err != nil {
		return err
	}

	defer db.Close()
	return b.serve(db)
}

func (b *Bootstrap) openDatabase() (*sqlite.Database, error) {
	db, err := sqlite.OpenApplication(context.Background(), b.path())
	if err != nil {
		return nil, fmt.Errorf("database startup: %w", err)
	}

	return db, nil
}

func (b *Bootstrap) serve(db *sqlite.Database) error {
	ui, err := bridge.New()
	if err != nil {
		return fmt.Errorf("UI startup: %w", err)
	}

	server, err := b.assemble(db, ui)
	if err != nil {
		return err
	}

	return http.ListenAndServe(b.address(), server.Handler())
}

func (b *Bootstrap) assemble(db *sqlite.Database, ui *bridge.Bridge) (*web.Server, error) {
	service := apporganisation.NewService(sqlite.NewOrganizationStore(db), apporganisation.NewLocalIdentity())
	organizations := weborganisation.NewHandler(service, ui, b.address())
	server := web.NewServer(system.NewProbe(), db)
	b.mount(server, organizations, ui)
	b.mountGoals(server, db, service, ui)
	b.mountProjects(server, db, service, ui)
	b.mountAgents(server, db, ui)
	b.mountDataScope(server, db, ui)
	b.mountCodexProfile(server, db, ui)
	b.mountSettings(server, ui)
	if err := b.mountModelProviders(server, ui); err != nil {
		return nil, err
	}

	return server, nil
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
	return b.bindAddress
}

func (b *Bootstrap) configureAddress() error {
	address := os.Getenv("APP_ADDR")
	if address == "" {
		address = "127.0.0.1:8080"
	}

	host, port, err := net.SplitHostPort(address)
	if err != nil || !b.validAddressPort(port) {
		return fmt.Errorf("APP_ADDR must be a local loopback address with a valid port")
	}

	host = b.normalizedHost(host)
	ip := net.ParseIP(host)
	if ip == nil || !ip.IsLoopback() {
		return fmt.Errorf("APP_ADDR must be a local loopback address")
	}

	b.bindAddress = net.JoinHostPort(host, port)
	return nil
}

func (b *Bootstrap) validAddressPort(raw string) bool {
	number, err := strconv.Atoi(raw)
	return err == nil && number > 0 && number <= 65535
}

func (b *Bootstrap) normalizedHost(host string) string {
	if strings.EqualFold(host, "localhost") {
		return "127.0.0.1"
	}

	return strings.ToLower(host)
}

func (b *Bootstrap) path() string {
	path := os.Getenv("APP_DB_PATH")
	if path == "" {
		return "agent-control-plane.db"
	}

	return path
}
