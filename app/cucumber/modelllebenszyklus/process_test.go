package modelllebenszyklus

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	credentialadapter "agentcontrolplane/app/internal/adapter/credentials"
	"agentcontrolplane/app/internal/adapter/model/codexauth"
	"agentcontrolplane/app/internal/app/modellverbindung"
)

func (s *Suite) reopenFlow() error {
	store, err := credentialadapter.NewStore(s.secretPath, s.keyPath)
	if err != nil {
		return err
	}
	client, err := codexauth.NewDirect(codexauth.Config{Issuer: s.issuer.URL, ClientID: codexauth.CodexCLIClientID})
	if err != nil {
		return err
	}
	s.store = store
	s.flow = modellverbindung.New(client, store)
	return nil
}

func (s *Suite) restartProcess() error {
	if s.processBinary == "" {
		if err := s.buildProcess(); err != nil {
			return err
		}
	}
	s.stopProcess()
	if err := s.startProcess(); err != nil {
		return err
	}
	if err := s.awaitHealth(); err != nil {
		return err
	}
	s.stopProcess()
	if err := s.startProcess(); err != nil {
		return err
	}
	return s.awaitHealth()
}

func (s *Suite) buildProcess() error {
	s.processBinary = filepath.Join(s.dir, "story29-server")
	command := exec.Command("go", "build", "-o", s.processBinary, "./cmd/server")
	command.Dir = "../.."
	output, err := command.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Server-Build: %w: %s", err, output)
	}
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		return err
	}
	s.processAddress = listener.Addr().String()
	return listener.Close()
}

func (s *Suite) startProcess() error {
	logfile, err := os.CreateTemp(s.dir, "story29-log-")
	if err != nil {
		return err
	}
	command := exec.Command(s.processBinary)
	command.Env = append(os.Environ(), "APP_ADDR="+s.processAddress, "APP_DB_PATH="+filepath.Join(s.dir, "app.sqlite"), "APP_CREDENTIALS_PATH="+s.secretPath, "APP_CREDENTIAL_KEY_PATH="+s.keyPath)
	command.Env = append(command.Env, "HTTP_PROXY=http://127.0.0.1:1", "HTTPS_PROXY=http://127.0.0.1:1", "ALL_PROXY=http://127.0.0.1:1", "NO_PROXY=127.0.0.1,localhost")
	command.Stdout, command.Stderr = logfile, logfile
	if err := command.Start(); err != nil {
		logfile.Close()
		return err
	}
	s.process, s.processLog, s.processDone = command, logfile, make(chan error, 1)
	go func() { s.processDone <- command.Wait() }()
	return nil
}

func (s *Suite) awaitHealth() error {
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		response, err := http.Get("http://" + s.processAddress + "/health")
		if err == nil {
			response.Body.Close()
			if response.StatusCode == http.StatusOK {
				return nil
			}
		}
		time.Sleep(25 * time.Millisecond)
	}
	logs, _ := os.ReadFile(s.processLog.Name())
	return fmt.Errorf("Serverneustart fehlgeschlagen: %s", logs)
}

func (s *Suite) stopProcess() {
	if s.process == nil {
		return
	}
	_ = s.process.Process.Kill()
	<-s.processDone
	_ = s.processLog.Close()
	s.process = nil
}

func (s *Suite) processCodexView() error {
	return s.processGet("/settings/modelle/codex/device/status")
}

func (s *Suite) processGet(path string) error {
	response, err := http.Get("http://" + s.processAddress + path)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	s.status = response.StatusCode
	s.response, err = io.ReadAll(io.LimitReader(response.Body, 1<<20))
	if err != nil {
		return err
	}
	if s.status != http.StatusOK {
		return fmt.Errorf("Prozess-HTTP %d für %s", s.status, path)
	}
	return nil
}

func (s *Suite) processAccess() error {
	_, _, err := s.flow.Access(context.Background())
	return err
}
