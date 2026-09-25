package berichtsweg

import (
	"bufio"
	"net/http"
	"os"
	"os/exec"
	"testing"
)

type Suite struct {
	t                                      *testing.T
	binary, database, address              string
	process                                *exec.Cmd
	exited                                 chan error
	logFile                                *os.File
	client                                 *http.Client
	organizations, agents, projects, tasks map[string]string
	response                               Response
	rejectedReason                         string
	foreignURL                             string
	foreignClose                           func()
	foreignReplies, unknownReplies         []Response
	browser                                *exec.Cmd
	browserInput                           *bufio.Writer
	browserOutput                          *bufio.Scanner
	page                                   BrowserPage
	selectedAgent                          string
}

type Response struct {
	Status int
	Body   []byte
	Header http.Header
}
type Organization struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
type Agent struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	OrganizationID string `json:"organization_id"`
	ExecutionKind  string `json:"execution_kind"`
}
type Project struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
type Task struct {
	ID         string `json:"id"`
	Title      string `json:"title"`
	AssigneeID string `json:"assignee_id"`
}
type Lines struct {
	OrganizationID string `json:"organization_id"`
	Lines          []Line `json:"lines"`
}
type Line struct {
	AgentID       string `json:"agent_id"`
	Name          string `json:"name"`
	ParentID      string `json:"parent_id"`
	ExecutionKind string `json:"execution_kind"`
}
type BrowserPage struct {
	URL     string            `json:"url"`
	Heading string            `json:"heading"`
	Text    string            `json:"text"`
	Links   map[string]string `json:"links"`
	Fields  map[string]string `json:"fields"`
	Alert   string            `json:"alert"`
	Reports map[string]string `json:"reports"`
}
