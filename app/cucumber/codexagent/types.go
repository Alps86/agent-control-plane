package codexagent

import (
	"agentcontrolplane/app/internal/adapter/sqlite"
	"agentcontrolplane/app/internal/domain/rechte"
	"bufio"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"testing"
)

type Suite struct {
	t                                                                *testing.T
	binary, dbPath, address, orgID, agentID, agentName, cliReference string
	process                                                          *exec.Cmd
	browser                                                          *exec.Cmd
	input                                                            *bufio.Writer
	output                                                           *bufio.Scanner
	client                                                           *http.Client
	response                                                         Response
	page                                                             Page
	negative                                                         *httptest.Server
	negativeDB                                                       *sqlite.Database
}

type UnassignedIdentity struct{}

func (UnassignedIdentity) Actors() []rechte.Actor { return nil }

type Response struct {
	Status   int
	Body     []byte
	Location string
}
type Organization struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
type Agent struct {
	ID             string    `json:"id"`
	OrganizationID string    `json:"organization_id"`
	Name           string    `json:"name"`
	ExecutionKind  string    `json:"execution_kind"`
	Capabilities   []string  `json:"capabilities"`
	Readiness      Readiness `json:"readiness"`
}
type Readiness struct {
	Ready  bool   `json:"ready"`
	Code   string `json:"code"`
	Reason string `json:"reason"`
}
type AgentList struct {
	Agents []Agent `json:"agents"`
}
type BrowserReply struct {
	OK    bool   `json:"ok"`
	Error string `json:"error"`
	Page  Page   `json:"page"`
}
type Page struct {
	URL, Heading, Text, Name, ExecutionKind, TemplateID string
	Epoch                                               float64
	Links                                               []string
}
