package aktivitaet

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

	runner := godog.TestSuite{ScenarioInitializer: suite.initialize,
		Options: &godog.Options{Format: "pretty", Paths: []string{
			"../../../features/app/aktivitaet/story-18.feature",
			"../../../features/ui/aktivitaet/story-18.feature",
		}, TestingT: t}}
	if runner.Run() != 0 {
		t.Fatal("ACT-01-Godog-Szenarien fehlgeschlagen")
	}
}
