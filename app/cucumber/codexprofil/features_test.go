package codexprofil

import (
	"os"
	"testing"

	"github.com/cucumber/godog"
)

func TestFeatures(t *testing.T) {
	suite := NewSuite(t)
	t.Cleanup(suite.cleanup)
	tags := "~@pending-real-run"
	if selected := os.Getenv("ACP_STORY21_TAGS"); selected != "" {
		tags = selected
	}
	runner := godog.TestSuite{ScenarioInitializer: suite.initialize, Options: &godog.Options{
		Format: "pretty", Paths: []string{
			"../../../features/app/berechtigungen/story-21.feature",
			"../../../features/ui/berechtigungen/story-21.feature",
			"../../../features/app/berechtigungen/story-21-runtime.feature",
		}, Tags: tags, TestingT: t,
	}}
	if runner.Run() != 0 {
		t.Fatal("Story-21-Godog-Szenarien fehlgeschlagen")
	}
}
