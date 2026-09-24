package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"agentcontrolplane/app/internal/adapter/sqlite"
	"agentcontrolplane/app/internal/adapter/web"
	"agentcontrolplane/app/internal/app/system"
)

func main() {
	address := os.Getenv("APP_ADDR")
	if address == "" {
		address = "127.0.0.1:8080"
	}

	path := os.Getenv("APP_DB_PATH")
	if path == "" {
		path = "agent-control-plane.db"
	}

	db, err := sqlite.Open(context.Background(), path)
	if err != nil {
		log.Fatalf("database startup: %v", err)
	}

	server := web.NewServer(system.NewProbe(), db)
	err = http.ListenAndServe(address, server.Handler())
	db.Close()
	log.Fatal(err)
}
