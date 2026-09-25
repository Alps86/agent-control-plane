package aufgaben

import (
	"bufio"
	"net/http"
	"os/exec"
	"testing"
)

type Suite struct {
	t               *testing.T
	binary          string
	database        string
	address         string
	process         *exec.Cmd
	client          *http.Client
	response        Response
	organizations   map[string]string
	projects        map[string]string
	agents          map[string]string
	tasks           map[string]string
	activeOrg       string
	activeProject   string
	createdID       string
	createdAssignee string
	detailPath      string
	detailText      string
	browser         *exec.Cmd
	input           *bufio.Writer
	output          *bufio.Scanner
	page            Page
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
	ID   string `json:"id"`
	Name string `json:"name"`
}

type ProjectList struct {
	Projects []Project `json:"projects"`
}

type Agent struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	OrganizationID string `json:"organization_id"`
	ExecutionKind  string `json:"execution_kind"`
	Status         string `json:"status"`
}

type Task struct {
	ID             string `json:"id"`
	OrganizationID string `json:"organization_id"`
	ProjectID      string `json:"project_id"`
	Title          string `json:"title"`
	Description    string `json:"description"`
	Priority       string `json:"priority"`
	AssigneeID     string `json:"assignee_id"`
	Status         string `json:"status"`
}

type TaskList struct {
	Tasks []Task `json:"tasks"`
}

type FieldError struct {
	Error       string            `json:"error"`
	FieldErrors map[string]string `json:"field_errors"`
}

type BrowserReply struct {
	OK    bool   `json:"ok"`
	Error string `json:"error"`
	Page  Page   `json:"page"`
}

type Page struct {
	URL                 string   `json:"url"`
	Epoch               float64  `json:"epoch"`
	Heading             string   `json:"heading"`
	Text                string   `json:"text"`
	Alert               string   `json:"alert"`
	Title               string   `json:"title"`
	Description         string   `json:"description"`
	Priority            string   `json:"priority"`
	Assignee            string   `json:"assignee"`
	AssigneeMultiple    bool     `json:"assigneeMultiple"`
	AssigneeSelectCount int      `json:"assigneeSelectCount"`
	AssigneeOptions     []string `json:"assigneeOptions"`
	TitleError          string   `json:"titleError"`
	AssigneeError       string   `json:"assigneeError"`
	ProjectError        string   `json:"projectError"`
	ProjectName         string   `json:"projectName"`
	ProjectReadOnly     bool     `json:"projectReadOnly"`
	TaskLinks           []string `json:"taskLinks"`
}
