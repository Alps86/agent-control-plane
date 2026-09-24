package ops01

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

	runner := godog.TestSuite{
		ScenarioInitializer: suite.InitializeScenario,
		Options: &godog.Options{
			Format: "pretty", Paths: []string{"../../../features/app/betrieb/story-10.feature"}, TestingT: t,
		},
	}
	if runner.Run() != 0 {
		t.Fatal("OPS-01-Godog-Szenarien fehlgeschlagen")
	}
}
