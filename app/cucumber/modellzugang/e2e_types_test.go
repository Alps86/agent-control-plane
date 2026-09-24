package modellzugang

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"

	"agentcontrolplane/app/internal/app/modellpruefung"
	"agentcontrolplane/app/internal/app/modellverbindung"
)

type httpE2ESuite struct {
	t              *testing.T
	issuer         *e2eIssuer
	issuerServer   *httptest.Server
	responses      *e2eResponses
	responseServer *httptest.Server
	service        *modellverbindung.Service
	appServer      *httptest.Server
	client         *http.Client
	secretPath     string
	lastStatus     int
	lastBody       string
	probeBody      string
	probeStatus    int
	nachweis       *modellpruefung.Nachweis
	routeOnly      bool
	probeRequired  bool
}

type e2eIssuer struct {
	mu           sync.Mutex
	starts       int
	polls        int
	exchanges    int
	accessToken  string
	accountID    string
	refreshToken string
}

type e2eResponses struct {
	mu         sync.Mutex
	mode       string
	requests   []e2eRequest
	metadata   []e2eMetadata
	echoID     string
	echoModel  string
	cancelled  chan struct{}
	firstDelta chan struct{}
	once       sync.Once
}

type e2eRequest struct {
	bearer  string
	account string
	body    string
}

type e2eMetadata struct {
	requestID string
	model     string
}

type e2eReroute struct {
	base   http.RoundTripper
	target *url.URL
}
