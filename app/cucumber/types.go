package cucumber

import (
	"net/http"
	"os/exec"
	"testing"
)

// Suite hält nur den Zustand des gestarteten Blackbox-Prozesses.
type Suite struct {
	t        *testing.T
	process  *exec.Cmd
	baseURL  string
	response *http.Response
}
