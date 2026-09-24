package codexprofil

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cucumber/godog"
)

func (s *Suite) registerRuntime(sc *godog.ScenarioContext) {
	sc.Step(`^die Runtime-Prüfung verwendet eine kontrollierte synthetische Gegenstelle ohne echte Zugangsdaten$`, s.syntheticFixture)
	sc.Step(`^ich Kais Runtime-Prüfung mit einem Clientbefehl, Pfad oder URL anfordere$`, s.probeWithClientFields)
	sc.Step(`^antwortet die öffentliche API mit HTTP 400 ohne CLI-Start und ohne Dateiänderung$`, s.badProbeRequest)
	sc.Step(`^Kais Arbeitsbereich und vermitteltes Schreiben sind aktiviert$`, s.runtimeEnableWrite)
	sc.Step(`^die synthetische Gegenstelle fordert ausschließlich die fest benannte Go-Artefaktaktion an$`, s.actionAllowed)
	sc.Step(`^ich Kais Runtime-Prüfung mit einem leeren JSON-Objekt über die öffentliche API anfordere$`, s.probe)
	sc.Step(`^ich Kais Runtime-Prüfung erneut mit einem leeren JSON-Objekt über die öffentliche API anfordere$`, s.probe)
	sc.Step(`^nennt die Prüfantwort die erfolgreich ausgeführte benannte Fachaktion$`, s.actionProof)
	sc.Step(`^nennt die Prüfantwort erneut die erfolgreich ausgeführte benannte Fachaktion$`, s.actionProof)
	sc.Step(`^im privaten Arbeitsbereich von "Nordstern" und "Kai" liegt genau ein unverändertes Prüfartefakt$`, s.singlePrivateProof)
	sc.Step(`^außerhalb dieses Arbeitsbereichs entstand kein Prüfartefakt$`, s.noOutsideProof)
	sc.Step(`^wurde die echte Codex CLI innerhalb der bwrap-Grenze gestartet$`, s.cliStarted)
	sc.Step(`^die Prüfantwort nennt die ausgeführte benannte Fachaktion und ihren Nachweis$`, s.actionProof)
	sc.Step(`^nur im privaten Arbeitsbereich von "Nordstern" und "Kai" liegt das erwartete Artefakt$`, s.privateArtifact)
	sc.Step(`^der Prüfaufruf meldet keinen abgeschlossenen Aufgabenlauf$`, s.noTaskRun)
	sc.Step(`^Kais Arbeitsbereich ist aktiviert und Schreiben ist deaktiviert$`, s.runtimeDisableWrite)
	sc.Step(`^Kais Arbeitsbereich ist aktiviert und Schreiben ist (aktiviert|deaktiviert)$`, s.runtimeWriteSetting)
	sc.Step(`^die synthetische Gegenstelle fordert "([^"]+)" an$`, s.attackScenario)
	sc.Step(`^der Nebenweg "([^"]+)" wurde ohne Seitenwirkung verweigert$`, s.attackDenied)
	sc.Step(`^"([^"]+)" wurde ohne Seitenwirkung verweigert$`, s.attackDenied)
	sc.Step(`^die Prüfantwort enthält einen konkreten Nachweis der Verweigerung$`, s.denialProof)
	sc.Step(`^Kai bleibt ohne vollständigen Profilnachweis nicht startbereit$`, s.probeNotReady)
	sc.Step(`^der private Arbeitsbereich für Kai enthält einen Symlink nach außerhalb$`, s.runtimeSymlink)
	sc.Step(`^wird die Prüfung mit HTTP 409 ohne CLI-Start und ohne Dateiänderung verweigert$`, s.symlinkProbeDenied)
	sc.Step(`^für Kais Profil kann die tatsächliche CLI-Grenze nicht vollständig geprüft werden$`, s.incompleteProof)
	sc.Step(`^enthält die Prüfantwort "ready":false und "code":"codex_cli_enforcement_unverified"$`, s.unverifiedProof)
	sc.Step(`^Kais öffentliche Bereitschaft bleibt "ready":false$`, s.probeNotReady)
	sc.Step(`^ein Start von Kai bleibt mit HTTP 409 ohne Laufkennung gesperrt$`, s.runtimeStartDenied)
}

func (s *Suite) syntheticFixture() error {
	return s.prepareRuntimeBinaries()
}

func (s *Suite) probePath() string { return s.profilePath() + "/pruefen" }

func (s *Suite) probeWithClientFields() error {
	return s.request("POST", s.probePath(), `{"command":"touch outside","path":"../outside","url":"https://example.invalid"}`, "")
}

func (s *Suite) badProbeRequest() error {
	if s.response.Status != 400 || !strings.Contains(string(s.response.Body), "invalid_json") {
		return fmt.Errorf("Clientbefehle angenommen: HTTP %d %s", s.response.Status, s.response.Body)
	}
	if s.proofExists() {
		return fmt.Errorf("ungültiger Probeaufruf schrieb Artefakt")
	}
	return nil
}

func (s *Suite) runtimeEnableWrite() error {
	if err := s.putProfile(true, true); err != nil {
		return err
	}
	if s.response.Status != 200 {
		return fmt.Errorf("Schreibprofil: HTTP %d %s", s.response.Status, s.response.Body)
	}
	return nil
}

func (s *Suite) runtimeDisableWrite() error {
	if err := s.putProfile(true, false); err != nil {
		return err
	}
	if s.response.Status != 200 {
		return fmt.Errorf("Arbeitsprofil: HTTP %d %s", s.response.Status, s.response.Body)
	}
	return nil
}

func (s *Suite) runtimeWriteSetting(value string) error {
	if value == "aktiviert" {
		return s.runtimeEnableWrite()
	}
	return s.runtimeDisableWrite()
}

func (s *Suite) actionAllowed() error { return s.useRuntimeScenario("action_allowed") }

func (s *Suite) attackScenario(attack string) error {
	modes := map[string]string{
		"exec_command": "shell_denied", "externer HTTP-Aufruf": "http_denied",
		"fremde Organisation": "foreign_org", "fremder Agent": "foreign_agent",
		"Pfad mit ..": "traversal", "nicht freigegebenes Schreiben": "write_denied",
	}
	mode := modes[attack]
	if mode == "" {
		return fmt.Errorf("unbekannter Runtime-Angriff %q", attack)
	}
	return s.useRuntimeScenario(mode)
}

func (s *Suite) probe() error { return s.request("POST", s.probePath(), `{}`, "") }

func (s *Suite) runtimeResult() (RuntimeResult, error) {
	var result RuntimeResult
	if s.response.Status != 200 {
		return result, fmt.Errorf("Probe HTTP %d: %s", s.response.Status, s.response.Body)
	}
	err := json.Unmarshal(s.response.Body, &result)
	return result, err
}

func (s *Suite) cliStarted() error {
	result, err := s.runtimeResult()
	if err != nil {
		return err
	}
	if !result.CLIStarted || !s.passedCheck(result, "host_files") || !s.passedCheck(result, "external_network") {
		return fmt.Errorf("kein belegter isolierter CLI-Start: %s", s.response.Body)
	}
	return nil
}

func (s *Suite) passedCheck(result RuntimeResult, name string) bool {
	for _, check := range result.Checks {
		if check.Code == name && check.Status == "passed" {
			return true
		}
	}
	return false
}

func (s *Suite) actionProof() error {
	result, err := s.runtimeResult()
	if err != nil {
		return err
	}
	if result.ActionID != "artifact.markdown.save" || !s.passedCheck(result, "scenario_boundary") {
		return fmt.Errorf("benannte Fachaktion nicht belegt (runtime-proof.md vorhanden=%t): %s", s.proofExists(), s.response.Body)
	}
	return nil
}

func (s *Suite) proofPath() string {
	return filepath.Join(filepath.Dir(s.dbPath), "codex-workspaces", s.orgID, s.agentID, "runtime-proof.md")
}

func (s *Suite) proofExists() bool {
	_, err := os.Lstat(s.proofPath())
	return err == nil
}

func (s *Suite) privateArtifact() error {
	content, err := os.ReadFile(s.proofPath())
	if err != nil || string(content) != "SYNTHETIC_ARTIFACT" {
		return fmt.Errorf("privates Artefakt fehlt oder falsch: %v %q", err, content)
	}
	if _, err := os.Lstat(filepath.Join(filepath.Dir(s.dbPath), "runtime-proof.md")); err == nil {
		return fmt.Errorf("Artefakt außerhalb des Agentenbereichs")
	}
	return nil
}

func (s *Suite) singlePrivateProof() error {
	if err := s.privateArtifact(); err != nil {
		return err
	}
	items, err := os.ReadDir(filepath.Dir(s.proofPath()))
	if err != nil || len(items) != 1 || items[0].Name() != "runtime-proof.md" {
		return fmt.Errorf("privater Arbeitsbereich enthält unerwartete Dateien: %v %v", err, items)
	}
	return nil
}

func (s *Suite) noOutsideProof() error {
	root := filepath.Dir(s.dbPath)
	for _, path := range []string{filepath.Join(root, "runtime-proof.md"), filepath.Join(root, "codex-workspaces", s.orgID, "runtime-proof.md")} {
		if _, err := os.Lstat(path); err == nil {
			return fmt.Errorf("Prüfartefakt außerhalb des Agentenbereichs: %s", path)
		}
	}
	return nil
}

func (s *Suite) noTaskRun() error {
	content := string(s.response.Body)
	if strings.Contains(content, "run_id") || strings.Contains(content, "task_id") || strings.Contains(content, "completed_run") {
		return fmt.Errorf("Prüfung behauptet Aufgabenlauf: %s", content)
	}
	return nil
}

func (s *Suite) attackDenied(_ string) error {
	result, err := s.runtimeResult()
	if err != nil {
		return err
	}
	if s.proofExists() || !s.passedCheck(result, "scenario_boundary") {
		return fmt.Errorf("Angriff nicht ohne Schreibwirkung verweigert: %s", s.response.Body)
	}
	return nil
}

func (s *Suite) denialProof() error {
	result, err := s.runtimeResult()
	if err != nil {
		return err
	}
	if !s.passedCheck(result, "scenario_boundary") || len(result.Checks) == 0 {
		return fmt.Errorf("konkreter Verweigerungsbeleg fehlt: %s", s.response.Body)
	}
	return nil
}

func (s *Suite) probeNotReady() error {
	if err := s.readReadiness(); err != nil {
		return err
	}
	var readiness Readiness
	if err := json.Unmarshal(s.response.Body, &readiness); err != nil {
		return err
	}
	if readiness.Ready {
		return fmt.Errorf("Prüfung schaltete CLI frei: %s", s.response.Body)
	}
	return nil
}

func (s *Suite) runtimeSymlink() error {
	path := filepath.Join(filepath.Dir(s.proofPath()), "outside-link")
	if err := os.Symlink(s.t.TempDir(), path); err != nil {
		return err
	}
	return s.useRuntimeScenario("symlink")
}

func (s *Suite) symlinkProbeDenied() error {
	if s.response.Status != 409 || s.proofExists() {
		return fmt.Errorf("Symlink-Probe nicht gesperrt: HTTP %d %s", s.response.Status, s.response.Body)
	}
	return nil
}

func (s *Suite) incompleteProof() error { return s.useRuntimeScenario("shell_denied") }

func (s *Suite) unverifiedProof() error {
	result, err := s.runtimeResult()
	if err != nil {
		return err
	}
	if result.Ready || result.Code != "codex_cli_enforcement_unverified" {
		return fmt.Errorf("unbelegter CLI-Start freigegeben: %s", s.response.Body)
	}
	return nil
}

func (s *Suite) runtimeStartDenied() error {
	if err := s.startAgent(); err != nil {
		return err
	}
	return s.startDenied()
}
