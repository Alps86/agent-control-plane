package organisation

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
	t            *testing.T
	binary       string
	dbPath       string
	address      string
	process      *exec.Cmd
	exited       chan error
	logFile      *os.File
	client       *http.Client
	response     *HTTPResponse
	unknown      *HTTPResponse
	organization Organization
	negative     *NegativeServer
	browser      *exec.Cmd
	stdin        *bufio.Writer
	stdout       *bufio.Scanner
	page         BrowserPage
}

type HTTPResponse struct {
	Status   int
	Location string
	Body     []byte
	Header   http.Header
}

type Organization struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	WorkflowPolicy Policy `json:"workflow_policy"`
}

type Policy struct {
	WorkTransitions       []string `json:"work_transitions"`
	DelegationTransitions []string `json:"delegation_transitions"`
}

type OrganizationList struct {
	Organizations []Organization `json:"organizations"`
}
type FieldError struct {
	Error       string            `json:"error"`
	FieldErrors map[string]string `json:"field_errors"`
}

type NegativeServer struct {
	DB     *sqlite.Database
	Server *httptest.Server
}

type FixedIdentity struct{ Actor rechte.Actor }

type BrowserReply struct {
	OK    bool        `json:"ok"`
	Error string      `json:"error"`
	Page  BrowserPage `json:"page"`
}
type BrowserPage struct {
	URL         string        `json:"url"`
	Epoch       float64       `json:"epoch"`
	Text        string        `json:"text"`
	Heading     string        `json:"heading"`
	Alert       string        `json:"alert"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Cards       []BrowserCard `json:"cards"`
}
type BrowserCard struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}
