package ops01

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

func NewSuite(t *testing.T) *Suite {
	return &Suite{t: t}
}

func (s *Suite) InitializeScenario(sc *godog.ScenarioContext) {
	sc.After(s.afterScenario)
	sc.Step(`^ein neuer beschreibbarer SQLite-Datenpfad ist über APP_DB_PATH konfiguriert$`, s.newPath)
	sc.Step(`^eine vorhandene unzugängliche SQLite-Datei ist über APP_DB_PATH konfiguriert$`, s.inaccessibleFile)
	sc.Step(`^ich den lokalen Go-Server starte$`, s.start)
	sc.Step(`^antwortet seine öffentliche HTTP-Statusroute erfolgreich mit einer Schemaversion$`, s.healthy)
	sc.Step(`^am konfigurierten Pfad liegt eine SQLite-Datenbank$`, s.databaseExists)
	sc.Step(`^ich die Schemaversion über HTTP festhalte$`, s.captureVersion)
	sc.Step(`^ich den Server beende und mit demselben APP_DB_PATH erneut starte$`, s.restart)
	s.initializeFailureSteps(sc)
}

func (s *Suite) initializeFailureSteps(sc *godog.ScenarioContext) {
	sc.Step(`^die Schemaversion ist unverändert$`, s.versionUnchanged)
	sc.Step(`^beendet sich der Server mit einem sichtbaren Datenbank-Startfehler$`, s.startupFailed)
	sc.Step(`^seine öffentliche HTTP-Statusroute ist nicht erreichbar$`, s.notReachable)
	sc.Step(`^die vorhandene SQLite-Datei wurde nicht ersetzt$`, s.fileUnchanged)
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

func (s *Suite) newPath() error {
	s.dbPath = filepath.Join(s.t.TempDir(), "app.sqlite")
	return nil
}

func (s *Suite) inaccessibleFile() error {
	if err := s.newPath(); err != nil {
		return err
	}

	if err := s.start(); err != nil {
		return err
	}

	if err := s.waitHealthy(); err != nil {
		return err
	}

	s.stop()
	return s.protectExistingFile()
}

func (s *Suite) protectExistingFile() error {
	info, err := os.Stat(s.dbPath)
	if err != nil {
		return err
	}

	s.fileInfo = info
	return os.Chmod(s.dbPath, 0000)
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

func (s *Suite) health() (int, error) {
	client := &http.Client{Timeout: 200 * time.Millisecond}
	response, err := client.Get("http://" + s.address + "/health")
	if err != nil {
		return 0, err
	}

	defer response.Body.Close()
	if response.StatusCode != http.StatusOK || !strings.HasPrefix(response.Header.Get("Content-Type"), "application/json") {
		return 0, fmt.Errorf("HTTP-Status %d, Content-Type %q", response.StatusCode, response.Header.Get("Content-Type"))
	}

	var body Health
	err = json.NewDecoder(response.Body).Decode(&body)
	return body.SchemaVersion, err
}

func (s *Suite) waitHealthy() error {
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if version, err := s.health(); err == nil && version > 0 {
			return nil
		}

		time.Sleep(20 * time.Millisecond)
	}

	return fmt.Errorf("Server unter %s nicht mit Schemaversion erreichbar", s.address)
}

func (s *Suite) healthy() error {
	return s.waitHealthy()
}

func (s *Suite) databaseExists() error {
	info, err := os.Stat(s.dbPath)
	if err != nil {
		return err
	}

	if !info.Mode().IsRegular() || info.Size() == 0 {
		return fmt.Errorf("keine SQLite-Datenbank am konfigurierten Pfad")
	}

	return nil
}

func (s *Suite) captureVersion() error {
	version, err := s.health()
	if err != nil {
		return err
	}

	if version < 1 {
		return fmt.Errorf("ungültige Schemaversion %d", version)
	}

	s.version = version
	return nil
}

func (s *Suite) restart() error {
	s.stop()
	return s.start()
}

func (s *Suite) versionUnchanged() error {
	version, err := s.health()
	if err != nil {
		return err
	}

	if version != s.version {
		return fmt.Errorf("Schemaversion änderte sich von %d auf %d", s.version, version)
	}

	return nil
}

func (s *Suite) startupFailed() error {
	if err := s.waitForExit(); err != nil {
		return err
	}

	s.process = nil
	s.logFile.Close()
	output, err := os.ReadFile(s.logFile.Name())
	if err != nil {
		return err
	}

	s.startupLog = string(output)
	if !strings.Contains(strings.ToLower(s.startupLog), "database startup:") {
		return fmt.Errorf("kein sichtbarer Datenbank-Startfehler: %s", s.startupLog)
	}

	return nil
}

func (s *Suite) waitForExit() error {
	select {
	case err := <-s.exited:
		if err == nil {
			return fmt.Errorf("Server beendete sich ohne Startfehler")
		}

	case <-time.After(5 * time.Second):
		return fmt.Errorf("Server beendete sich nach Datenbankfehler nicht")
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
	return fmt.Errorf("HTTP-Statusroute trotz Startfehler erreichbar: %d", response.StatusCode)
}

func (s *Suite) fileUnchanged() error {
	info, err := os.Stat(s.dbPath)
	if err != nil {
		return err
	}

	if !os.SameFile(info, s.fileInfo) || info.Size() != s.fileInfo.Size() || info.Mode().Perm() != 0000 {
		return fmt.Errorf("vorhandene Datenbank wurde ersetzt oder verändert")
	}

	return nil
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
	if s.dbPath != "" {
		os.Chmod(s.dbPath, 0600)
	}
}

func (s *Suite) afterScenario(ctx context.Context, _ *godog.Scenario, _ error) (context.Context, error) {
	s.cleanup()
	s.dbPath, s.address, s.version, s.fileInfo, s.startupLog = "", "", 0, nil, ""
	return ctx, nil
}
