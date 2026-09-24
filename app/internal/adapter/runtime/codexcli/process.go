package codexcli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

func (r *Runner) execute(ctx context.Context, scratch, socket string, result Result) Result {
	if err := r.validSocket(socket); err != nil {
		result.fail("action_broker_unavailable", "Der kontrollierte Aktionskanal ist nicht verfügbar.")
		return result
	}

	command := exec.CommandContext(ctx, filepath.Join(scratch, "bwrap"), r.bwrapArgs(scratch, socket)...)
	command.Env = r.environment()
	output := &limitedOutput{limit: 1 << 20}
	command.Stdout = output
	err := command.Run()
	if err != nil {
		result.fail("cli_process_failed", "Der isolierte CLI-Prozess wurde nicht vollständig ausgeführt.")
		return result
	}

	return r.readReport(output.Bytes(), result)
}

func (o *limitedOutput) Write(data []byte) (int, error) {
	if o.Len()+len(data) > o.limit {
		return 0, io.ErrShortWrite
	}
	return o.Buffer.Write(data)
}

func (r *Runner) validSocket(path string) error {
	if filepath.Base(path) != "action.sock" {
		return errors.New("unexpected action socket")
	}

	info, err := os.Lstat(path)
	if err != nil || info.Mode()&os.ModeSocket == 0 {
		return errors.New("action socket missing")
	}

	return nil
}

func (r *Runner) bwrapArgs(scratch, socket string) []string {
	return []string{
		"--ro-bind", "/usr", "/usr", "--ro-bind", "/lib", "/lib",
		"--ro-bind", "/lib64", "/lib64", "--ro-bind", "/bin", "/bin",
		"--proc", "/proc", "--dev", "/dev", "--tmpfs", "/tmp",
		"--dir", "/opt", "--ro-bind", filepath.Join(scratch, "codex"), "/opt/codex",
		"--ro-bind", filepath.Join(scratch, "action-server"), "/opt/codex-action-server",
		"--dir", "/work", "--bind", scratch, "/work",
		"--dir", "/run", "--bind", filepath.Dir(socket), "/run/acp",
		"--chdir", "/work", "--unshare-user", "--unshare-pid", "--unshare-net",
		"--die-with-parent", "/opt/codex-action-server", "--probe-runtime",
	}
}

func (r *Runner) environment() []string {
	return []string{"HOME=/work/home", "CODEX_HOME=/work/home", "PATH=/usr/bin:/bin",
		"TMPDIR=/work/tmp", "ACP_CODEX_PROBE_SCENARIO=" + r.config.Scenario,
		"ACP_CODEX_ACTION_SOCKET=/run/acp/action.sock"}
}

func (r *Runner) readReport(output []byte, result Result) Result {
	var report childReport
	if json.Unmarshal(output, &report) != nil {
		result.fail("runtime_report_invalid", "Der isolierte CLI-Prozess lieferte keinen gültigen Nachweis.")
		return result
	}

	result.CLIStarted = report.Started
	result.ActionID = report.ActionID
	r.assess(&result, report)
	return result
}

func (r *Runner) assess(result *Result, report childReport) {
	result.check("host_files", report.HostFilesDenied, "Hostdateien sind im Namespace nicht erreichbar.")
	result.check("external_network", report.NetworkDenied, "Externes Netzwerk ist im Namespace gesperrt.")
	result.check("cli_started", report.Started, "Die echte Codex CLI wurde gestartet.")
	result.check("cli_completed", report.Completed && report.CLIError == "", "Der synthetische CLI-Turn wurde beendet.")
	result.check("tool_inventory", r.allowedTools(report.ToolNames), "Direkte Systemtools wurden nicht angeboten.")
	result.check("scenario_boundary", r.scenarioPassed(report), fmt.Sprintf("Feste Prüfaktion: status=%s, code=%s.", report.ActionStatus, report.ActionCode))
}

func (r *Result) check(code string, passed bool, detail string) {
	status := "failed"
	if passed {
		status = "passed"
	}

	r.Checks = append(r.Checks, Check{Code: code, Status: status, Detail: detail})
}

func (r *Runner) allowedTools(names []string) bool {
	expected := map[string]bool{"list_mcp_resources": true, "list_mcp_resource_templates": true,
		"read_mcp_resource": true, "request_user_input": true, "mcp__fach": true,
		"artifact_markdown_save": true}
	foundAction := false
	for _, name := range names {
		if !expected[name] {
			return false
		}

		foundAction = foundAction || name == "artifact_markdown_save"
	}

	return foundAction
}

func (r *Runner) scenarioPassed(report childReport) bool {
	if report.CLIError != "" || !report.Completed {
		return false
	}

	if r.config.Scenario == "action_allowed" {
		return report.ActionID == "artifact.markdown.save" && report.ActionStatus == "completed" && report.ActionCode == "artifact_saved"
	}

	if r.config.Scenario == "shell_denied" || r.config.Scenario == "http_denied" {
		return report.UnsupportedTool
	}

	return r.deniedAction(report)
}

func (r *Runner) deniedAction(report childReport) bool {
	if report.ActionID != "artifact.markdown.save" || report.ActionStatus != "failed" {
		return false
	}
	return report.ActionCode == "action_denied"
}
