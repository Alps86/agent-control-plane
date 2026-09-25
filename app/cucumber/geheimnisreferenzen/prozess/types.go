package prozess

import (
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"testing"
)

const key = "sk-or-story79-browser-synthetic"

type Suite struct {
	t          *testing.T
	binary     string
	address    string
	dbPath     string
	secretPath string
	keyPath    string
	process    *exec.Cmd
	logFile    *os.File
	provider   *httptest.Server
	response   []byte
	code       int
	page       Page
}

type Page struct {
	OK             bool   `json:"ok"`
	Error          string `json:"error"`
	Codex          string `json:"codex"`
	OpenRouter     string `json:"openrouter"`
	CodexAuth      string `json:"codexAuth"`
	OpenRouterAuth string `json:"openrouterAuth"`
	CodexRef       string `json:"codexRef"`
	OpenRouterRef  string `json:"openrouterRef"`
	HTML           string `json:"html"`
}

var _ http.Handler = http.HandlerFunc(nil)
