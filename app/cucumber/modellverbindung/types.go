package modellverbindung

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	appflow "agentcontrolplane/app/internal/app/modellverbindung"
	credentialport "agentcontrolplane/app/internal/port/credentials"
)

// Suite hält den Zustand eines fachlichen HTTP-Szenarios.
type Suite struct {
	t           *testing.T
	issuer      *FakeIssuer
	service     *appflow.Service
	handler     http.Handler
	response    *httptestResponse
	store       credentialport.Store
	secretPath  string
	keyPath     string
	openErr     error
	accessToken string
	accountID   string
	accessErr   error
	logs        bytes.Buffer
	priorBundle []byte
}

type httptestResponse struct {
	code int
	body string
	data appflow.Status
}

// FakeIssuer emuliert nur die öffentlich erreichbaren Anbieter-Endpunkte.
type FakeIssuer struct {
	mu             sync.Mutex
	server         *httptest.Server
	startDisabled  bool
	pollOutcome    string
	refreshOutcome string
	polls          int
	starts         int
	refreshes      int
	issueLifetime  int64
	issuedToken    string
	rotatedToken   string
}
