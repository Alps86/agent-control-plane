package installation

import (
	"testing"

	"github.com/cucumber/godog"
)

func TestFeatures(t *testing.T) {
	suite := NewSuite(t)
	if err := suite.build(); err != nil {
		t.Fatal(err)
	}

	t.Cleanup(suite.stop)
	runner := godog.TestSuite{ScenarioInitializer: suite.InitializeScenario,
		Options: &godog.Options{Format: "pretty",
			Paths: []string{"../../../features/app/betrieb/story-22.feature"}, TestingT: t}}
	if runner.Run() != 0 {
		t.Fatal("OPS-05-Godog-Szenarien fehlgeschlagen")
	}
}
