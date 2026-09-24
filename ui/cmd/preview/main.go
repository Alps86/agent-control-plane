package main

import (
	"log"

	"agentcontrolplane/ui/internal/preview"
)

func main() {
	server, err := preview.New()
	if err != nil {
		log.Fatal(err)
	}

	if err := server.Serve(); err != nil {
		log.Fatal(err)
	}
}
