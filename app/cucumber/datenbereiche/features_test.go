package datenbereiche

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
			"../../../features/app/datenbereiche/story-19.feature",
			"../../../features/ui/datenbereiche/story-19.feature",
		}, TestingT: t}}
	if runner.Run() != 0 {
		t.Fatal("PERM-01-Godog-Szenarien fehlgeschlagen")
	}
}
