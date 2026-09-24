package codexcli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	portprofil "agentcontrolplane/app/internal/port/codexprofil"
)

// NewRunner bindet nur serverseitig konfigurierte, gepinnte Programme.
func NewRunner(config Config, policy Policy, brokers portprofil.BrokerFactory) *Runner {
	return &Runner{config: config, policy: policy, brokers: brokers, active: make(chan struct{}, 1)}
}

// Run prüft eine feste synthetische Runtime ohne Aufgabenlauf oder Clientbefehle.
func (r *Runner) Run(ctx context.Context, organizationID, agentID string) (Result, error) {
	result := Result{Code: unverifiedCode, Reason: "Die Codex-CLI-Grenze ist nicht vollständig nachgewiesen.", Checks: []Check{}}
	if r == nil || r.policy == nil || r.brokers == nil {
		result.fail("runtime_unconfigured", "Runtime-Prüfung ist nicht konfiguriert.")
		return result, nil
	}
	if !r.admit() {
		result.fail("runtime_busy", "Runtime-Prüfung ist bereits ausgelastet.")
		return result, nil
	}
	defer func() { <-r.active }()
	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	if err := r.policy.Authorize(ctx, organizationID, agentID, "runtime.probe"); err != nil {
		return result, err
	}

	return r.runAuthorized(ctx, organizationID, agentID, result), nil
}

func (r *Runner) admit() bool {
	select {
	case r.active <- struct{}{}:
		return true
	default:
		return false
	}
}

func (r *Runner) runAuthorized(ctx context.Context, organizationID, agentID string, result Result) Result {
	if err := r.config.verify(); err != nil {
		result.fail("runtime_binary_unverified", "Eine Runtime-Binärdatei ist nicht verifiziert.")
		return result
	}

	session, err := r.brokers.Start(ctx, organizationID, agentID)
	if err != nil {
		result.fail("action_broker_unavailable", "Der kontrollierte Aktionskanal ist nicht verfügbar.")
		return result
	}

	defer session.Close()
	return r.withSession(ctx, session, result)
}

func (r *Runner) withSession(ctx context.Context, session portprofil.Session, result Result) Result {
	scratch, err := r.scratch()
	if err != nil {
		result.fail("private_runtime_unavailable", "Die private Runtime ist nicht verfügbar.")
		return result
	}

	defer os.RemoveAll(scratch)
	if err := r.config.materialize(scratch); err != nil {
		result.fail("runtime_binary_unverified", "Eine Runtime-Binärdatei ist nicht verifiziert.")
		return result
	}

	if err := r.writeConfig(scratch); err != nil {
		result.fail("private_runtime_unavailable", "Die private Runtime ist nicht verfügbar.")
		return result
	}

	return r.execute(ctx, scratch, session.SocketPath(), result)
}

func (r *Runner) scratch() (string, error) {
	absolute, err := filepath.Abs(r.config.DatabasePath)
	if err != nil || absolute == "" {
		return "", errors.New("database path unavailable")
	}

	parent, err := filepath.EvalSymlinks(filepath.Dir(absolute))
	if err != nil {
		return "", fmt.Errorf("database parent: %w", err)
	}

	return os.MkdirTemp(parent, ".codex-probe-")
}

func (r *Runner) writeConfig(scratch string) error {
	for _, name := range []string{"home", "workspace", "tmp"} {
		if err := os.Mkdir(filepath.Join(scratch, name), 0700); err != nil {
			return err
		}
	}

	return os.WriteFile(filepath.Join(scratch, "home", "config.toml"), []byte(r.config.toml()), 0600)
}

func (r *Result) fail(code, detail string) {
	r.Checks = append(r.Checks, Check{Code: code, Status: "failed", Detail: detail})
}
