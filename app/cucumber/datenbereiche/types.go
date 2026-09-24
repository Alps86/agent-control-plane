package datenbereiche

import (
	"bufio"
	"net/http"
	"os"
	"os/exec"
	"testing"

	"agentcontrolplane/app/internal/adapter/sqlite"
	appdatenbereich "agentcontrolplane/app/internal/app/datenbereich"
	"agentcontrolplane/app/internal/domain/rechte"
)

type Suite struct {
	t             *testing.T
	binary        string
	database      string
	address       string
	process       *exec.Cmd
	exited        chan error
	logFile       *os.File
	client        *http.Client
	response      *HTTPResponse
	previous      *HTTPResponse
	organizations map[string]string
	agents        map[string]string
	projects      map[string]string
	goals         map[string]string
	dataStore     *sqlite.Database
	dataService   *appdatenbereich.Service
	agentActor    rechte.Actor
	activeAgent   string
	lastError     error
	previousError error
	lastProject   string
	lastName      string
	lastDetail    string
	lastID        string
	lastOrgID     string
	lastAction    string
	beforeDetail  string
	denialCount   int
	browser       *exec.Cmd
	stdin         *bufio.Writer
	stdout        *bufio.Scanner
	page          BrowserPage
}

type HTTPResponse struct {
	Status int
	Body   []byte
	Header http.Header
}

type BrowserReply struct {
	OK    bool        `json:"ok"`
	Error string      `json:"error"`
	Page  BrowserPage `json:"page"`
}

type BrowserPage struct {
	URL     string `json:"url"`
	Heading string `json:"heading"`
	Text    string `json:"text"`
	Status  string `json:"status"`
}

type Scope struct {
	AgentID        string `json:"agent_id"`
	OrganizationID string `json:"organization_id"`
	ProjectID      string `json:"project_id"`
	CanRead        bool   `json:"can_read"`
	CanWrite       bool   `json:"can_write"`
}

type ScopeResponse struct {
	Scopes []Scope `json:"scopes"`
}

type Organization struct {
	ID string `json:"id"`
}

type Agent struct {
	ID             string    `json:"id"`
	OrganizationID string    `json:"organization_id"`
	Name           string    `json:"name"`
	ExecutionKind  string    `json:"execution_kind"`
	Readiness      Readiness `json:"readiness"`
}

type Readiness struct {
	Ready bool   `json:"ready"`
	Code  string `json:"code"`
}

type Goal struct {
	ID string `json:"id"`
}

type FixedIdentity struct {
	Actor rechte.Actor
}
