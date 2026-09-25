package berichtsweg

import (
	"github.com/cucumber/godog"
	"testing"
)

func TestFeatures(t *testing.T) {
	s := (&Suite{t: t}).newSuite()
	if err := s.build(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(s.stop)
	runner := godog.TestSuite{ScenarioInitializer: s.initialize, Options: &godog.Options{Format: "pretty", Paths: []string{"../../../features/app/agenten/story-37.feature", "../../../features/ui/agenten/story-37.feature"}, TestingT: t}}
	if runner.Run() != 0 {
		t.Fatal("AGT-04-Godog-Szenarien fehlgeschlagen")
	}
}
