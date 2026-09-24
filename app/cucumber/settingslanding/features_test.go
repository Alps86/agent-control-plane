package settingslanding

import (
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/cucumber/godog"
)

func TestFeatures(t *testing.T) {
	binary := filepath.Join(t.TempDir(), "app-server")
	command := exec.Command("go", "build", "-o", binary, "./cmd/server")
	command.Dir = "../.."
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("build public server: %v: %s", err, output)
	}
	suite := &Suite{t: t, binary: binary}
	runner := godog.TestSuite{ScenarioInitializer: suite.InitializeScenario, Options: &godog.Options{
		Format: "pretty", Paths: []string{"../../../features/ui/modelle-sprache/settings-landing.feature"},
		Strict: true, TestingT: t,
	}}
	if runner.Run() != 0 {
		t.Fatal("Settings landing acceptance failed")
	}
}
