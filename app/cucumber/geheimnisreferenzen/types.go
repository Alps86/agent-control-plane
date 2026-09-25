package geheimnisreferenzen

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"agentcontrolplane/app/internal/adapter/sqlite"
	appgrant "agentcontrolplane/app/internal/app/modellfreigabe"
	"agentcontrolplane/app/internal/app/modellverbindung"
	"agentcontrolplane/app/internal/app/openrouterverbindung"
)

const testKey = "sk-or-story79-synthetic-secret"
const testAccess = "story79-synthetic-access-token"
const testRefresh = "story79-synthetic-refresh-token"

type Suite struct {
	t             *testing.T
	dir           string
	address       string
	app           *httptest.Server
	issuer        *httptest.Server
	provider      *httptest.Server
	issuerState   *Issuer
	providerCalls atomic.Int64
	db            *sqlite.Database
	codex         *modellverbindung.Service
	router        *openrouterverbindung.Service
	grants        *appgrant.Service
	orgID         string
	agentID       string
	last          []byte
	status        int
	snapshots     [][]byte
	before        []byte
	invokeErr     error
	later         bool
}

type Issuer struct {
	revoked atomic.Bool
	short   atomic.Bool
}

type ProviderView struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	AuthType  string `json:"auth_type"`
	Status    string `json:"status"`
	Detail    string `json:"status_detail"`
	Reference string `json:"connection_reference"`
	Ready     bool   `json:"ready"`
}

type StatusView struct {
	Providers []ProviderView `json:"providers"`
}

type Caller struct{ suite *Suite }

type LaterPort struct{}
type ChoiceStore struct{ db *sqlite.Database }
type ChoiceGrant struct{ grants *appgrant.Service }

var _ http.Handler = (*Issuer)(nil)
