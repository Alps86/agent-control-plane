package prozess

import (
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"sync/atomic"
	"testing"
)

const syntheticKey = "sk-or-story23-PROCESS-TEST-ONLY-482951"
const apiPath = "/api/settings/modellanbieter/openrouter"
const htmlPath = "/settings/modellanbieter/openrouter"

type Suite struct {
	t           *testing.T
	binary      string
	dbPath      string
	secretPath  string
	keyPath     string
	address     string
	server      *exec.Cmd
	logFile     *os.File
	exited      chan error
	client      *http.Client
	status      int
	contentType string
	body        []byte
	html        []byte
	reference   string
	requested   []string
	proxy       *httptest.Server
	proxyCalls  atomic.Int64
}

type PublicConnection struct {
	Connected bool   `json:"connected"`
	Reference string `json:"reference"`
	Status    string `json:"status"`
}

type ChromeResult struct {
	OK    bool   `json:"ok"`
	Error string `json:"error"`
}
