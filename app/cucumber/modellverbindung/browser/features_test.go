//go:build browser

package browser

import (
	"testing"

	"github.com/cucumber/godog"
)

func TestFeatures(t *testing.T) {
	suite := NewSuite(t)
	runner := godog.TestSuite{
		ScenarioInitializer: suite.InitializeScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"../../../../features/ui/modelle-sprache/story-09.feature"},
			Strict:   true,
			TestingT: t,
		},
	}
	if runner.Run() != 0 {
		t.Fatal("Browser-Vertrag fehlgeschlagen")
	}
}
