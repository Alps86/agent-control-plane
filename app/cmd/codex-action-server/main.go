package main

import (
	"log"
	"os"

	"agentcontrolplane/app/internal/adapter/runtime/codexcli"
)

func main() {
	if len(os.Args) == 2 && os.Args[1] == "--probe-runtime" {
		if err := codexcli.NewChild().Run(); err != nil {
			log.Fatal(err)
		}

		return
	}

	if len(os.Args) != 1 {
		os.Exit(2)
	}

	if err := (&ActionServer{input: os.Stdin, output: os.Stdout}).Run(); err != nil {
		log.Fatal(err)
	}
}
