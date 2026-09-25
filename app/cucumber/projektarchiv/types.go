package projektarchiv

import (
	"agentcontrolplane/app/internal/domain/rechte"
	"bufio"
	"net/http"
	"os"
	"os/exec"
	"testing"
)

type Suite struct {
	t             *testing.T
	binary        string
	database      string
	address       string
	server        *exec.Cmd
	browser       *exec.Cmd
	input         *bufio.Writer
	output        *bufio.Scanner
	client        *http.Client
	response      Response
	page          BrowserPage
	organizations map[string]string
	projects      map[string]string
	agents        map[string]string
	tasks         map[string]string
	activeOrg     string
	activeProject string
	rejectedTitle string
	runID         string
	runDenied     bool
	activity      ActivityEvent
	raceArchive   Response
	raceTask      Response
	raceTitle     string
	log           *os.File
}

type Response struct {
	Status   int
	Body     []byte
	Location string
}

type Organization struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
type Goal struct {
	ID string `json:"id"`
}
type Project struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
	Tasks  []Task `json:"tasks"`
}
type StatusResponse struct {
	Status string `json:"status"`
}
type ProjectList struct {
	Projects []Project `json:"projects"`
}
type Agent struct {
	ID string `json:"id"`
}
type Task struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	ProjectID string `json:"project_id"`
}
type TaskList struct {
	Tasks []Task `json:"tasks"`
}
type ActivityEvent struct {
	ID             string `json:"id"`
	OrganizationID string `json:"organization_id"`
	ProjectID      string `json:"project_id"`
	TaskID         string `json:"task_id"`
	Kind           string `json:"kind"`
	Actor          string `json:"actor"`
	Source         string `json:"source"`
	OccurredAt     string `json:"occurred_at"`
	ObjectTitle    string `json:"object_title"`
	DeepLink       string `json:"deep_link"`
}
type ActivityList struct {
	Events []ActivityEvent `json:"events"`
}
type ConcurrentResult struct {
	Kind     string
	Response Response
	Err      error
}
type BrowserReply struct {
	OK    bool        `json:"ok"`
	Error string      `json:"error"`
	Page  BrowserPage `json:"page"`
}
type BrowserPage struct {
	URL     string   `json:"url"`
	Heading string   `json:"heading"`
	Text    string   `json:"text"`
	Links   []string `json:"links"`
	Forms   []string `json:"forms"`
}

type RunDirectory struct {
	Actor          rechte.Actor
	OrganizationID string
	Task           rechte.Resource
}
