package modellwahl

import (
	"github.com/cucumber/godog"
	"testing"
)

func TestFeatures(t *testing.T) {
	suite := &Suite{t: t}
	t.Cleanup(suite.cleanup)
	if err := suite.build(); err != nil {
		t.Fatal(err)
	}
	runner := godog.TestSuite{ScenarioInitializer: suite.initialize, Options: &godog.Options{
		Format: "pretty", Paths: []string{
			"../../../features/app/modelle-sprache/story-25.feature",
			"../../../features/ui/modelle-sprache/story-25.feature",
		}, TestingT: t,
	}}
	if runner.Run() != 0 {
		t.Fatal("Story-25-Abnahme fehlgeschlagen")
	}
}
