package kommentare

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
	runner := godog.TestSuite{ScenarioInitializer: suite.initialize,
		Options: &godog.Options{Format: "pretty", Paths: []string{
			"../../../features/app/aufgaben/story-42.feature",
			"../../../features/ui/aufgaben/story-42.feature",
		}, TestingT: t}}
	if runner.Run() != 0 {
		t.Fatal("TASK-05-Godog-Szenarien fehlgeschlagen")
	}
}
