package modelllebenszyklus

import (
	"testing"

	"github.com/cucumber/godog"
)

func TestFeatures(t *testing.T) {
	suite := &Suite{t: t}
	runner := godog.TestSuite{
		ScenarioInitializer: suite.InitializeScenario,
		Options: &godog.Options{
			Paths:    []string{"../../../features/app/modelle-sprache/story-29.feature"},
			Tags:     "~@pending-openrouter-aufrufgrenze",
			TestingT: t,
			Format:   "pretty",
		},
	}
	if runner.Run() != 0 {
		t.Fatal("Story-29-Godog-Szenarien fehlgeschlagen")
	}
}
