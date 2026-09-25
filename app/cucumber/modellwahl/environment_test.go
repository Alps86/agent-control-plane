package modellwahl

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/cucumber/godog"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func (s *Suite) build() error {
	s.client = &http.Client{Timeout: 3 * time.Second}
	s.binary = filepath.Join(s.t.TempDir(), "server")
	cmd := exec.Command("go", "build", "-o", s.binary, "./cmd/server")
	cmd.Dir = "../.."
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Server-Build: %w: %s", err, output)
	}
	return nil
}

func (s *Suite) fresh() error {
	s.stopServer()
	s.dbPath = filepath.Join(s.t.TempDir(), "modellwahl.sqlite")
	s.credentialDir = filepath.Join(filepath.Dir(s.dbPath), "secrets")
	if err := os.Mkdir(s.credentialDir, 0700); err != nil {
		return err
	}
	s.catalogPath = filepath.Join(s.t.TempDir(), "catalog.json")
	s.secret = "sk-story25-synthetic-never-display"
	s.secretStored = false
	if err := os.WriteFile(s.catalogPath, []byte(catalogFixture), 0600); err != nil {
		return err
	}
	if err := s.startProvider(); err != nil {
		return err
	}
	return s.start()
}

func (s *Suite) start() error {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return err
	}
	s.address = listener.Addr().String()
	listener.Close()
	return s.launchProcess()
}

func (s *Suite) launchProcess() error {
	s.process = exec.Command(s.binary)
	s.process.Env = append(s.serverEnv(), "APP_ADDR="+s.address, "APP_DB_PATH="+s.dbPath,
		"APP_CREDENTIALS_PATH="+filepath.Join(s.credentialDir, "credentials.enc"),
		"APP_CREDENTIAL_KEY_PATH="+filepath.Join(s.credentialDir, "master.key"),
		"APP_OPENROUTER_PROBE_URL="+s.providerServer.URL+"/api/v1/key", "HTTPS_PROXY="+s.providerServer.URL, "HTTP_PROXY="+s.providerServer.URL,
		"NO_PROXY=localhost,127.0.0.1")
	if s.catalogPath != "" {
		s.process.Env = append(s.process.Env, "APP_MODEL_CATALOG_PATH="+s.catalogPath)
	}
	s.process.Stdout, s.process.Stderr = os.Stderr, os.Stderr
	if err := s.process.Start(); err != nil {
		return err
	}
	return s.healthy()
}

func (s *Suite) serverEnv() []string {
	result := make([]string, 0, len(os.Environ()))
	for _, item := range os.Environ() {
		if strings.HasPrefix(item, "APP_MODEL_CATALOG_PATH=") {
			continue
		}
		result = append(result, item)
	}
	return result
}

func (s *Suite) healthy() error {
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		response, err := s.client.Get(s.url("/health"))
		if err == nil {
			response.Body.Close()
			if response.StatusCode == 200 {
				return nil
			}
		}
		time.Sleep(20 * time.Millisecond)
	}
	return fmt.Errorf("Server unter %s nicht erreichbar", s.address)
}

func (s *Suite) stopServer() {
	if s.process == nil {
		return
	}
	_ = s.process.Process.Kill()
	_ = s.process.Wait()
	s.process = nil
}

func (s *Suite) cleanup() { s.stopServer(); s.stopBrowser(); s.stopProvider() }
func (s *Suite) after(ctx context.Context, _ *godog.Scenario, _ error) (context.Context, error) {
	s.stopServer()
	s.stopBrowser()
	s.stopProvider()
	s.orgID, s.agentID, s.secret, s.initial = "", "", "", nil
	s.secretStored = false
	s.invalidCatalogPath, s.invalidOutput, s.invalidExited = "", nil, false
	s.credentialDir, s.settingsStatus = "", 0
	s.providerCalls.Store(0)
	s.blockedExternal.Store(0)
	s.last, s.page = Response{}, BrowserPage{}
	return ctx, nil
}

func (s *Suite) url(path string) string { return "http://" + s.address + path }
func (s *Suite) choicePath() string {
	return "/api/organisationen/" + s.orgID + "/agenten/" + s.agentID + "/modellwahl"
}
func (s *Suite) pagePath() string {
	return "/organisationen/" + s.orgID + "/agenten/" + s.agentID + "/modellwahl"
}

func (s *Suite) request(method, path, body string) error {
	req, err := http.NewRequest(method, s.url(path), strings.NewReader(body))
	if err != nil {
		return err
	}
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	if method != "GET" {
		req.Header.Set("Origin", "http://"+s.address)
	}
	response, err := s.client.Do(req)
	if err != nil {
		return err
	}
	return s.recordResponse(response)
}

func (s *Suite) recordResponse(response *http.Response) error {
	defer response.Body.Close()
	data, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}
	s.last = Response{Status: response.StatusCode, Body: data, Location: response.Header.Get("Location")}
	return nil
}

func (s *Suite) create(name, path string, target any) error {
	if err := s.request("POST", path, name); err != nil {
		return err
	}
	if s.last.Status != 201 {
		return fmt.Errorf("Anlage HTTP %d: %s", s.last.Status, s.last.Body)
	}
	return json.Unmarshal(s.last.Body, target)
}

func (s *Suite) captureInitial() error {
	if err := s.request("GET", s.choicePath(), ""); err != nil {
		return err
	}
	if s.last.Status != 200 {
		return fmt.Errorf("Modellwahl GET HTTP %d: %s", s.last.Status, s.last.Body)
	}
	s.initial = append([]byte(nil), s.last.Body...)
	return nil
}

func (s *Suite) restart() error { s.stopServer(); return s.start() }

const catalogFixture = `{
  "providers": [
    {"id":"codex-abo","name":"Codex-Abo","auth_type":"device_code","connections":[{"reference":"codex-chatgpt","label":"Codex-Abo"}],"models":[
      {"id":"gpt-5.3-codex","source":"https://developers.openai.com/api/docs/models/all","observed_at":"2026-09-24T23:55:13Z","check_status":"unverified","capabilities":{
        "text":{"status":"unverified","source":"https://developers.openai.com/api/docs/models/all","checked_at":"2026-09-24T23:55:13Z"},
        "tools":{"status":"unverified","source":"https://developers.openai.com/api/docs/models/all","checked_at":"2026-09-24T23:55:13Z"}}}]},
    {"id":"openrouter","name":"OpenRouter","auth_type":"api_key","connections":[{"reference":"openrouter-central","label":"OpenRouter"}],"models":[
      {"id":"openai/gpt-4","source":"https://openrouter.ai/docs/api/api-reference/models/get-models","observed_at":"2026-09-24T23:55:13Z","check_status":"unverified","capabilities":{
        "text":{"status":"unverified","source":"https://openrouter.ai/docs/api/api-reference/models/get-models","checked_at":"2026-09-24T23:55:13Z"},
        "tools":{"status":"unverified","source":"https://openrouter.ai/docs/api/api-reference/models/get-models","checked_at":"2026-09-24T23:55:13Z"},
        "audio_input":{"status":"unverified","source":"https://openrouter.ai/docs/api/api-reference/models/get-models","checked_at":"2026-09-24T23:55:13Z"},
        "audio_output":{"status":"unverified","source":"https://openrouter.ai/docs/api/api-reference/models/get-models","checked_at":"2026-09-24T23:55:13Z"},
        "realtime":{"status":"unverified","source":"https://openrouter.ai/docs/api/api-reference/models/get-models","checked_at":"2026-09-24T23:55:13Z"}}}]}
  ]
}`
