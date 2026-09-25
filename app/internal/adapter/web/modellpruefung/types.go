package modellpruefung

import (
	"context"
	"io"
	"net/http"
	"sync"

	appmodellpruefung "agentcontrolplane/app/internal/app/modellpruefung"
	"agentcontrolplane/app/internal/port/credentials"
	"github.com/cloudwego/eino/schema"
)

const toolName = "verbindungsstatus_lesen"
const expectedAnswer = "MS01-PROBE-OK"

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
	modelID       string
	mu            sync.Mutex
	observations  []probeObservation
	toolRequested bool
	toolCalled    bool
	toolState     string
}

type probeObservation struct {
	requestID          string
	status             string
	usageComplete      bool
	modelMatchesConfig bool
}

type boundedTransport struct {
	base                 http.RoundTripper
	mu                   sync.Mutex
	requests             int
	chunkBeforeCompleted bool
	progress             chan struct{}
	progressSent         bool
}

type streamDelivery struct {
	message  *schema.Message
	err      error
	progress bool
}

type streamStart struct {
	reader *schema.StreamReader[*schema.Message]
	err    error
}

type auditBody struct {
	inner          io.ReadCloser
	transport      *boundedTransport
	index          int
	line           []byte
	eventType      string
	deltaSeen      bool
	deltaPending   bool
	eventValid     bool
	completedValid bool
	dataLines      int
	overflowed     bool
}

type auditEvent struct {
	Type     string `json:"type"`
	Delta    string `json:"delta"`
	Response struct {
		Status string `json:"status"`
	} `json:"response"`
}

type sseEvent struct {
	State                      string `json:"state,omitempty"`
	Kind                       string `json:"kind,omitempty"`
	Model                      string `json:"model,omitempty"`
	RequestID                  string `json:"request_id,omitempty"`
	Tool                       bool   `json:"tool,omitempty"`
	ToolState                  string `json:"tool_state,omitempty"`
	Sequence                   int    `json:"sequence,omitempty"`
	Requests                   int    `json:"requests,omitempty"`
	UsageComplete              bool   `json:"usage_complete,omitempty"`
	ProviderModelMatchesConfig *bool  `json:"provider_model_matches_config,omitempty"`
	AnswerMatchesExpected      *bool  `json:"answer_matches_expected,omitempty"`
	ChunkBeforeCompleted       *bool  `json:"chunk_before_response_completed,omitempty"`
}

type streamState struct {
	chunks       int
	bytes        int
	finalText    string
	tool         bool
	latestModel  string
	observations []probeObservation
}

var _ interface {
	Info(context.Context) (*schema.ToolInfo, error)
} = (*probeRun)(nil)
