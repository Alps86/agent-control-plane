package openrouterverbindung

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

const apiPath = "/api/settings/modellanbieter/openrouter"
const htmlPath = "/settings/modellanbieter/openrouter"
const validKey = "sk-or-story23-VALID-Alpha-931527"
const replacementKey = "sk-or-story23-VALID-Beta-824016"
const invalidKey = "invalid-test-key"

func (s *Suite) getSettings() error {
	return s.sendJSON(http.MethodGet, apiPath, nil)
}

func (s *Suite) saveKey(key string) error {
	s.previousKey = s.currentKey
	s.currentKey = key
	return s.sendJSON(http.MethodPost, apiPath, map[string]string{"key": key})
}

func (s *Suite) checkConnection() error {
	return s.sendJSON(http.MethodPost, apiPath+"/pruefen", nil)
}

func (s *Suite) disconnect() error {
	return s.sendJSON(http.MethodDelete, apiPath, nil)
}

func (s *Suite) sendJSON(method, path string, body any) error {
	encoded, err := json.Marshal(body)
	if err != nil {
		return err
	}

	if err := s.send(method, path, "application/json", encoded); err != nil {
		return err
	}

	s.last = PublicResponse{}
	if err := json.Unmarshal(s.responseBody, &s.last); err != nil {
		return fmt.Errorf("ungültige JSON-Antwort: %w: %s", err, s.responseBody)
	}

	return nil
}

func (s *Suite) expectStatusText(status string) error {
	if s.last.Status != status {
		return fmt.Errorf("Status %q statt %q: %s", s.last.Status, status, s.responseBody)
	}

	return nil
}

func (s *Suite) expectNoSecret() error {
	for _, key := range []string{s.currentKey, s.previousKey} {
		if len(key) > 8 && strings.Contains(string(s.responseBody), key) {
			return fmt.Errorf("Klartextschlüssel in öffentlicher Antwort")
		}
	}

	return nil
}
