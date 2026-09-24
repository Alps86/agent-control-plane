package modellzugang

import (
	"os"
	"testing"

	"github.com/cucumber/godog"
)

func TestFeatures(t *testing.T) {
	suite := NewSuite(t)
	tags := "@offline_schnittstelle"
	if os.Getenv("MS01_LIVE_GATE") == "1" {
		tags = "@echtes_codex_abo"
		if os.Getenv("MS01_LIVE_TOKEN") == "" {
			t.Skip("MS-01-Live-Gate offen: freigegebener Abo-Testzugang fehlt")
		}
	}

	runner := godog.TestSuite{ScenarioInitializer: suite.InitializeScenario,
		Options: &godog.Options{Format: "pretty", Paths: []string{"../../../features/app/modelle-sprache/story-08.feature"}, Tags: tags, TestingT: t}}
	if runner.Run() != 0 {
		t.Fatal("Story-08-Godog-Szenarien fehlgeschlagen")
	}
}
