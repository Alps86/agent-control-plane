package prozess

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

func (s *Suite) start() error {
	directory := s.t.TempDir()
	s.binary = filepath.Join(directory, "server")
	s.dbPath = filepath.Join(directory, "app.sqlite")
	s.secretPath = filepath.Join(directory, "credentials", "secrets.enc")
	s.keyPath = filepath.Join(directory, "master", "master.key")
	if err := s.chooseAddress(); err != nil {
		return err
	}
	s.startProxy()

	if err := s.build(); err != nil {
		return err
	}

	if err := s.launch(); err != nil {
		return err
	}

	return s.waitHealthy()
}

func (s *Suite) chooseAddress() error {
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		return err
	}

	s.address = listener.Addr().String()
	return listener.Close()
}

func (s *Suite) startProxy() {
	s.proxy = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		s.proxyCalls.Add(1)
		http.Error(w, "external provider access blocked by process test", http.StatusBadGateway)
	}))
}

func (s *Suite) build() error {
	command := exec.Command("go", "build", "-o", s.binary, "./cmd/server")
	command.Dir = "../../.."
	output, err := command.CombinedOutput()
	if err != nil {
		return fmt.Errorf("lokalen Server bauen: %w: %s", err, output)
	}

	return nil
}

func (s *Suite) launch() error {
	logFile, err := os.CreateTemp(s.t.TempDir(), "story23-server-log-")
	if err != nil {
		return err
	}

	command := exec.Command(s.binary)
	command.Env = s.serverEnvironment()
	command.Stdout, command.Stderr = logFile, logFile
	if err := command.Start(); err != nil {
		logFile.Close()
		return err
	}

	s.logFile = logFile
	s.server = command
	s.exited = make(chan error, 1)
	go s.waitProcess()
	return nil
}

func (s *Suite) serverEnvironment() []string {
	return append(os.Environ(),
		"APP_ADDR="+s.address,
		"APP_DB_PATH="+s.dbPath,
		"APP_CREDENTIALS_PATH="+s.secretPath,
		"APP_CREDENTIAL_KEY_PATH="+s.keyPath,
		"HTTP_PROXY="+s.proxy.URL,
		"HTTPS_PROXY="+s.proxy.URL,
		"ALL_PROXY="+s.proxy.URL,
		"NO_PROXY=127.0.0.1,localhost",
	)
}

func (s *Suite) waitProcess() {
	s.exited <- s.server.Wait()
}

func (s *Suite) waitHealthy() error {
	deadline := time.Now().Add(5 * time.Second)
	var lastErr error
	for time.Now().Before(deadline) {
		lastErr = s.get("/health")
		if lastErr == nil && s.status == 200 {
			return nil
		}

		time.Sleep(25 * time.Millisecond)
	}

	logs, _ := os.ReadFile(s.logFile.Name())
	return fmt.Errorf("gebauter Server %s: letzter HTTP-Fehler %v, Status %d, Body %q, Logs %q", s.address, lastErr, s.status, s.body, logs)
}

func (s *Suite) stop() {
	if s.server == nil {
		return
	}

	s.server.Process.Kill()
	<-s.exited
	s.logFile.Close()
	s.server = nil
}

func (s *Suite) restart() error {
	s.stop()
	if err := s.launch(); err != nil {
		return err
	}

	return s.waitHealthy()
}
