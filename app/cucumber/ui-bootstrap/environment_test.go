package uibootstrap

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

func NewSuite(t *testing.T) *Suite {
	return &Suite{t: t}
}

func (s *Suite) start() error {
	web, err := filepath.Abs("../../../ui/web")
	if err != nil {
		return err
	}

	build := exec.Command("npm", "run", "build")
	build.Dir = web
	if output, err := build.CombinedOutput(); err != nil {
		return fmt.Errorf("Vite build: %w: %s", err, output)
	}

	return s.startPreview(web)
}

func (s *Suite) startPreview(web string) error {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return err
	}

	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()
	s.baseURL = fmt.Sprintf("http://127.0.0.1:%d", port)
	s.server = exec.Command("node", filepath.Join(web, "node_modules/vite/bin/vite.js"), "preview", "--port", fmt.Sprint(port), "--strictPort")
	s.server.Dir = web
	s.server.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if err := s.server.Start(); err != nil {
		return err
	}

	return s.waitPreview()
}

func (s *Suite) waitPreview() error {
	client := &http.Client{Timeout: time.Second}
	for attempt := 0; attempt < 100; attempt++ {
		response, err := client.Get(s.baseURL)
		if err == nil {
			response.Body.Close()
			return s.startBrowser()
		}

		time.Sleep(50 * time.Millisecond)
	}

	return fmt.Errorf("Vite preview did not start at %s", s.baseURL)
}

func (s *Suite) startBrowser() error {
	script, err := filepath.Abs("browser.mjs")
	if err != nil {
		return err
	}

	s.browser = exec.Command("node", script)
	s.browser.Stderr = os.Stderr
	s.browser.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	return s.connectBrowser()
}

func (s *Suite) connectBrowser() error {
	input, err := s.browser.StdinPipe()
	if err != nil {
		return err
	}

	output, err := s.browser.StdoutPipe()
	if err != nil {
		return err
	}

	s.stdin = bufio.NewWriter(input)
	s.input = bufio.NewScanner(output)
	return s.browser.Start()
}

func (s *Suite) ask(request map[string]any) error {
	if err := s.send(request); err != nil {
		return err
	}

	return s.receive()
}

func (s *Suite) send(request map[string]any) error {
	encoded, err := json.Marshal(request)
	if err != nil {
		return err
	}

	if _, err := s.stdin.Write(append(encoded, '\n')); err != nil {
		return err
	}

	return s.stdin.Flush()
}

func (s *Suite) receive() error {
	if !s.input.Scan() {
		return fmt.Errorf("Chrome helper ended: %w", s.input.Err())
	}

	var reply Reply
	if err := json.Unmarshal(s.input.Bytes(), &reply); err != nil {
		return err
	}

	if !reply.OK {
		return fmt.Errorf("Chrome helper: %s", reply.Error)
	}

	s.page = reply.Page
	return nil
}

func (s *Suite) stop() {
	s.stopProcess(s.browser)
	s.stopProcess(s.server)
}

func (s *Suite) stopProcess(command *exec.Cmd) {
	if command == nil || command.Process == nil {
		return
	}

	syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
	command.Wait()
}
