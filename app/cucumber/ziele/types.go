package ziele

import (
	"bufio"
	"net/http"
	"os"
	"os/exec"
	"testing"

	"agentcontrolplane/app/internal/adapter/sqlite"
	"agentcontrolplane/app/internal/domain/rechte"
	"net/http/httptest"
)

type Suite struct {
	t                *testing.T
	binary           string
	database         string
	address          string
	bindHost         string
	process          *exec.Cmd
	exited           chan error
	logFile          *os.File
	client           *http.Client
	response         *HTTPResponse
	organizations    map[string]string
	active           string
	createdID        string
	unknownBody      []byte
	negative         *NegativeServer
	browser          *exec.Cmd
	stdin            *bufio.Writer
	stdout           *bufio.Scanner
	page             BrowserPage
	wildcardRejected bool
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
	ID             string  `json:"id"`
	Name           string  `json:"name"`
	OrganizationID string  `json:"organization_id"`
	ParentGoalID   *string `json:"parent_goal_id"`
	ProjectID      *string `json:"project_id"`
}

type GoalList struct {
	Goals []Goal `json:"goals"`
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
	URL     string   `json:"url"`
	Epoch   float64  `json:"epoch"`
	Heading string   `json:"heading"`
	Text    string   `json:"text"`
	Alert   string   `json:"alert"`
	Name    string   `json:"name"`
	Links   []string `json:"links"`
	Goals   []string `json:"goals"`
}
