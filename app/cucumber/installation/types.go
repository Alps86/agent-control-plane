package installation

import (
	"net/http"
	"os/exec"
	"testing"
)

type Suite struct {
	t                *testing.T
	binary           string
	dbPath           string
	address          string
	bindAddress      string
	process          *exec.Cmd
	exited           chan error
	status           int
	header           http.Header
	body             []byte
	orgID            string
	budget           string
	created          bool
	rejected         []byte
	rejectedHeader   http.Header
	freshDB          bool
	wildcardRejected bool
}

type Health struct {
	SchemaVersion int `json:"schema_version"`
}
