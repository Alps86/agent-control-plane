package sqlite

import (
	"os"
	"os/exec"
	"testing"

	"agentcontrolplane/app/internal/adapter/sqlite"
	"agentcontrolplane/app/internal/app/persistence"
)

// Suite verwaltet nur den externen Serverprozess und seine öffentliche Antwort.
type Suite struct {
	t          *testing.T
	binary     string
	dbPath     string
	address    string
	process    *exec.Cmd
	logFile    *os.File
	exited     chan error
	version    int
	versionSet bool
	store      *sqlite.Database
	service    *persistence.Service
	failure    error
	retried    bool
}

type Health struct {
	SchemaVersion int `json:"schema_version"`
}
