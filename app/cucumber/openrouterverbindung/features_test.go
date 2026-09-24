package openrouterverbindung

import (
	"testing"

	"github.com/cucumber/godog"
)

func TestFeatures(t *testing.T) {
	suite := NewSuite(t)
	runner := godog.TestSuite{
		ScenarioInitializer: suite.InitializeScenario,
		Options: &godog.Options{
			Format: "pretty", Paths: []string{
				"../../../features/app/modelle-sprache/story-23.feature",
				"../../../features/ui/modelle-sprache/story-23.feature",
			}, TestingT: t,
		},
	}
	if runner.Run() != 0 {
		t.Fatal("Story-23-Godog-Szenarien fehlgeschlagen")
	}
}
