package prozess

import (
	"os"
	"testing"

	"github.com/cucumber/godog"
)

func TestFeatures(t *testing.T) {
	suite := NewSuite(t)
	tags := "@prozess-kern"
	if os.Getenv("STORY23_PROCESS_NAV") == "1" {
		tags = "@navigation"
	}
	runner := godog.TestSuite{
		ScenarioInitializer: suite.InitializeScenario,
		Options: &godog.Options{
			Format: "pretty", Paths: []string{"../../../../features/app/modelle-sprache/story-23-prozess.feature"}, Tags: tags, TestingT: t,
		},
	}
	if runner.Run() != 0 {
		t.Fatal("Story-23-Prozess-Godog-Szenario fehlgeschlagen")
	}
}
