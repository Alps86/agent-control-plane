package cucumber

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/cucumber/godog"
)

// NewSuite erzeugt die Blackbox-Testumgebung.
func NewSuite(t *testing.T) *Suite {
	return &Suite{t: t}
}

func (s *Suite) InitializeScenario(sc *godog.ScenarioContext) {
	sc.Step(`^der Go-Server ist auf einem freien lokalen Port gestartet$`, s.startServer)
	sc.Step(`^ich über HTTP GET "([^"]*)" an den Server sende$`, s.get)
	sc.Step(`^antwortet der Server mit dem HTTP-Status (\d+)$`, s.status)
}

func (s *Suite) startServer() error {
	binary := filepath.Join(s.t.TempDir(), "server")
	build := exec.Command("go", "build", "-o", binary, "./cmd/server")
	build.Dir = ".."
	if output, err := build.CombinedOutput(); err != nil {
		return fmt.Errorf("server build: %w: %s", err, output)
	}

	return s.launch(binary)
}

func (s *Suite) launch(binary string) error {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return err
	}

	address := listener.Addr().String()
	listener.Close()
	s.baseURL = "http://" + address
	s.process = exec.Command(binary)
	s.process.Env = append(os.Environ(), "APP_ADDR="+address)
	if err := s.process.Start(); err != nil {
		return err
	}

	return s.waitForServer(address)
}

func (s *Suite) waitForServer(address string) error {
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		connection, err := net.DialTimeout("tcp", address, 100*time.Millisecond)
		if err == nil {
			connection.Close()
			return nil
		}

		time.Sleep(20 * time.Millisecond)
	}

	return fmt.Errorf("server did not listen on %s", address)
}

func (s *Suite) get(path string) error {
	client := &http.Client{Timeout: 2 * time.Second}
	response, err := client.Get(s.baseURL + path)
	if err != nil {
		return err
	}

	s.response = response
	return nil
}

func (s *Suite) status(expected int) error {
	defer s.response.Body.Close()
	if s.response.StatusCode != expected {
		return fmt.Errorf("HTTP status: got %d, want %d", s.response.StatusCode, expected)
	}

	return nil
}

func (s *Suite) stop() {
	if s.process == nil {
		return
	}

	s.process.Process.Kill()
	s.process.Wait()
}
