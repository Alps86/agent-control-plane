package projekte

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
			"../../../features/app/projekte/story-14.feature",
			"../../../features/ui/projekte/story-14.feature",
		}, TestingT: t}}
	if runner.Run() != 0 {
		t.Fatal("PROJ-01-Godog-Szenarien fehlgeschlagen")
	}
}
