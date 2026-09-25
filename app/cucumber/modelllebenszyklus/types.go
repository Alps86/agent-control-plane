package modelllebenszyklus

import (
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"sync"
	"sync/atomic"
	"testing"

	"agentcontrolplane/app/internal/adapter/sqlite"
	"agentcontrolplane/app/internal/app/modellfreigabe"
	"agentcontrolplane/app/internal/app/modellverbindung"
	"agentcontrolplane/app/internal/app/openrouterverbindung"
	"agentcontrolplane/app/internal/port/credentials"
	"agentcontrolplane/app/internal/port/modellschluessel"
)

const priorToken = "synthetic-codex-access-prior"
const priorRefresh = "synthetic-codex-refresh-prior"
const nextRefresh = "synthetic-codex-refresh-next"
const accountID = "synthetic-account-29"
const oldKey = "sk-or-synthetic-old-29"
const boundKey = "sk-or-synthetic-bound-29"
const newKey = "sk-or-synthetic-new-29"

type Suite struct {
	t               *testing.T
	dir             string
	secretPath      string
	keyPath         string
	store           credentials.Store
	flow            *modellverbindung.Service
	routerService   *openrouterverbindung.Service
	grantService    *modellfreigabe.Service
	db              *sqlite.Database
	organizationID  string
	agentID         string
	invokeErr       error
	binding         modellschluessel.Binding
	issuer          *httptest.Server
	provider        *httptest.Server
	app             *httptest.Server
	issuerState     *IssuerState
	providerState   *ProviderState
	response        []byte
	publicBodies    [][]byte
	status          int
	priorReference  string
	firstProbeCount int
	access          [2]AccessResult
	accessDone      chan struct{}
	process         *exec.Cmd
	processLog      *os.File
	processDone     chan error
	processAddress  string
	processBinary   string
}

type AccessResult struct {
	token   string
	account string
	err     error
}

type IssuerState struct {
	mu      sync.Mutex
	calls   int
	revoked bool
	entered chan struct{}
	release chan struct{}
}

type ProviderState struct {
	mu    sync.Mutex
	calls []string
	posts atomic.Int64
}

type ControlledCaller struct{ suite *Suite }

type CodexView struct {
	State  string `json:"state"`
	Reason string `json:"reason"`
}

type OpenRouterView struct {
	Connected    bool   `json:"connected"`
	Reference    string `json:"reference"`
	Status       string `json:"status"`
	StatusDetail string `json:"status_detail"`
}

var _ http.Handler = (*Suite)(nil)
