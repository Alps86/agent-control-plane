package codexprofil

import (
	"bufio"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"testing"

	"agentcontrolplane/app/internal/adapter/sqlite"
)

type Suite struct {
	t                                           *testing.T
	dbPath, address, orgID, otherOrgID, agentID string
	browser                                     *exec.Cmd
	server                                      *httptest.Server
	db                                          *sqlite.Database
	negative                                    *httptest.Server
	negativeDB                                  *sqlite.Database
	runtimeScenario                             string
	actionServerPath                            string
	cliPath, bwrapPath                          string
	cliSHA, bwrapSHA, actionServerSHA           string
	input                                       *bufio.Writer
	output                                      *bufio.Scanner
	client                                      *http.Client
	response                                    Response
	page                                        Page
}

type Response struct {
	Method   string
	Status   int
	Body     []byte
	Location string
}

type Organization struct {
	ID string `json:"id"`
}

type Agent struct {
	ID string `json:"id"`
}

type Profile struct {
	AgentID          string `json:"agent_id"`
	OrganizationID   string `json:"organization_id"`
	WorkspaceEnabled bool   `json:"workspace_enabled"`
	WriteEnabled     bool   `json:"write_enabled"`
}

type Readiness struct {
	Ready  bool   `json:"ready"`
	Code   string `json:"code"`
	Reason string `json:"reason"`
}

type RuntimeResult struct {
	Ready      bool           `json:"ready"`
	Code       string         `json:"code"`
	Reason     string         `json:"reason"`
	CLIStarted bool           `json:"cli_started"`
	ActionID   string         `json:"action_id"`
	Checks     []RuntimeCheck `json:"checks"`
}

type RuntimeCheck struct {
	Code   string `json:"code"`
	Status string `json:"status"`
	Detail string `json:"detail"`
}

type BrowserReply struct {
	OK    bool   `json:"ok"`
	Error string `json:"error"`
	Page  Page   `json:"page"`
}

type Page struct {
	URL, Heading, Text                        string
	Epoch                                     float64
	WorkspaceEnabled, WriteEnabled, PathInput bool
}
