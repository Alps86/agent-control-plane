package openrouterverbindung

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"
)

func (s *Suite) readyStatus() error {
	if err := s.expectStatusText("einsatzbereit"); err != nil {
		return err
	}

	return s.expectLatestKey()
}

func (s *Suite) expectNoOldKey() error {
	if s.previousKey == "" {
		return fmt.Errorf("kein vorheriger Schlüssel für Rotation")
	}

	for _, header := range s.providerCalls {
		if header == "Bearer "+s.previousKey {
			return fmt.Errorf("alter Schlüssel wurde nach Rotation verwendet")
		}
	}

	return nil
}

func (s *Suite) missingRejected() error {
	if err := s.expectStatus(http.StatusBadRequest); err != nil {
		return err
	}

	if s.last.Error != "Schlüssel fehlt" {
		return fmt.Errorf("fehlender Schlüssel ohne passenden Hinweis: %s", s.responseBody)
	}

	return nil
}

func (s *Suite) notReady() error {
	if err := s.getSettings(); err != nil {
		return err
	}

	if s.last.Connected || s.last.Status == "einsatzbereit" {
		return fmt.Errorf("Verbindung fälschlich einsatzbereit: %s", s.responseBody)
	}

	return nil
}

func (s *Suite) invalidStatus() error {
	if err := s.expectStatusText("nicht einsatzbereit"); err != nil {
		return err
	}

	if !strings.Contains(s.last.StatusDetail, "abgelehnt") {
		return fmt.Errorf("kein verständlicher Authentifizierungsgrund: %s", s.responseBody)
	}

	return nil
}

func (s *Suite) onlyConfiguredProvider() error {
	if len(s.providerCalls) != 1 {
		return fmt.Errorf("erwartete genau eine Anfrage an den konfigurierten Anbieter, erhalten %d", len(s.providerCalls))
	}

	return s.expectLatestKey()
}

func (s *Suite) notConnectedStatus() error {
	return s.expectStatusText("nicht eingerichtet")
}

func (s *Suite) disconnectedNotice() error {
	if s.last.Notice != "Verbindung getrennt" {
		return fmt.Errorf("Trennungshinweis fehlt: %s", s.responseBody)
	}

	return s.expectStatusText("nicht eingerichtet")
}

func (s *Suite) sameAfterRestart() error {
	if err := s.getSettings(); err != nil {
		return err
	}

	return s.expectSameReference()
}

func (s *Suite) bothNoSecret() error {
	if bytes.Contains(s.htmlBody, []byte(s.currentKey)) || bytes.Contains(s.jsonBody, []byte(s.currentKey)) {
		return fmt.Errorf("Klartextschlüssel in HTML oder JSON")
	}

	return nil
}

func (s *Suite) bothNoFragments() error {
	for _, fragment := range []string{"VALID-Alpha-931527", "Alpha-931527"} {
		if bytes.Contains(s.htmlBody, []byte(fragment)) || bytes.Contains(s.jsonBody, []byte(fragment)) {
			return fmt.Errorf("Schlüsselfragment in HTML oder JSON")
		}
	}

	return nil
}

func (s *Suite) bothPublicMetadata() error {
	if s.last.Reference != "openrouter-central" || !bytes.Contains(s.htmlBody, []byte("openrouter-central")) {
		return fmt.Errorf("zentrale Referenz fehlt in HTML oder JSON")
	}

	return s.expectNoGrants()
}

func (s *Suite) onlyAfterCheck() error {
	if err := s.expectNoProviderCall(); err != nil {
		return err
	}

	if err := s.checkConnection(); err != nil {
		return err
	}

	return s.readyStatus()
}

func (s *Suite) forbidden() error {
	return s.expectStatus(http.StatusForbidden)
}

func (s *Suite) unchangedKey() error {
	if err := s.checkConnection(); err != nil {
		return err
	}

	return s.expectLatestKey()
}
