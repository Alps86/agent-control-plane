package uipreview

import (
	"testing"

	"github.com/cucumber/godog"
)

func TestFeatures(t *testing.T) {
	suite := NewSuite(t)
	if err := suite.start(); err != nil {
		t.Fatal(err)
	}

	t.Cleanup(suite.stop)
	runner := godog.TestSuite{
		ScenarioInitializer: suite.initializeScenario,
		Options: &godog.Options{Format: "pretty", Paths: []string{
			"../../../features/ui/architektur/story-07.feature",
		}, TestingT: t},
	}
	if runner.Run() != 0 {
		t.Fatal("Godog-Szenarien fehlgeschlagen")
	}
}
