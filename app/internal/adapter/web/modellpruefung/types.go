package modellpruefung

import (
	"context"
	"net/http"
	"sync"

	appmodellpruefung "agentcontrolplane/app/internal/app/modellpruefung"
	"agentcontrolplane/app/internal/port/credentials"
	"github.com/cloudwego/eino/schema"
)

const toolName = "verbindungsstatus_lesen"

// StatusAction is the application-validated read-only connection action.
type StatusAction interface {
	Lies(context.Context, string) (appmodellpruefung.Verbindungszustand, error)
}

// Handler exposes one local technical subscription probe.
type Handler struct {
	access      credentials.AccessResolver
	status      StatusAction
	modelID     string
	bindAddress string
	client      *http.Client
	slot        chan struct{}
}

type probeRun struct {
	status        StatusAction
	mu            sync.Mutex
	observations  []probeObservation
	toolRequested bool
	toolCalled    bool
	toolState     string
}

type probeObservation struct {
	requestID     string
	status        string
	usageComplete bool
}

type boundedTransport struct {
	base     http.RoundTripper
	mu       sync.Mutex
	requests int
}

type sseEvent struct {
	State         string `json:"state,omitempty"`
	Kind          string `json:"kind,omitempty"`
	Model         string `json:"model,omitempty"`
	RequestID     string `json:"request_id,omitempty"`
	Tool          bool   `json:"tool,omitempty"`
	ToolState     string `json:"tool_state,omitempty"`
	Sequence      int    `json:"sequence,omitempty"`
	Requests      int    `json:"requests,omitempty"`
	UsageComplete bool   `json:"usage_complete,omitempty"`
}

type streamState struct {
	chunks       int
	bytes        int
	tool         bool
	latestModel  string
	observations []probeObservation
}

var _ interface {
	Info(context.Context) (*schema.ToolInfo, error)
} = (*probeRun)(nil)
