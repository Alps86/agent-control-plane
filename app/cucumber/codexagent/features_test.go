package codexagent

import (
	"github.com/cucumber/godog"
	"testing"
)

func TestFeatures(t *testing.T) {
	suite := NewSuite(t)
	t.Cleanup(suite.cleanup)
	if err := suite.build(); err != nil {
		t.Fatal(err)
	}
	runner := godog.TestSuite{ScenarioInitializer: suite.initialize, Options: &godog.Options{
		Format: "pretty", Paths: []string{
			"../../../features/app/agenten/codex-cli.feature",
			"../../../features/ui/agenten/codex-cli.feature",
		}, TestingT: t,
	}}
	if runner.Run() != 0 {
		t.Fatal("Codex-CLI-Godog-Szenarien fehlgeschlagen")
	}
}
