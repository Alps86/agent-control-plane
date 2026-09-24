//go:build browser

package browser

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func (s *Suite) lateBrowserStatus(outcome string) error {
	mode, err := raceMode(outcome)
	if err != nil {
		return err
	}
	app := &raceServer{bridge: s.bridge, mode: mode}
	server := httptest.NewServer(app.handler())
	defer server.Close()
	app.origin = server.URL
	result, err := app.runChrome(server.URL)
	if err != nil {
		return err
	}
	s.browserRace = result
	return app.verifyRequests()
}

func raceMode(outcome string) (string, error) {
	if outcome == "Bestätigung" {
		return "connected", nil
	}
	if outcome == "Abbruch" {
		return "cancelled", nil
	}
	return "", fmt.Errorf("unknown browser outcome %q", outcome)
}

func (a *raceServer) handler() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/", a.bridge.Assets())
	mux.HandleFunc("GET /{$}", a.servePage)
	mux.HandleFunc("POST /settings/modelle/codex/device/start", a.serveStart)
	mux.HandleFunc("GET /settings/modelle/codex/device/status", a.serveStatus)
	mux.HandleFunc("POST /settings/modelle/codex/device/cancel", a.serveCancel)
	return mux
}

func (a *raceServer) servePage(w http.ResponseWriter, _ *http.Request) {
	data, err := a.bridge.LoadFixture("codex-verbindung-idle")
	if err != nil {
		http.Error(w, "fixture missing", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := a.bridge.Render(w, "modelle/codex/page", data); err != nil {
		http.Error(w, "page render failed", http.StatusInternalServerError)
	}
}

func (a *raceServer) serveFixture(w http.ResponseWriter, state string) {
	data, err := a.bridge.LoadFixture("codex-verbindung-" + state)
	if err != nil {
		http.Error(w, "fixture missing", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(data)
}

func (a *raceServer) serveStart(w http.ResponseWriter, r *http.Request) {
	a.starts.Add(1)
	a.startOrigin.Store(r.Header.Get("Origin"))
	if !a.allowOrigin(w, r) {
		return
	}
	a.serveFixture(w, "pending")
}

func (a *raceServer) serveStatus(w http.ResponseWriter, _ *http.Request) {
	a.statuses.Add(1)
	a.serveFixture(w, a.mode)
}

func (a *raceServer) serveCancel(w http.ResponseWriter, r *http.Request) {
	a.cancels.Add(1)
	a.cancelOrigin.Store(r.Header.Get("Origin"))
	if !a.allowOrigin(w, r) {
		return
	}
	a.serveFixture(w, "cancelled")
}

func (a *raceServer) allowOrigin(w http.ResponseWriter, r *http.Request) bool {
	if r.Header.Get("Origin") == a.origin && r.Host == strings.TrimPrefix(a.origin, "http://") {
		return true
	}
	http.Error(w, "origin denied", http.StatusForbidden)
	return false
}

func (a *raceServer) runChrome(url string) (browserRaceResult, error) {
	script, err := filepath.Abs("race.mjs")
	if err != nil {
		return browserRaceResult{}, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, "node", script, url, a.mode)
	var output bytes.Buffer
	command.Stdout, command.Stderr = &output, &output
	if err := command.Run(); err != nil {
		return browserRaceResult{}, fmt.Errorf("Chrome: %w: %s", err, output.String())
	}
	return parseChromeResult(output.Bytes())
}

func parseChromeResult(output []byte) (browserRaceResult, error) {
	var result browserRaceResult
	if err := json.Unmarshal(bytes.TrimSpace(output), &result); err != nil {
		return result, fmt.Errorf("Chrome result: %w: %s", err, output)
	}
	if result.Error != "" {
		return result, fmt.Errorf("Chrome: %s", result.Error)
	}
	return result, nil
}

func (a *raceServer) verifyRequests() error {
	if a.starts.Load() != 1 {
		return fmt.Errorf("browser start requests = %d", a.starts.Load())
	}
	if a.startOrigin.Load() != a.origin {
		return fmt.Errorf("browser start Origin = %v, want %s", a.startOrigin.Load(), a.origin)
	}
	if a.mode == "connected" && a.statuses.Load() < 1 {
		return fmt.Errorf("browser did not request status")
	}
	if a.mode == "cancelled" && a.cancels.Load() != 1 {
		return fmt.Errorf("browser cancel requests = %d", a.cancels.Load())
	}
	if a.mode == "cancelled" && a.cancelOrigin.Load() != a.origin {
		return fmt.Errorf("browser cancel Origin = %v, want %s", a.cancelOrigin.Load(), a.origin)
	}
	return nil
}

func (s *Suite) browserStatusStays(expected string) error {
	want := map[string]string{"Verbunden": "connected", "Anmeldung abgebrochen": "cancelled"}[expected]
	if want == "" || s.browserRace.State != want || s.browserRace.Text != expected || s.browserRace.CodeVisible || s.browserRace.Code != "" {
		return fmt.Errorf("stale pending response changed browser UI: %+v", s.browserRace)
	}
	if strings.Contains(strings.ToLower(s.browserRace.Text), "account-test") {
		return fmt.Errorf("account ID visible in browser")
	}
	return nil
}
