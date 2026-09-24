package lauf

import (
	"agentcontrolplane/app/internal/adapter/sqlite"
	applauf "agentcontrolplane/app/internal/app/lauf"
	domainlauf "agentcontrolplane/app/internal/domain/lauf"
	domainrechte "agentcontrolplane/app/internal/domain/rechte"
)

// Suite beobachtet ausschließlich Ergebnisse der öffentlichen Lauf-Anwendung.
type Suite struct {
	directory *Directory
	database  *sqlite.Database
	service   *applauf.Service
	path      string
	taskID    string
	actorID   string
	run       domainlauf.Run
	lastRun   domainlauf.Run
	accepted  bool
	lastError error
	parallel  []Reservation
}

// Reservation hält eine Antwort eines gleichzeitigen Anwendungsaufrufs.
type Reservation struct {
	run      domainlauf.Run
	accepted bool
	err      error
}

// Directory ist die serverseitige Rechtequelle für den Blackbox-Test.
type Directory struct {
	actors        []domainrechte.Actor
	organizations map[string][]string
	resources     map[string]domainrechte.Resource
}
