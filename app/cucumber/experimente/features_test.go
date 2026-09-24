package experimente

import (
	"fmt"
	"testing"

	"github.com/cucumber/godog"
)

func TestFeatures(t *testing.T) {
	suite := NewSuite()
	if err := suite.runAppFeatures(t); err != nil {
		t.Fatal(err)
	}

	if err := suite.runBrowserFeatures(t); err != nil {
		t.Fatal(err)
	}
}

func (s *Suite) runAppFeatures(t *testing.T) error {
	runner := godog.TestSuite{
		ScenarioInitializer: s.InitializeScenario,
		Options: &godog.Options{
			Format: "pretty", Paths: []string{"../../../features/app/experimente/story-88.feature"}, TestingT: t,
		},
	}
	if runner.Run() != 0 {
		return fmt.Errorf("App-Godog-Szenarien fehlgeschlagen")
	}

	return nil
}

func (s *Suite) runBrowserFeatures(t *testing.T) error {
	runner := godog.TestSuite{
		ScenarioInitializer: s.InitializeBrowserScenario,
		Options: &godog.Options{
			Format: "pretty", Paths: []string{"../../../features/ui/experimente/story-88.feature"}, TestingT: t,
		},
	}
	if runner.Run() != 0 {
		return fmt.Errorf("Browser-Godog-Szenarien fehlgeschlagen")
	}

	return nil
}
