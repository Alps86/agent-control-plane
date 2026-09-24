package cucumber

import (
	"testing"

	"github.com/cucumber/godog"
)

func TestFeatures(t *testing.T) {
	suite := NewSuite(t)
	t.Cleanup(suite.stop)
	runner := godog.TestSuite{
		ScenarioInitializer: suite.InitializeScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"../../features/app/architektur/story-01.feature"},
			TestingT: t,
		},
	}
	if runner.Run() != 0 {
		t.Fatal("Godog-Szenario fehlgeschlagen")
	}
}
