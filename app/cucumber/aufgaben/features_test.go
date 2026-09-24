package aufgaben

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

	runner := godog.TestSuite{ScenarioInitializer: suite.initialize,
		Options: &godog.Options{Format: "pretty", Paths: []string{
			"../../../features/app/aufgaben/story-17.feature",
			"../../../features/ui/aufgaben/story-17.feature",
		}, TestingT: t}}
	if runner.Run() != 0 {
		t.Fatal("TASK-01-Godog-Szenarien fehlgeschlagen")
	}
}
