package prozess

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/cucumber/godog"
)

func NewSuite(t *testing.T) *Suite {
	return &Suite{t: t, client: &http.Client{Timeout: 5 * time.Second}}
}

func (s *Suite) InitializeScenario(sc *godog.ScenarioContext) {
	sc.After(s.cleanup)
	sc.Step(`^ein gebauter lokaler Server mit temporärer SQLite-Datei und getrennten Credential- und Masterkey-Dateien ist gestartet$`, s.start)
	sc.Step(`^ich Settings über die öffentliche Navigation öffne$`, s.openNavigation)
	sc.Step(`^ich die OpenRouter-Einstellungen über die öffentliche Settings-Seite öffne$`, s.openSettings)
	sc.Step(`^erreiche ich die OpenRouter-Settings-Ansicht und ihr eingebautes Stylesheet$`, s.checkSettingsAssets)
	sc.Step(`^der echte Browser zeigt die montierte Settings-Seite ohne Schlüssel im DOM$`, s.chromePage)
	sc.Step(`^die öffentliche JSON-Antwort zeigt "nicht eingerichtet"$`, s.notConnected)
	sc.Step(`^zeigt die öffentliche JSON-Antwort "nicht eingerichtet"$`, s.notConnected)
	sc.Step(`^ich einen synthetischen OpenRouter-Schlüssel über das öffentliche Formular speichere$`, s.saveSyntheticKey)
	sc.Step(`^zeigen HTML und JSON dieselbe zentrale Verbindungsreferenz ohne den Schlüssel$`, s.connectionNoSecret)
	sc.Step(`^die geschützte Credential-Datei enthält den Schlüssel nicht im Klartext$`, s.encryptedFile)
	sc.Step(`^ich den Server mit denselben Daten- und Secret-Pfaden neu starte$`, s.restart)
	sc.Step(`^zeigt die öffentliche JSON-Antwort dieselbe zentrale Verbindungsreferenz ohne den Schlüssel$`, s.sameAfterRestart)
	sc.Step(`^ich die Verbindung über die öffentliche Settings-API trenne$`, s.disconnect)
	sc.Step(`^es wurde keine OpenRouter-Statusprüfung ausgelöst$`, s.noProbe)
}

func (s *Suite) cleanup(ctx context.Context, _ *godog.Scenario, _ error) (context.Context, error) {
	s.stop()
	if s.proxy != nil {
		s.proxy.Close()
	}
	return ctx, nil
}
