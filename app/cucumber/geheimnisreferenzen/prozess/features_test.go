package prozess

import (
	"testing"

	"github.com/cucumber/godog"
)

func TestFeatures(t *testing.T) {
	suite := &Suite{t: t}
	runner := godog.TestSuite{ScenarioInitializer: suite.InitializeScenario, Options: &godog.Options{
		Paths: []string{"../../../../features/ui/betrieb/story-79.feature"}, Format: "pretty", TestingT: t,
	}}
	if runner.Run() != 0 {
		t.Fatal("Story-79-Browserabnahme fehlgeschlagen")
	}
}
