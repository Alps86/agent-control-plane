package kommentare

import (
	"bufio"
	"net/http"
	"os/exec"
	"testing"

	"agentcontrolplane/app/internal/adapter/sqlite"
	appkommentar "agentcontrolplane/app/internal/app/kommentar"
	"agentcontrolplane/app/internal/domain/rechte"
	portkommentar "agentcontrolplane/app/internal/port/kommentar"
)

type Suite struct {
	t                                                                *testing.T
	client                                                           *http.Client
	binary, database, address                                        string
	server, browser                                                  *exec.Cmd
	input                                                            *bufio.Writer
	output                                                           *bufio.Scanner
	response                                                         Response
	page                                                             Page
	organizations, projects, agents, tasks                           map[string]string
	activeAgent, commentID, createdAt, sourceKind, sourceID, content string
	savedTimestamp                                                   string
	lastComment                                                      Comment
	lastError                                                        error
	resolver                                                         *ArtifactFixture
	store                                                            *sqlite.Database
	service                                                          *appkommentar.Service
}

type Response struct {
	Status   int
	Body     []byte
	Location string
}
type Entity struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
type Comment struct {
	ID             string     `json:"id"`
	OrganizationID string     `json:"organization_id"`
	TaskID         string     `json:"task_id"`
	Content        string     `json:"content"`
	SourceKind     string     `json:"source_kind"`
	SourceID       string     `json:"source_id"`
	SourceName     string     `json:"source_name"`
	CreatedAt      string     `json:"created_at"`
	UpdatedAt      string     `json:"updated_at"`
	Reference      *Reference `json:"reference"`
}
type Reference struct {
	Type           string `json:"type"`
	ID             string `json:"id"`
	OrganizationID string `json:"organization_id"`
	DisplayName    string `json:"display_name"`
	Link           string `json:"link"`
}
type CommentList struct {
	Comments []Comment `json:"comments"`
}
type FieldError struct {
	Error       string            `json:"error"`
	FieldErrors map[string]string `json:"field_errors"`
}
type Page struct {
	URL, Heading, Text, Alert, TitleError string
	Epoch                                 float64
	Timestamps                            []Timestamp `json:"timestamps"`
	CommentCount                          int         `json:"commentCount"`
	Links                                 []PageLink  `json:"links"`
}
type Timestamp struct {
	Value string `json:"value"`
	Text  string `json:"text"`
}
type PageLink struct {
	Href string `json:"href"`
	Text string `json:"text"`
}
type BrowserReply struct {
	OK    bool   `json:"ok"`
	Error string `json:"error"`
	Page  Page   `json:"page"`
}
type AgentIdentity struct{ Actor rechte.Actor }
type ArtifactFixture struct {
	metadata portkommentar.ArtifactMetadata
	missing  bool
}

func (i AgentIdentity) Actors() []rechte.Actor { return []rechte.Actor{i.Actor} }
