package prozess

import (
	"testing"

	"github.com/cucumber/godog"
)

func TestFeatures(t *testing.T) {
	s := &suite{t: t}
	if err := s.build(); err != nil {
		t.Fatal(err)
	}
	runner := godog.TestSuite{ScenarioInitializer: s.initialize, Options: &godog.Options{
		Format: "pretty", Paths: []string{"../../../../features/app/modelle-sprache/story-24-probe.feature"}, TestingT: t,
	}}
	if runner.Run() != 0 {
		t.Fatal("Story-24-Prozessprobe fehlgeschlagen")
	}
}
