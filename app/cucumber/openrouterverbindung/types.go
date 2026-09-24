package openrouterverbindung

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

type Suite struct {
	t               *testing.T
	app             *httptest.Server
	provider        *httptest.Server
	providerCalls   []string
	responseStatus  int
	responseBody    []byte
	firstReference  string
	previousKey     string
	currentKey      string
	storeDir        string
	last            PublicResponse
	originMode      string
	hostOverride    string
	htmlBody        []byte
	jsonBody        []byte
	probeEntered    chan string
	probeRelease    chan struct{}
	probeOnce       sync.Once
	releaseOnce     sync.Once
	mutationArrived chan struct{}
	mutationHandled chan struct{}
	productHandler  http.Handler
	checkDone       chan AsyncResult
	mutationDone    chan AsyncResult
	callsAtRelease  int
	remoteServer    *httptest.Server
	remotePeerAddr  chan string
	remoteStatus    int
	remoteBody      []byte
}

type AsyncResult struct {
	Status int
	Body   []byte
	Err    error
}

type PublicResponse struct {
	Connected    bool   `json:"connected"`
	Reference    string `json:"reference"`
	Status       string `json:"status"`
	StatusDetail string `json:"status_detail"`
	Notice       string `json:"notice"`
	Error        string `json:"error"`
}
