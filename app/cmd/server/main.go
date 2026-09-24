package main

import (
	"log"
	"net/http"
	"os"

	"agentcontrolplane/app/internal/adapter/web"
	"agentcontrolplane/app/internal/app/system"
)

func main() {
	address := os.Getenv("APP_ADDR")
	if address == "" {
		address = "127.0.0.1:8080"
	}

	server := web.NewServer(system.NewProbe())
	log.Fatal(http.ListenAndServe(address, server.Handler()))
}
