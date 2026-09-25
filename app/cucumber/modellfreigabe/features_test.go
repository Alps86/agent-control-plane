package modellfreigabe

import (
	"testing"

	"github.com/cucumber/godog"
)

func TestFeatures(t *testing.T) {
	s := newSuite(t)
	runner := godog.TestSuite{ScenarioInitializer: s.initialize,
		Options: &godog.Options{Format: "pretty", Paths: []string{
			"../../../features/app/modelle-sprache/story-24.feature",
			"../../../features/ui/modelle-sprache/story-24.feature",
		}, TestingT: t}}
	if runner.Run() != 0 {
		t.Fatal("Story-24-Godog-Szenarien fehlgeschlagen")
	}
}
