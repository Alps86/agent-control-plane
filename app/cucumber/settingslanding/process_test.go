package settingslanding

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/cucumber/godog"
)

func (s *Suite) startLocal() error {
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		return err
	}
	address := listener.Addr().String()
	listener.Close()
	s.baseURL = "http://" + address
	dbPath := filepath.Join(s.t.TempDir(), "settings.db")
	s.process = exec.Command(s.binary)
	s.process.Env = append(os.Environ(), "APP_ADDR="+address, "APP_DB_PATH="+dbPath)
	if err := s.process.Start(); err != nil {
		return err
	}
	s.exited = make(chan error, 1)
	process := s.process
	go func() { s.exited <- process.Wait() }()
	return s.waitReady()
}

func (s *Suite) waitReady() error {
	client := &http.Client{Timeout: 300 * time.Millisecond}
	for attempt := 0; attempt < 100; attempt++ {
		response, err := client.Get(s.baseURL + "/health")
		if err == nil {
			response.Body.Close()
			if response.StatusCode == http.StatusOK {
				return nil
			}
		}
		select {
		case err := <-s.exited:
			s.process = nil
			return fmt.Errorf("server exited before ready: %v", err)
		default:
		}
		time.Sleep(30 * time.Millisecond)
	}
	return fmt.Errorf("server not healthy at %s", s.baseURL)
}

func (s *Suite) stop() {
	if s.process == nil {
		return
	}
	_ = s.process.Process.Kill()
	<-s.exited
	s.process = nil
}

func (s *Suite) cleanup(ctx context.Context, _ *godog.Scenario, _ error) (context.Context, error) {
	s.stopBrowser()
	s.stop()
	s.status, s.contentType, s.body, s.baseURL = 0, "", "", ""
	s.browser, s.unsafeFailed, s.unsafeOutput = browserResult{}, false, ""
	return ctx, nil
}

func (s *Suite) unsafeAddress(address string) error {
	if address != "0.0.0.0:0" {
		return fmt.Errorf("unexpected unsafe bind %q", address)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, s.binary)
	command.Env = append(os.Environ(), "APP_ADDR="+address, "APP_DB_PATH="+filepath.Join(s.t.TempDir(), "unsafe.db"))
	output, err := command.CombinedOutput()
	if ctx.Err() != nil {
		return fmt.Errorf("unsafe server remained running")
	}
	s.unsafeFailed = err != nil
	s.unsafeOutput = string(output)
	return nil
}

func (s *Suite) startupFinished() error { return nil }

func (s *Suite) noPublicListener() error {
	if !s.unsafeFailed {
		return fmt.Errorf("unsafe bind did not fail startup")
	}
	if !strings.Contains(s.unsafeOutput, "APP_ADDR must be a local loopback address") {
		return fmt.Errorf("startup failed without unsafe-bind rejection")
	}
	return nil
}
