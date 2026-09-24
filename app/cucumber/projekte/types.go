package projekte

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
	t             *testing.T
	binary        string
	database      string
	address       string
	bindHost      string
	process       *exec.Cmd
	exited        chan error
	logFile       *os.File
	client        *http.Client
	response      *HTTPResponse
	organizations map[string]string
	goals         map[string]string
	projects      map[string]string
	active        string
	createdID     string
	unknownBody   []byte
	rejectedAddr  string
	rejectedLog   string
	rejectedErr   error
	routePages    map[string]HTTPResponse
	negative      *NegativeServer
	browser       *exec.Cmd
	stdin         *bufio.Writer
	stdout        *bufio.Scanner
	page          BrowserPage
}

type HTTPResponse struct {
	Status   int
	Location string
	Body     []byte
	Header   http.Header
}

type Organization struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Goal struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	OrganizationID string `json:"organization_id"`
}

type GoalList struct {
	Goals []Goal `json:"goals"`
}

type Project struct {
	ID             string `json:"id"`
	OrganizationID string `json:"organization_id"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	GoalID         string `json:"goal_id"`
	Tasks          []any  `json:"tasks"`
}

type ProjectList struct {
	Projects []Project `json:"projects"`
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
	URL         string   `json:"url"`
	Epoch       float64  `json:"epoch"`
	Heading     string   `json:"heading"`
	Text        string   `json:"text"`
	Alert       string   `json:"alert"`
	NameError   string   `json:"nameError"`
	GoalError   string   `json:"goalError"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Selected    string   `json:"selected"`
	Options     []string `json:"options"`
	Links       []string `json:"links"`
	Projects    []string `json:"projects"`
}
