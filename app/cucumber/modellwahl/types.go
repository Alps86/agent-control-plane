package modellwahl

import (
	"bufio"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"sync/atomic"
	"testing"
)

type Suite struct {
	t                                                            *testing.T
	client                                                       *http.Client
	binary, dbPath, catalogPath, address, orgID, agentID, secret string
	credentialDir                                                string
	process, browser                                             *exec.Cmd
	input                                                        *bufio.Writer
	output                                                       *bufio.Scanner
	last                                                         Response
	crossRead, crossWrite, unknown                               Response
	initial                                                      []byte
	page                                                         BrowserPage
	secretStored                                                 bool
	invalidCatalogPath                                           string
	invalidOutput                                                []byte
	invalidExited                                                bool
	settingsStatus                                               int
	providerServer                                               *httptest.Server
	providerCalls, blockedExternal                               atomic.Int64
}

type Response struct {
	Status   int
	Body     []byte
	Location string
}

type Organization struct {
	ID string `json:"id"`
}
type Agent struct {
	ID            string `json:"id"`
	ExecutionKind string `json:"execution_kind"`
}

type Selection struct {
	Provider   string `json:"provider_id"`
	Connection string `json:"connection_id"`
	Model      string `json:"model_id"`
}

type ChoiceView struct {
	ExecutionKind string     `json:"execution_kind"`
	Selection     Selection  `json:"selection"`
	Providers     []Provider `json:"providers"`
}

type Provider struct {
	ID          string       `json:"id"`
	AuthType    string       `json:"auth_type"`
	Selectable  bool         `json:"selectable"`
	Connections []Connection `json:"connections"`
	Models      []Model      `json:"models"`
}
type Connection struct {
	Reference string `json:"reference"`
}
type Model struct {
	ID           string                `json:"id"`
	Source       string                `json:"source"`
	ObservedAt   string                `json:"observed_at"`
	CheckStatus  string                `json:"check_status"`
	Capabilities map[string]Capability `json:"capabilities"`
}
type Capability struct {
	Status    string `json:"status"`
	Source    string `json:"source"`
	CheckedAt string `json:"checked_at"`
}

type OpenRouterStatus struct {
	Reference string `json:"reference"`
	Status    string `json:"status"`
}
type GrantStatus struct {
	Reference           string `json:"reference"`
	ConnectionStatus    string `json:"connection_status"`
	OrganizationGranted bool   `json:"organization_granted"`
	AgentGranted        bool   `json:"agent_granted"`
	Allowed             bool   `json:"allowed"`
	Reason              string `json:"reason"`
}

type BrowserReply struct {
	OK    bool        `json:"ok"`
	Error string      `json:"error"`
	Page  BrowserPage `json:"page"`
}

type BrowserPage struct {
	URL           string                       `json:"url"`
	Text          string                       `json:"text"`
	HTML          string                       `json:"html"`
	Selected      string                       `json:"selected"`
	Expanded      bool                         `json:"expanded"`
	ModelEvidence BrowserModelEvidence         `json:"modelEvidence"`
	Capabilities  map[string]BrowserCapability `json:"capabilities"`
}

type BrowserModelEvidence struct {
	Source     string `json:"source"`
	ObservedAt string `json:"observedAt"`
	Status     string `json:"status"`
}
type BrowserCapability struct {
	Text      string `json:"text"`
	CheckedAt string `json:"checkedAt"`
}
