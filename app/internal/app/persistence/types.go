package persistence

import "agentcontrolplane/app/internal/port/persistence"

// Service bietet technische Speicheroperationen an der Anwendungsgrenze an.
type Service struct {
	store persistence.Store
}
