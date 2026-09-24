package openrouterverbindung

import (
	"bytes"
	"fmt"

	"github.com/cucumber/godog"
)

func (s *Suite) registerUISteps(sc *godog.ScenarioContext) {
	sc.Step(`^ich öffne "Settings > Modellanbieter" ohne OpenRouter-Verbindung$`, s.openEmptySettings)
	sc.Step(`^sehe ich vor der Schlüsseleingabe einen Hinweis auf die separate OpenRouter-Abrechnung$`, s.billingNotice)
	sc.Step(`^die Oberfläche bietet OpenRouter als bewusste Verbindung an$`, s.consciousChoice)
	sc.Step(`^sie stellt OpenRouter nicht als automatischen Codex-Ersatz dar$`, s.noCodexFallback)
	sc.Step(`^ich einen OpenRouter-Schlüssel in das verdeckte Eingabefeld eingebe und speichere$`, s.saveThroughHTML)
	sc.Step(`^zeigt die Oberfläche nur eine maskierte Verbindungsreferenz$`, s.htmlReference)
	sc.Step(`^sie zeigt den Schlüssel weder im Eingabefeld noch im sichtbaren Seiteninhalt erneut an$`, s.htmlNoSecret)
	sc.Step(`^ich die Statusprüfung auslöse$`, s.checkThroughHTML)
	sc.Step(`^sehe ich den verständlichen Verbindungsstatus des kontrollierten Testanbieters$`, s.htmlReady)
	sc.Step(`^ich öffne "Settings > Modellanbieter" mit bestehender OpenRouter-Verbindung$`, s.openConnectedSettings)
	sc.Step(`^ich den Schlüssel über die Ersetzen-Aktion ändere$`, s.replaceThroughHTML)
	sc.Step(`^bleibt die zentrale Verbindungsreferenz sichtbar und der neue Schlüssel verborgen$`, s.htmlReference)
	sc.Step(`^ich die Verbindung trenne$`, s.disconnectThroughHTML)
	sc.Step(`^sehe ich "Verbindung getrennt" und kann keinen einsatzbereiten Status annehmen$`, s.htmlDisconnected)
	sc.Step(`^derselbe vollständige Bedienweg funktioniert in einem echten Browser$`, s.runBrowser)
}

func (s *Suite) openEmptySettings() error {
	return s.getHTML()
}

func (s *Suite) openConnectedSettings() error {
	if err := s.validConnection(); err != nil {
		return err
	}

	return s.getHTML()
}

func (s *Suite) billingNotice() error {
	if !bytes.Contains(s.responseBody, []byte("Separate Abrechnung")) {
		return fmt.Errorf("Abrechnungshinweis fehlt")
	}

	return nil
}

func (s *Suite) consciousChoice() error {
	if !bytes.Contains(s.responseBody, []byte("ausdrückliche Wahl")) {
		return fmt.Errorf("bewusste Anbieterwahl fehlt")
	}

	return nil
}

func (s *Suite) noCodexFallback() error {
	if bytes.Contains(s.responseBody, []byte("automatisch zum Codex-Abo")) {
		return fmt.Errorf("UI behauptet automatischen Codex-Ersatz")
	}

	return nil
}

func (s *Suite) saveThroughHTML() error {
	s.currentKey = validKey
	return s.sendForm(htmlPath, validKey)
}

func (s *Suite) replaceThroughHTML() error {
	s.previousKey, s.currentKey = s.currentKey, replacementKey
	return s.sendForm(htmlPath, replacementKey)
}

func (s *Suite) checkThroughHTML() error {
	return s.sendForm(htmlPath+"/pruefen", "")
}

func (s *Suite) disconnectThroughHTML() error {
	return s.sendForm(htmlPath+"/trennen", "")
}

func (s *Suite) htmlReference() error {
	if !bytes.Contains(s.responseBody, []byte("openrouter-central")) {
		return fmt.Errorf("zentrale Referenz fehlt in HTML")
	}

	return s.htmlNoSecret()
}

func (s *Suite) htmlNoSecret() error {
	if bytes.Contains(s.responseBody, []byte(s.currentKey)) {
		return fmt.Errorf("Schlüssel im HTML-Antwortkörper")
	}

	if !bytes.Contains(s.responseBody, []byte(`type="password"`)) {
		return fmt.Errorf("Passwortfeld fehlt")
	}

	return nil
}

func (s *Suite) htmlReady() error {
	if !bytes.Contains(s.responseBody, []byte("einsatzbereit")) {
		return fmt.Errorf("Status nicht einsatzbereit: %s", s.responseBody)
	}

	return s.expectLatestKey()
}

func (s *Suite) htmlDisconnected() error {
	if !bytes.Contains(s.responseBody, []byte("Verbindung getrennt")) {
		return fmt.Errorf("Trennungshinweis fehlt")
	}

	if bytes.Contains(s.responseBody, []byte(`class="provider-status" role="status">einsatzbereit`)) {
		return fmt.Errorf("getrennte Verbindung erscheint einsatzbereit")
	}

	return nil
}
