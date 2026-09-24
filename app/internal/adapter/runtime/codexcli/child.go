package codexcli

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

// NewChild erzeugt den festen Prüfprozess innerhalb der privaten OS-Grenze.
func NewChild() *Child {
	return &Child{}
}

// Run startet nur den synthetischen Modellendpunkt und die gepinnte CLI.
func (c *Child) Run() error {
	report := childReport{HostFilesDenied: c.hostFilesDenied(), NetworkDenied: c.networkDenied()}
	fixture := &fixture{scenario: os.Getenv("ACP_CODEX_PROBE_SCENARIO")}
	listener, err := net.Listen("tcp", "127.0.0.1:18761")
	if err != nil {
		return err
	}

	server := &http.Server{Handler: fixture}
	go server.Serve(listener)
	defer server.Close()
	c.runCLI(&report)
	report.ToolNames = fixture.names()
	return json.NewEncoder(os.Stdout).Encode(report)
}

func (c *Child) hostFilesDenied() bool {
	_, hostErr := os.Stat("/home")
	_, systemErr := os.Stat("/etc/passwd")
	return errors.Is(hostErr, os.ErrNotExist) && errors.Is(systemErr, os.ErrNotExist)
}

func (c *Child) networkDenied() bool {
	conn, err := net.DialTimeout("tcp", "1.1.1.1:443", 200*time.Millisecond)
	if err != nil {
		return true
	}

	conn.Close()
	return false
}

func (c *Child) runCLI(report *childReport) {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "/opt/codex", "exec", "--strict-config", "--skip-git-repo-check", "--ephemeral", "-C", "/work/workspace", "--json", "Synthetic fixed runtime probe.")
	cmd.Env = []string{"HOME=/work/home", "CODEX_HOME=/work/home", "PATH=/usr/bin:/bin", "TMPDIR=/work/tmp", "ACP_CODEX_ACTION_SOCKET=/run/acp/action.sock"}
	cmd.Stdin = strings.NewReader("")
	stdout, stderr := &limitedOutput{limit: 1 << 20}, &limitedOutput{limit: 1 << 20}
	cmd.Stdout, cmd.Stderr = stdout, stderr
	err := cmd.Run()
	c.inspectOutput(report, append(stdout.Bytes(), stderr.Bytes()...))
	if err != nil {
		report.CLIError = "cli_process_failed"
	}
}

func (c *Child) inspectOutput(report *childReport, output []byte) {
	report.UnsupportedTool = bytes.Contains(output, []byte("unsupported call:"))
	scanner := bufio.NewScanner(bytes.NewReader(output))
	scanner.Buffer(make([]byte, 4096), 1048576)
	for scanner.Scan() {
		c.inspectLine(report, scanner.Bytes())
	}
}

func (c *Child) inspectLine(report *childReport, line []byte) {
	var event map[string]any
	if json.Unmarshal(line, &event) != nil {
		return
	}

	c.inspectEvent(report, event)
}

func (c *Child) inspectEvent(report *childReport, event map[string]any) {
	kind, _ := event["type"].(string)
	if kind == "thread.started" {
		report.Started = true
	}

	if kind == "turn.completed" {
		report.Completed = true
	}

	item, _ := event["item"].(map[string]any)
	c.inspectItem(report, item)
}

func (c *Child) inspectItem(report *childReport, item map[string]any) {
	if item["type"] != "mcp_tool_call" || item["status"] == "in_progress" {
		return
	}

	report.ActionID, _ = item["tool"].(string)
	report.ActionStatus, _ = item["status"].(string)
	result, _ := item["result"].(map[string]any)
	contents, _ := result["content"].([]any)
	if len(contents) == 0 {
		return
	}

	content, _ := contents[0].(map[string]any)
	report.ActionCode, _ = content["text"].(string)
}
