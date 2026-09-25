package projektort

import (
	"bufio"
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
	"syscall"
	"testing"
	"time"

	"github.com/cucumber/godog"
)

type suite struct {
	t                       *testing.T
	binary, dbPath, address string
	process                 *exec.Cmd
	exited                  chan error
	log                     *os.File
	client                  *http.Client
	response                result
	organizations, projects map[string]string
	lastProject, lastOrg    string
	browser                 *exec.Cmd
	input                   *bufio.Writer
	output                  *bufio.Scanner
	page                    browserPage
}

type result struct {
	status   int
	body     []byte
	location string
}
type record struct {
	ID string `json:"id"`
}
type browserPage struct {
	URL       string `json:"url"`
	Heading   string `json:"heading"`
	Text      string `json:"text"`
	Alert     string `json:"alert"`
	PathInput bool   `json:"pathInput"`
	Enabled   bool   `json:"enabled"`
}
type browserReply struct {
	OK    bool        `json:"ok"`
	Error string      `json:"error"`
	Page  browserPage `json:"page"`
}

func newSuite(t *testing.T) *suite {
	return &suite{t: t, client: &http.Client{Timeout: 6 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}, organizations: map[string]string{}, projects: map[string]string{}}
}

func (s *suite) build() error {
	s.binary = filepath.Join(s.t.TempDir(), "server")
	cmd := exec.Command("go", "build", "-o", s.binary, "./cmd/server")
	cmd.Dir = "../.."
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Server bauen: %w: %s", err, out)
	}
	return nil
}

func (s *suite) fresh() error {
	s.close()
	s.dbPath = filepath.Join(s.t.TempDir(), "private", "app.sqlite")
	if err := os.MkdirAll(filepath.Dir(s.dbPath), 0700); err != nil {
		return err
	}
	s.organizations = map[string]string{}
	s.projects = map[string]string{}
	return s.start()
}

func (s *suite) start() error {
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		return err
	}
	s.address = listener.Addr().String()
	listener.Close()
	s.log, err = os.CreateTemp(s.t.TempDir(), "projektort-server-")
	if err != nil {
		return err
	}
	s.process = exec.Command(s.binary)
	s.process.Env = append(os.Environ(), "APP_ADDR="+s.address, "APP_DB_PATH="+s.dbPath)
	s.process.Stdout, s.process.Stderr = s.log, s.log
	if err := s.process.Start(); err != nil {
		return err
	}
	s.exited = make(chan error, 1)
	go func() { s.exited <- s.process.Wait() }()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		res, err := s.client.Get(s.url("/health"))
		if err == nil {
			res.Body.Close()
			if res.StatusCode == http.StatusOK {
				return nil
			}
		}
		time.Sleep(25 * time.Millisecond)
	}
	logs, _ := os.ReadFile(s.log.Name())
	return fmt.Errorf("Server nicht erreichbar: %s", logs)
}

func (s *suite) restart() error { s.stopServer(); return s.start() }
func (s *suite) stopServer() {
	if s.process != nil {
		_ = s.process.Process.Kill()
		<-s.exited
		s.process = nil
	}
	if s.log != nil {
		_ = s.log.Close()
		s.log = nil
	}
}
func (s *suite) close() { s.stopBrowser(); s.stopServer() }
func (s *suite) after(ctx context.Context, _ *godog.Scenario, _ error) (context.Context, error) {
	s.close()
	s.response = result{}
	s.page = browserPage{}
	s.lastProject, s.lastOrg = "", ""
	return ctx, nil
}
func (s *suite) url(path string) string { return "http://" + s.address + path }
func (s *suite) call(method, path string, payload any) error {
	var body io.Reader
	if payload != nil {
		raw, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		body = strings.NewReader(string(raw))
	}
	req, err := http.NewRequest(method, s.url(path), body)
	if err != nil {
		return err
	}
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	res, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	raw, err := io.ReadAll(res.Body)
	s.response = result{status: res.StatusCode, body: raw, location: res.Header.Get("Location")}
	return err
}
func (s *suite) key(org, project string) string { return org + "\x00" + project }
func (s *suite) path(org, project string) string {
	return "/api/organisationen/" + s.organizations[org] + "/projekte/" + s.projects[s.key(org, project)] + "/ausfuehrungsort"
}
func (s *suite) pagePath(org, project string) string {
	return strings.TrimPrefix(s.path(org, project), "/api")
}
func (s *suite) status(expected int) error {
	if s.response.status != expected {
		return fmt.Errorf("HTTP %d statt %d: %s", s.response.status, expected, s.response.body)
	}
	return nil
}

func (s *suite) startBrowser() error {
	script, err := filepath.Abs("browser.mjs")
	if err != nil {
		return err
	}
	s.browser = exec.Command("node", script)
	s.browser.Stderr = os.Stderr
	s.browser.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	in, err := s.browser.StdinPipe()
	if err != nil {
		return err
	}
	out, err := s.browser.StdoutPipe()
	if err != nil {
		return err
	}
	s.input, s.output = bufio.NewWriter(in), bufio.NewScanner(out)
	return s.browser.Start()
}
func (s *suite) stopBrowser() {
	if s.browser != nil && s.browser.Process != nil {
		_ = syscall.Kill(-s.browser.Process.Pid, syscall.SIGKILL)
		_ = s.browser.Wait()
		s.browser = nil
	}
}
func (s *suite) browserCommand(command map[string]any) error {
	raw, err := json.Marshal(command)
	if err != nil {
		return err
	}
	if _, err = s.input.Write(append(raw, '\n')); err != nil {
		return err
	}
	if err = s.input.Flush(); err != nil {
		return err
	}
	if !s.output.Scan() {
		return fmt.Errorf("Chrome beendet: %v", s.output.Err())
	}
	var reply browserReply
	if err = json.Unmarshal(s.output.Bytes(), &reply); err != nil {
		return err
	}
	if !reply.OK {
		return fmt.Errorf("Chrome: %s", reply.Error)
	}
	s.page = reply.Page
	return nil
}
