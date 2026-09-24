package codexcli

import (
	"bytes"
	"context"
	"sync"

	portprofil "agentcontrolplane/app/internal/port/codexprofil"
)

const unverifiedCode = "codex_cli_enforcement_unverified"

// Config enthält ausschließlich vertrauenswürdige serverseitige Runtime-Werte.
type Config struct {
	DatabasePath       string
	CLIPath            string
	CLISHA256          string
	BwrapPath          string
	BwrapSHA256        string
	ActionServerPath   string
	ActionServerSHA256 string
	Scenario           string
}

// Check ist ein inhaltlich begrenzter öffentlicher Prüfbeleg.
type Check struct {
	Code   string `json:"code"`
	Status string `json:"status"`
	Detail string `json:"detail"`
}

// Result meldet nur belegte Grenzen, keinen abgeschlossenen Aufgabenlauf.
type Result struct {
	Ready      bool    `json:"ready"`
	Code       string  `json:"code"`
	Reason     string  `json:"reason"`
	CLIStarted bool    `json:"cli_started"`
	ActionID   string  `json:"action_id,omitempty"`
	Checks     []Check `json:"checks"`
}

// Policy prüft die serverseitige Profilfreigabe vor jedem Start.
type Policy interface {
	Authorize(context.Context, string, string, string) error
}

// Runner bindet den echten CLI-Prozess an den privaten Host-Broker.
type Runner struct {
	config  Config
	policy  Policy
	brokers portprofil.BrokerFactory
	active  chan struct{}
}

type limitedOutput struct {
	bytes.Buffer
	limit int
}

// Child ist der vertrauenswürdige Einstieg innerhalb des OS-Namespace.
type Child struct{}

type childReport struct {
	Started         bool     `json:"started"`
	Completed       bool     `json:"completed"`
	ActionID        string   `json:"action_id,omitempty"`
	ActionStatus    string   `json:"action_status,omitempty"`
	ActionCode      string   `json:"action_code,omitempty"`
	ToolNames       []string `json:"tool_names"`
	HostFilesDenied bool     `json:"host_files_denied"`
	NetworkDenied   bool     `json:"network_denied"`
	UnsupportedTool bool     `json:"unsupported_tool"`
	CLIError        string   `json:"cli_error,omitempty"`
}

type fixture struct {
	mu           sync.Mutex
	scenario     string
	requestCount int
	tools        []string
}
