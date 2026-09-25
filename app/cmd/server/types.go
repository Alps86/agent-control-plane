package main

import (
	"agentcontrolplane/app/internal/adapter/sqlite"
	appfreigabe "agentcontrolplane/app/internal/app/modellfreigabe"
)

// Bootstrap verdrahtet den lokalen Server beim technischen Einstieg.
type Bootstrap struct {
	bindAddress string
}

type modelChoiceStore struct {
	db *sqlite.Database
}

type modelGrantAccess struct {
	grants *appfreigabe.Service
}
