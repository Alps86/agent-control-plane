package agentenvorlagen

import (
	"bufio"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"testing"

	"agentcontrolplane/app/internal/adapter/sqlite"
	"agentcontrolplane/app/internal/domain/rechte"
)

type Suite struct {
	t              *testing.T
	binary         string
	dbPath         string
	address        string
	listenAddress  string
	process        *exec.Cmd
	exited         chan error
	logFile        *os.File
	client         *http.Client
	response       *HTTPResponse
	orgIDs         map[string]string
	agentIDs       map[string]string
	initialID      string
	initialProfile Profile
	initialURL     string
	selectedTeam   string
	selectedAgent  string
	selectedOrg    string
	enteredName    string
	browser        *exec.Cmd
	stdin          *bufio.Writer
	stdout         *bufio.Scanner
	page           BrowserPage
	negative       *httptest.Server
	negativeDB     *sqlite.Database
}

type HTTPResponse struct {
	Status   int
	Location string
	Body     []byte
	Header   http.Header
}

type Profile struct {
	ID             string    `json:"id"`
	OrganizationID string    `json:"organization_id"`
	Name           string    `json:"name"`
	Role           string    `json:"role"`
	Instructions   string    `json:"instructions"`
	ExecutionKind  string    `json:"execution_kind"`
	TemplateID     string    `json:"template_id"`
	Capabilities   []string  `json:"capabilities"`
	Readiness      Readiness `json:"readiness"`
}

type Readiness struct {
	Ready  bool   `json:"ready"`
	Code   string `json:"code"`
	Reason string `json:"reason"`
}

type AgentList struct {
	Agents []Profile `json:"agents"`
}

type Template struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	Role          string   `json:"role"`
	Instructions  string   `json:"instructions"`
	ExecutionKind string   `json:"execution_kind"`
	Capabilities  []string `json:"capabilities"`
}

type TeamTemplate struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	TemplateIDs []string `json:"template_ids"`
}

type BrowserReply struct {
	OK    bool        `json:"ok"`
	Error string      `json:"error"`
	Page  BrowserPage `json:"page"`
}

type BrowserPage struct {
	URL         string            `json:"url"`
	Epoch       float64           `json:"epoch"`
	Text        string            `json:"text"`
	Heading     string            `json:"heading"`
	Alert       string            `json:"alert"`
	Cards       []string          `json:"cards"`
	Fields      map[string]string `json:"fields"`
	Options     []string          `json:"options"`
	TeamEntries []string          `json:"teamEntries"`
}

type FixedIdentity struct{ Actor rechte.Actor }
