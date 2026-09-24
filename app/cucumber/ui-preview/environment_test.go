package uipreview

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

func NewSuite(t *testing.T) *Suite {
	return &Suite{t: t}
}

func (s *Suite) start() error {
	if err := s.build("../../../ui"); err != nil {
		return err
	}

	if err := s.startServer("../../../ui"); err != nil {
		return err
	}

	return s.startBrowser()
}

func (s *Suite) build(directory string) error {
	command := exec.Command("go", "build", "-buildvcs=false", "-o", filepath.Join(s.t.TempDir(), "preview"), "./cmd/preview")
	command.Dir = directory
	output, err := command.CombinedOutput()
	if err != nil {
		return fmt.Errorf("preview build: %w: %s", err, output)
	}

	s.binary = command.Args[4]
	return nil
}

func (s *Suite) startServer(directory string) error {
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		return err
	}

	s.port = listener.Addr().(*net.TCPAddr).Port
	listener.Close()
	s.baseURL = "http://127.0.0.1:" + strconv.Itoa(s.port)
	s.server = exec.Command(s.binary)
	s.server.Dir = directory
	s.server.Env = append(os.Environ(), "ACP_PREVIEW_PORT="+strconv.Itoa(s.port))
	s.server.Stderr = os.Stderr
	if err := s.server.Start(); err != nil {
		return err
	}

	return s.waitReady()
}

func (s *Suite) waitReady() error {
	for attempt := 0; attempt < 100; attempt++ {
		response, err := http.Get(s.baseURL + "/?view=organization")
		if err == nil {
			response.Body.Close()
			return nil
		}

		time.Sleep(50 * time.Millisecond)
	}

	return fmt.Errorf("preview not reachable on %s", s.baseURL)
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
	s.stdout = bufio.NewScanner(output)
	return s.browser.Start()
}

func (s *Suite) stop() {
	if s.browser != nil && s.browser.Process != nil {
		syscall.Kill(-s.browser.Process.Pid, syscall.SIGKILL)
		s.browser.Wait()
	}

	s.stopServer()
}

func (s *Suite) fetch(path string, htmx bool) (Response, error) {
	request, err := http.NewRequest(http.MethodGet, s.baseURL+path, nil)
	if err != nil {
		return Response{}, err
	}

	if htmx {
		request.Header.Set("HX-Request", "true")
	}

	result, err := http.DefaultClient.Do(request)
	if err != nil {
		return Response{}, err
	}

	defer result.Body.Close()
	content, err := io.ReadAll(result.Body)
	return Response{Body: string(content), Status: result.StatusCode, Header: result.Header.Get("Content-Type")}, err
}

func (s *Suite) browserCommand(request map[string]any) (Page, error) {
	message, err := json.Marshal(request)
	if err != nil {
		return Page{}, err
	}

	if _, err := s.stdin.Write(append(message, '\n')); err != nil {
		return Page{}, err
	}

	if err := s.stdin.Flush(); err != nil {
		return Page{}, err
	}

	return s.readBrowser()
}

func (s *Suite) readBrowser() (Page, error) {
	if !s.stdout.Scan() {
		return Page{}, fmt.Errorf("Chrome ended: %w", s.stdout.Err())
	}

	var reply Reply
	if err := json.Unmarshal(s.stdout.Bytes(), &reply); err != nil {
		return Page{}, err
	}

	if !reply.OK {
		return Page{}, fmt.Errorf("Chrome: %s", reply.Error)
	}

	return reply.Page, nil
}

func (s *Suite) browse(path string) error {
	page, err := s.browserCommand(map[string]any{"url": s.baseURL + path})
	s.page = page
	return err
}

func (s *Suite) contains(value string) error {
	if !strings.Contains(s.page.Text, value) {
		return fmt.Errorf("page does not contain %q: %s", value, s.page.Text)
	}

	return nil
}
