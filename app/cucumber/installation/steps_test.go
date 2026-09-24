package installation

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/cucumber/godog"
)

func NewSuite(t *testing.T) *Suite {
	return &Suite{t: t}
}

func (s *Suite) InitializeScenario(sc *godog.ScenarioContext) {
	sc.After(s.afterScenario)
	sc.Step(`^der installierte Go-Server ist mit einer neuen lokalen Datenbank gestartet$`, s.start)
	sc.Step(`^der Start des installierten Go-Servers mit neuer Datenbank auf allen Netzadressen wurde versucht$`, s.startExposed)
	sc.Step(`^ist der Prozess mit einem APP_ADDR-Loopback-Konfigurationsfehler beendet$`, s.wildcardConfigRejected)
	sc.Step(`^auf dem gewählten Wildcard-Port ist kein HTTP-Server erreichbar$`, s.noWildcardListener)
	sc.Step(`^ich den Server mit derselben Datenbank auf Loopback starte$`, s.startExistingLoopback)
	sc.Step(`^ich ohne Cookie oder Authorization-Header GET "([^"]*)" an den Server sende$`, s.get)
	sc.Step(`^antwortet der Server mit HTTP 200 und einer Schemaversion$`, s.healthy)
	sc.Step(`^die Antwort verlangt weder Login noch Registrierung$`, s.noLogin)
	sc.Step(`^die JSON-Antwort enthält keine Budgetfelder oder Budgetlimits$`, s.noBudget)
	s.registerPageSteps(sc)
	s.registerAPISteps(sc)
}

func (s *Suite) build() error {
	s.binary = filepath.Join(s.t.TempDir(), "server")
	cmd := exec.Command("go", "build", "-o", s.binary, "./cmd/server")
	cmd.Dir = "../.."
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Go-Server bauen: %w: %s", err, output)
	}

	return nil
}

func (s *Suite) start() error {
	return s.startAt("127.0.0.1:0")
}

func (s *Suite) startExposed() error {
	return s.startAt("0.0.0.0:0")
}

func (s *Suite) startAt(bind string) error {
	s.dbPath = filepath.Join(s.t.TempDir(), "ops22.sqlite")
	if _, err := os.Stat(s.dbPath); !os.IsNotExist(err) {
		return fmt.Errorf("Testdatenbank ist nicht neu: %v", err)
	}

	s.freshDB = true
	if err := s.reserveAddress(bind); err != nil {
		return err
	}
	if bind == "0.0.0.0:0" {
		return s.launchRejected()
	}
	return s.launch()
}

func (s *Suite) reserveAddress(bind string) error {
	listener, err := net.Listen("tcp", bind)
	if err != nil {
		return fmt.Errorf("Loopback-Port reservieren: %w", err)
	}

	s.bindAddress = listener.Addr().String()
	_, port, _ := net.SplitHostPort(s.bindAddress)
	s.address = net.JoinHostPort("127.0.0.1", port)
	listener.Close()
	return nil
}

func (s *Suite) launchRejected() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, s.binary)
	command.Env = append(os.Environ(), "APP_ADDR="+s.bindAddress, "APP_DB_PATH="+s.dbPath)
	output, err := command.CombinedOutput()
	if ctx.Err() != nil || err == nil {
		return fmt.Errorf("wildcard server did not fail startup")
	}
	s.wildcardRejected = strings.Contains(string(output), "APP_ADDR must be a local loopback address")
	return nil
}

func (s *Suite) wildcardConfigRejected() error {
	if !s.wildcardRejected {
		return fmt.Errorf("wildcard startup lacked APP_ADDR loopback rejection")
	}
	return nil
}

func (s *Suite) noWildcardListener() error {
	client := &http.Client{Timeout: time.Second, Transport: &http.Transport{Proxy: nil}}
	response, err := client.Get("http://" + s.address + "/health")
	if err != nil {
		return nil
	}
	response.Body.Close()
	return fmt.Errorf("wildcard startup left HTTP listener reachable")
}

func (s *Suite) startExistingLoopback() error {
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		return err
	}
	s.bindAddress, s.address = listener.Addr().String(), listener.Addr().String()
	listener.Close()
	return s.launch()
}

func (s *Suite) launch() error {
	s.process = exec.Command(s.binary)
	s.process.Env = append(os.Environ(), "APP_ADDR="+s.bindAddress, "APP_DB_PATH="+s.dbPath)
	if err := s.process.Start(); err != nil {
		return err
	}

	s.exited = make(chan error, 1)
	go s.awaitExit()
	return s.waitHealthy()
}

func (s *Suite) awaitExit() {
	s.exited <- s.process.Wait()
}

func (s *Suite) waitHealthy() error {
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if err := s.get("/health"); err == nil && s.status == http.StatusOK {
			return nil
		}

		time.Sleep(20 * time.Millisecond)
	}

	return fmt.Errorf("gestarteter Server unter %s nicht gesund", s.address)
}

func (s *Suite) get(path string) error {
	request, err := http.NewRequest(http.MethodGet, "http://"+s.address+path, nil)
	if err != nil {
		return err
	}

	client := &http.Client{Timeout: time.Second, CheckRedirect: s.rejectRedirect}
	response, err := client.Do(request)
	if err != nil {
		return err
	}

	defer response.Body.Close()
	s.status, s.header = response.StatusCode, response.Header.Clone()
	s.body, err = io.ReadAll(response.Body)
	return err
}

func (s *Suite) post(path, contentType string, body io.Reader) error {
	request, err := http.NewRequest(http.MethodPost, "http://"+s.address+path, body)
	if err != nil {
		return err
	}

	request.Header.Set("Content-Type", contentType)
	client := &http.Client{Timeout: time.Second, CheckRedirect: s.rejectRedirect}
	response, err := client.Do(request)
	if err != nil {
		return err
	}

	defer response.Body.Close()
	s.status, s.header = response.StatusCode, response.Header.Clone()
	s.body, err = io.ReadAll(response.Body)
	return err
}

func (s *Suite) rejectRedirect(_ *http.Request, _ []*http.Request) error {
	return http.ErrUseLastResponse
}

func (s *Suite) healthy() error {
	if s.status != http.StatusOK || !strings.HasPrefix(s.header.Get("Content-Type"), "application/json") {
		return fmt.Errorf("Health: HTTP %d, Content-Type %q", s.status, s.header.Get("Content-Type"))
	}

	var body Health
	if err := json.Unmarshal(s.body, &body); err != nil {
		return err
	}

	if body.SchemaVersion < 1 {
		return fmt.Errorf("Schemaversion fehlt oder ist ungültig: %s", s.body)
	}

	return nil
}

func (s *Suite) noLogin() error {
	if s.status == http.StatusUnauthorized || s.status == http.StatusForbidden || s.header.Get("WWW-Authenticate") != "" {
		return fmt.Errorf("Health verlangt Anmeldung: HTTP %d", s.status)
	}

	if s.header.Get("Location") != "" {
		return fmt.Errorf("Health leitet weiter: %s", s.header.Get("Location"))
	}

	return nil
}

func (s *Suite) noBudget() error {
	return s.checkNoBudgetJSON(s.body)
}

func (s *Suite) checkNoBudgetJSON(raw []byte) error {
	var body any
	if err := json.Unmarshal(raw, &body); err != nil {
		return err
	}

	return s.checkBudget(body)
}

func (s *Suite) checkBudget(value any) error {
	if object, ok := value.(map[string]any); ok {
		return s.checkBudgetObject(object)
	}

	if entries, ok := value.([]any); ok {
		for _, entry := range entries {
			if err := s.checkBudget(entry); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *Suite) checkBudgetObject(object map[string]any) error {
	for key, value := range object {
		if strings.Contains(strings.ToLower(key), "budget") {
			return fmt.Errorf("Budgetfeld %q in JSON-Antwort", key)
		}

		if err := s.checkBudget(value); err != nil {
			return err
		}
	}

	return nil
}

func (s *Suite) stop() {
	if s.process == nil {
		return
	}

	s.process.Process.Kill()
	<-s.exited
	s.process = nil
}

func (s *Suite) afterScenario(ctx context.Context, _ *godog.Scenario, _ error) (context.Context, error) {
	s.stop()
	s.dbPath, s.address, s.bindAddress, s.status, s.header, s.body = "", "", "", 0, nil, nil
	s.orgID, s.budget, s.created = "", "", false
	s.rejected = nil
	s.rejectedHeader = nil
	s.freshDB = false
	s.wildcardRejected = false
	return ctx, nil
}
