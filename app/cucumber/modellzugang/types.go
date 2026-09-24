package modellzugang

import (
	"context"
	"io"
	"net/http"
	"sync"
	"testing"

	"agentcontrolplane/app/internal/adapter/model/codexabo"
	"agentcontrolplane/app/internal/app/modellpruefung"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/flow/agent/react"
	"github.com/cloudwego/eino/schema"
)

type Suite struct {
	t              *testing.T
	mu             sync.Mutex
	local          bool
	mode           string
	requests       []string
	fixture        *statusFixture
	pruefung       *modellpruefung.Pruefung
	nachweis       *modellpruefung.Nachweis
	agentOnly      *modellpruefung.Nachweis
	transportBeleg modellpruefung.Transportbeleg
	reportRequests []reportRequest
	reportPath     string
	reportVictim   string
	reportErr      error
	agent          *react.Agent
	rawAdapter     model.ToolCallingChatModel
	answer         string
	modelErr       error
	provider       string
	live           bool
	trace          *requestTrace
	observations   *observationCollector
	tool           *fachTool
	messages       []*schema.Message
	chunks         int
	proof          string
	evidence       limitEvidence
	cancel         context.CancelFunc
	streamDone     chan struct{}
	firstChunk     chan struct{}
	priorRequests  int
	cancelled      bool
	cancelBody     *cancelBody
	unknownUsage   bool
}

type statusFixture struct {
	mu       sync.Mutex
	status   modellpruefung.Status
	accesses int
}

type fachTool struct {
	pruefung *modellpruefung.Pruefung
	agent    string
	entered  chan struct{}
	hold     bool
}

type credential struct {
	token   string
	account string
}

type localTransport struct {
	suite *Suite
}

type requestTrace struct {
	mu              sync.Mutex
	base            http.RoundTripper
	expectedAccount string
	requests        []requestObservation
	bodyClosed      bool
}

type requestObservation struct {
	host           string
	path           string
	status         int
	requestID      string
	accountMatched bool
	hasBearer      bool
	hasToolOutput  bool
	hasMarker      bool
}

type observationCollector struct {
	mu      sync.Mutex
	entries []codexabo.Observation
}

type trackedBody struct {
	io.ReadCloser
	trace *requestTrace
}

type cancelBody struct {
	mu     sync.Mutex
	once   sync.Once
	data   []byte
	offset int
	done   chan struct{}
	closed bool
}

type limitEvidence struct {
	Provider   string `json:"provider"`
	AccountRef string `json:"account_ref"`
	ObservedAt string `json:"observed_at"`
	HTTPStatus int    `json:"http_status"`
	RequestID  string `json:"request_id"`
	Source     string `json:"source"`
}

type reportRequest struct {
	RequestID      string          `json:"request_id"`
	Model          string          `json:"model"`
	HTTPStatus     int             `json:"http_status"`
	ProviderStatus string          `json:"provider_status"`
	Usage          *codexabo.Usage `json:"usage,omitempty"`
}

type reportArtifact struct {
	Story             string                        `json:"story"`
	Execution         string                        `json:"execution"`
	Provider          string                        `json:"provider"`
	AccountSHA256     string                        `json:"account_sha256"`
	ObservedUTC       string                        `json:"observed_utc"`
	GateOpen          bool                          `json:"gate_open"`
	TransportEvidence modellpruefung.Transportbeleg `json:"transport_evidence"`
	Requests          []reportRequest               `json:"requests"`
}
