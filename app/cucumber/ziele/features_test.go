package ziele

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
			"../../../features/app/ziele/story-13.feature",
			"../../../features/ui/ziele/story-13.feature",
			"../../../features/app/ziele/story-32.feature",
			"../../../features/ui/ziele/story-32.feature",
		}, TestingT: t}}
	if runner.Run() != 0 {
		t.Fatal("GOAL-01-Godog-Szenarien fehlgeschlagen")
	}
}
