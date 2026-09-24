package ops01

import (
	"os"
	"os/exec"
	"testing"
)

type Suite struct {
	t          *testing.T
	binary     string
	dbPath     string
	address    string
	process    *exec.Cmd
	logFile    *os.File
	exited     chan error
	version    int
	fileInfo   os.FileInfo
	startupLog string
}

type Health struct {
	SchemaVersion int `json:"schema_version"`
}
