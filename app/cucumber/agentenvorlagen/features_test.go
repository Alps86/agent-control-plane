package agentenvorlagen

import (
	"testing"

	"github.com/cucumber/godog"
)

func TestFeatures(t *testing.T) {
	suite := NewSuite(t)
	t.Cleanup(suite.cleanup)
	if err := suite.build(); err != nil {
		t.Fatal(err)
	}

	runner := godog.TestSuite{ScenarioInitializer: suite.initializeScenario,
		Options: &godog.Options{Format: "pretty", Paths: []string{
			"../../../features/app/agenten/story-15.feature",
			"../../../features/ui/agenten/story-15.feature",
		}, TestingT: t}}
	if runner.Run() != 0 {
		t.Fatal("AGT-01-Godog-Szenarien fehlgeschlagen")
	}
}
