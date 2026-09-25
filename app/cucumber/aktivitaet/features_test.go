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

	story66 := &Story66{Suite: suite}
	runner := godog.TestSuite{ScenarioInitializer: story66.initialize66,
		Options: &godog.Options{Format: "pretty", Paths: []string{
			"../../../features/app/aktivitaet/story-18.feature",
			"../../../features/ui/aktivitaet/story-18.feature",
			"../../../features/app/aktivitaet/story-66.feature",
			"../../../features/ui/aktivitaet/story-66.feature",
		}, TestingT: t}}
	if runner.Run() != 0 {
		t.Fatal("Aktivitäts-Godog-Szenarien fehlgeschlagen")
	}
}
