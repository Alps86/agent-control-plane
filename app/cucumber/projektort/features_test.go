package projektort

import (
	"testing"

	"github.com/cucumber/godog"
)

func TestFeatures(t *testing.T) {
	s := newSuite(t)
	t.Cleanup(s.close)
	if err := s.build(); err != nil {
		t.Fatal(err)
	}
	runner := godog.TestSuite{ScenarioInitializer: s.steps, Options: &godog.Options{
		Format: "pretty",
		Paths: []string{
			"../../../features/app/projekte/story-35.feature",
			"../../../features/ui/projekte/story-35.feature",
		},
		TestingT: t,
	}}
	if runner.Run() != 0 {
		t.Fatal("Story-35-Godog-Szenarien fehlgeschlagen")
	}
}
