package sqlite

import (
	"context"
	"encoding/json"
	"fmt"
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

// NewSuite erzeugt eine isolierte externe Serverprobe.
func NewSuite(t *testing.T) *Suite {
	return &Suite{t: t}
}

func (s *Suite) InitializeScenario(sc *godog.ScenarioContext) {
	sc.After(s.afterScenario)
	sc.Step(`^ein neuer beschreibbarer lokaler Datenpfad ist für den Server konfiguriert$`, s.writablePath)
	sc.Step(`^ein für den Server unzugänglicher lokaler Datenpfad ist konfiguriert$`, s.inaccessiblePath)
	sc.Step(`^ich den Server starte$`, s.start)
	sc.Step(`^ist der Server über seine öffentliche HTTP-Grenze erreichbar$`, s.reachable)
	sc.Step(`^die Migrationsversion des lokalen Datenbestands ist ermittelt$`, s.captureVersion)
	sc.Step(`^ich den Server beende und mit demselben Datenpfad erneut starte$`, s.restart)
	sc.Step(`^die Migrationsversion des lokalen Datenbestands ist unverändert$`, s.unchangedVersion)
	sc.Step(`^beendet er den Start mit einem verständlichen Datenbankfehler$`, s.startupError)
	sc.Step(`^der Server ist über seine öffentliche HTTP-Grenze nicht erreichbar$`, s.notReachable)
	s.initializeMigrationScenario(sc)
}

func (s *Suite) initializeMigrationScenario(sc *godog.ScenarioContext) {
	sc.Step(`^ein lokaler Datenbestand mit öffentlich lesbarer Schemaversion ist vorhanden$`, s.migrationBase)
	sc.Step(`^eine neue versionierte Migration mit mehreren Änderungen ist konfiguriert$`, s.configureFailingMigration)
	sc.Step(`^ich diese Migration über die öffentliche Anwendungsfassade ausführe und ein späterer Schritt bewusst fehlschlägt$`, s.failingMigration)
	sc.Step(`^meldet die Anwendungsfassade den Migrationsfehler$`, s.migrationFailed)
	sc.Step(`^die öffentlich gelesene Schemaversion ist unverändert$`, s.versionRemained)
	sc.Step(`^ich die korrigierte Migration über dieselbe Anwendungsfassade erneut ausführe$`, s.correctedMigration)
	sc.Step(`^wird die korrigierte Migration einschließlich ihres unveränderten ersten Schritts ohne Teiländerung aus dem Fehlversuch erfolgreich abgeschlossen$`, s.firstChangeRolledBack)
	sc.Step(`^die neue Schemaversion ist öffentlich lesbar$`, s.correctedVersion)
	sc.Step(`^alle Änderungen dieser Migration sind gemeinsam wirksam$`, s.allChangesPresent)
}

func (s *Suite) build() error {
	s.binary = filepath.Join(s.t.TempDir(), "server")
	cmd := exec.Command("go", "build", "-o", s.binary, "./cmd/server")
	cmd.Dir = "../.."
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("server build: %w: %s", err, output)
	}

	return nil
}

func (s *Suite) writablePath() error {
	s.dbPath = filepath.Join(s.t.TempDir(), "app.sqlite")
	return nil
}

func (s *Suite) inaccessiblePath() error {
	blocked := filepath.Join(s.t.TempDir(), "blocked")
	if err := os.WriteFile(blocked, []byte("file"), 0600); err != nil {
		return err
	}

	s.dbPath = filepath.Join(blocked, "app.sqlite")
	return nil
}

func (s *Suite) start() error {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return err
	}

	s.address = listener.Addr().String()
	listener.Close()
	return s.launch()
}

func (s *Suite) launch() error {
	logFile, err := os.CreateTemp(s.t.TempDir(), "server-log-")
	if err != nil {
		return err
	}

	s.logFile = logFile
	s.process = exec.Command(s.binary)
	s.process.Env = append(os.Environ(), "APP_ADDR="+s.address, "APP_DB_PATH="+s.dbPath)
	s.process.Stdout, s.process.Stderr = logFile, logFile
	if err := s.process.Start(); err != nil {
		return err
	}

	s.exited = make(chan error, 1)
	go s.awaitExit()
	return nil
}

func (s *Suite) awaitExit() {
	s.exited <- s.process.Wait()
}

func (s *Suite) reachable() error {
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if err := s.checkHealth(); err == nil {
			return nil
		}

		time.Sleep(20 * time.Millisecond)
	}

	return fmt.Errorf("server not reachable at %s", s.address)
}

func (s *Suite) checkHealth() error {
	client := &http.Client{Timeout: 200 * time.Millisecond}
	response, err := client.Get("http://" + s.address + "/health")
	if err != nil {
		return err
	}

	response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("health HTTP status %d", response.StatusCode)
	}

	return nil
}

func (s *Suite) healthVersion() (int, error) {
	client := &http.Client{Timeout: 2 * time.Second}
	response, err := client.Get("http://" + s.address + "/health")
	if err != nil {
		return 0, err
	}

	defer response.Body.Close()
	if response.StatusCode != http.StatusOK || !strings.HasPrefix(response.Header.Get("Content-Type"), "application/json") {
		return 0, fmt.Errorf("health response: status %d, content type %q", response.StatusCode, response.Header.Get("Content-Type"))
	}

	var health Health
	if err := json.NewDecoder(response.Body).Decode(&health); err != nil {
		return 0, err
	}

	return health.SchemaVersion, nil
}

func (s *Suite) captureVersion() error {
	version, err := s.healthVersion()
	if err != nil {
		return err
	}

	if version < 1 {
		return fmt.Errorf("invalid schema version %d", version)
	}

	s.version, s.versionSet = version, true
	return nil
}

func (s *Suite) restart() error {
	s.stop()
	return s.start()
}

func (s *Suite) unchangedVersion() error {
	if !s.versionSet {
		return fmt.Errorf("initial schema version was not captured")
	}

	version, err := s.healthVersion()
	if err != nil {
		return err
	}

	if version != s.version {
		return fmt.Errorf("schema version changed from %d to %d", s.version, version)
	}

	return nil
}

func (s *Suite) startupError() error {
	if err := s.awaitStartupExit(); err != nil {
		return err
	}

	s.process = nil
	s.logFile.Close()
	output, err := os.ReadFile(s.logFile.Name())
	if err != nil {
		return err
	}

	if !strings.Contains(strings.ToLower(string(output)), "database startup:") {
		return fmt.Errorf("missing understandable database error: %s", output)
	}

	return nil
}

func (s *Suite) awaitStartupExit() error {
	select {
	case err := <-s.exited:
		if err == nil {
			return fmt.Errorf("server exited successfully instead of reporting database failure")
		}

	case <-time.After(5 * time.Second):
		return fmt.Errorf("server did not exit after database startup error")
	}

	return nil
}

func (s *Suite) notReachable() error {
	client := &http.Client{Timeout: 200 * time.Millisecond}
	response, err := client.Get("http://" + s.address + "/health")
	if err != nil {
		return nil
	}

	response.Body.Close()
	return fmt.Errorf("server unexpectedly reachable: HTTP %d", response.StatusCode)
}

func (s *Suite) stop() {
	if s.process == nil {
		return
	}

	s.process.Process.Kill()
	<-s.exited
	s.logFile.Close()
	s.process = nil
}

func (s *Suite) cleanup() {
	s.stop()
	if s.store != nil {
		s.store.Close()
		s.store = nil
	}
}

func (s *Suite) afterScenario(ctx context.Context, _ *godog.Scenario, _ error) (context.Context, error) {
	s.cleanup()
	s.service = nil
	s.failure = nil
	s.versionSet = false
	s.retried = false
	return ctx, nil
}
