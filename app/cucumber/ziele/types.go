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
	goals            map[string]string
	projects         map[string]string
	baseline         map[string]string
	childName        string
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
	Status         string  `json:"status"`
}

type ProjectPath struct {
	ProjectID   string   `json:"project_id"`
	ProjectName string   `json:"project_name"`
	GoalIDs     []string `json:"goal_ids"`
	GoalNames   []string `json:"goal_names"`
}

type GoalTree struct {
	Goals        []Goal        `json:"goals"`
	ProjectPaths []ProjectPath `json:"project_paths"`
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
	URL          string              `json:"url"`
	Epoch        float64             `json:"epoch"`
	Heading      string              `json:"heading"`
	Text         string              `json:"text"`
	Alert        string              `json:"alert"`
	Name         string              `json:"name"`
	Links        []string            `json:"links"`
	Goals        []string            `json:"goals"`
	GoalCards    []BrowserGoal       `json:"goal_cards"`
	ChildErrors  []BrowserChildError `json:"child_errors"`
	MoveErrors   []BrowserMoveError  `json:"move_errors"`
	ProjectPaths []string            `json:"project_paths"`
}

type BrowserChildError struct {
	ID      string `json:"id"`
	Message string `json:"message"`
}

type BrowserMoveError struct {
	ID      string `json:"id"`
	Message string `json:"message"`
}

type BrowserGoal struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Status   string `json:"status"`
	ParentID string `json:"parent_id"`
}
