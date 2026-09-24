package lauf

import (
	"testing"

	"github.com/cucumber/godog"
)

func TestFeatures(t *testing.T) {
	suite := NewSuite()
	runner := godog.TestSuite{
		ScenarioInitializer: suite.InitializeScenario,
		Options: &godog.Options{
			Format: "pretty", Paths: []string{"../../../features/app/lauf/story-04.feature"}, TestingT: t,
		},
	}
	if runner.Run() != 0 {
		t.Fatal("Godog-Szenario fehlgeschlagen")
	}
}
