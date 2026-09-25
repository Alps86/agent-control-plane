package projektarchiv

import (
	"github.com/cucumber/godog"
	"testing"
)

func TestFeatures(t *testing.T) {
	suite := NewSuite(t)
	t.Cleanup(suite.cleanup)
	if err := suite.build(); err != nil {
		t.Fatal(err)
	}
	runner := godog.TestSuite{ScenarioInitializer: suite.initialize, Options: &godog.Options{
		Format: "pretty", TestingT: t, Paths: []string{
			"../../../features/app/projekte/story-34.feature",
			"../../../features/ui/projekte/story-34.feature",
		},
	}}
	if runner.Run() != 0 {
		t.Fatal("PROJ-02-Godog-Szenarien fehlgeschlagen")
	}
}
