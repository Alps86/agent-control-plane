package codexprofil

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	runtimecli "agentcontrolplane/app/internal/adapter/runtime/codexcli"
	runtimeprofil "agentcontrolplane/app/internal/adapter/runtime/codexprofil"
	appcodexprofil "agentcontrolplane/app/internal/app/codexprofil"
)

func (s *Suite) runtimeRunner(profiles *appcodexprofil.Service) (*runtimecli.Runner, error) {
	if s.runtimeScenario == "" {
		return nil, nil
	}
	if err := s.prepareRuntimeBinaries(); err != nil {
		return nil, err
	}
	config := runtimecli.Config{
		DatabasePath: s.dbPath, CLIPath: s.cliPath, CLISHA256: s.cliSHA,
		BwrapPath: s.bwrapPath, BwrapSHA256: s.bwrapSHA,
		ActionServerPath: s.actionServerPath, ActionServerSHA256: s.actionServerSHA,
		Scenario: s.runtimeScenario,
	}
	return runtimecli.NewRunner(config, profiles, runtimeprofil.NewBrokerFactory(profiles)), nil
}

func (s *Suite) prepareRuntimeBinaries() error {
	if s.actionServerPath != "" {
		return nil
	}
	cli, err := s.pinnedExecutable("codex")
	if err != nil {
		return err
	}
	bwrap, err := s.pinnedExecutable("bwrap")
	if err != nil {
		return err
	}
	s.cliPath, s.bwrapPath = cli, bwrap
	if err := s.buildActionServer(); err != nil {
		return err
	}
	s.cliSHA, err = s.digest(cli)
	if err != nil {
		return err
	}
	s.bwrapSHA, err = s.digest(bwrap)
	if err != nil {
		return err
	}
	s.actionServerSHA, err = s.digest(s.actionServerPath)
	return err
}

func (s *Suite) pinnedExecutable(name string) (string, error) {
	path, err := exec.LookPath(name)
	if err != nil {
		return "", err
	}
	absolute, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	return filepath.EvalSymlinks(absolute)
}

func (s *Suite) buildActionServer() error {
	s.actionServerPath = filepath.Join(s.t.TempDir(), "codex-action-server")
	cmd := exec.Command("go", "build", "-o", s.actionServerPath, "./cmd/codex-action-server")
	cmd.Dir = "../.."
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Action-Server-Build: %w: %s", err, output)
	}
	return nil
}

func (s *Suite) digest(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func (s *Suite) useRuntimeScenario(name string) error {
	s.runtimeScenario = name
	return s.restart()
}
