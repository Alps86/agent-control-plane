package prozess

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/cucumber/godog"
)

func (s *Suite) InitializeScenario(sc *godog.ScenarioContext) {
	*s = Suite{t: s.t}
	sc.After(s.cleanup)
	s.registerSteps(sc)
}

func (s *Suite) start() error {
	dir := s.t.TempDir()
	s.binary = filepath.Join(dir, "server")
	s.dbPath = filepath.Join(dir, "app.sqlite")
	s.secretPath = filepath.Join(dir, "credentials", "secrets.enc")
	s.keyPath = filepath.Join(dir, "credentials", "master.key")
	s.provider = httptest.NewServer(http.HandlerFunc(s.probe))
	if err := s.chooseAddress(); err != nil {
		return err
	}
	if err := s.build(); err != nil {
		return err
	}
	return s.launch()
}

func (s *Suite) chooseAddress() error {
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		return err
	}
	s.address = listener.Addr().String()
	return listener.Close()
}

func (s *Suite) build() error {
	command := exec.Command("go", "build", "-o", s.binary, "./cmd/server")
	command.Dir = "../../.."
	output, err := command.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Server-Build: %w: %s", err, output)
	}
	return nil
}

func (s *Suite) launch() error {
	logFile, err := os.CreateTemp(s.t.TempDir(), "story79-server-")
	if err != nil {
		return err
	}
	command := exec.Command(s.binary)
	command.Env = s.environment()
	command.Stdout, command.Stderr = logFile, logFile
	if err := command.Start(); err != nil {
		_ = logFile.Close()
		return err
	}
	s.process, s.logFile = command, logFile
	return s.healthy()
}

func (s *Suite) environment() []string {
	return append(os.Environ(), "APP_ADDR="+s.address, "APP_DB_PATH="+s.dbPath,
		"APP_CREDENTIALS_PATH="+s.secretPath, "APP_CREDENTIAL_KEY_PATH="+s.keyPath,
		"APP_OPENROUTER_PROBE_URL="+s.provider.URL, "APP_MODEL_CATALOG_PATH=")
}

func (s *Suite) healthy() error {
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		err := s.request(http.MethodGet, "/health", "")
		if err == nil && s.code == http.StatusOK {
			return nil
		}
		time.Sleep(25 * time.Millisecond)
	}
	logs, _ := os.ReadFile(s.logFile.Name())
	return fmt.Errorf("Server nicht erreichbar: HTTP %d, %s", s.code, logs)
}

func (s *Suite) stop() {
	if s.process == nil {
		return
	}
	_ = s.process.Process.Kill()
	_ = s.process.Wait()
	_ = s.logFile.Close()
	s.process = nil
}

func (s *Suite) restart() error { s.stop(); return s.launch() }

func (s *Suite) cleanup(ctx context.Context, _ *godog.Scenario, _ error) (context.Context, error) {
	s.stop()
	if s.provider != nil {
		s.provider.Close()
	}
	return ctx, nil
}

func (s *Suite) probe(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/api/v1/key" || !strings.HasSuffix(r.Header.Get("Authorization"), key) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	_, _ = w.Write([]byte(`{"data":{"label":"story79"}}`))
}
