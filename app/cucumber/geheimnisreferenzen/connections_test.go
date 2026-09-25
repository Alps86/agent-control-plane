package geheimnisreferenzen

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func (s *Suite) connectCodex() error {
	if err := s.request(http.MethodPost, "/settings/modelle/codex/device/start", nil); err != nil {
		return err
	}
	if s.status != http.StatusOK {
		return fmt.Errorf("Gerätecode-Start HTTP %d: %s", s.status, s.last)
	}
	if err := s.request(http.MethodGet, "/settings/modelle/codex/device/status", nil); err != nil {
		return err
	}
	if s.status != http.StatusOK || !strings.Contains(string(s.last), `"connected"`) {
		return fmt.Errorf("Codex nicht verbunden: %s", s.last)
	}
	return nil
}

func (s *Suite) connectRouter() error {
	if err := s.request(http.MethodPost, "/api/settings/modellanbieter/openrouter", map[string]string{"key": testKey}); err != nil {
		return err
	}
	if s.status != http.StatusOK {
		return fmt.Errorf("OpenRouter-Speichern HTTP %d: %s", s.status, s.last)
	}
	if err := s.request(http.MethodPost, "/api/settings/modellanbieter/openrouter/pruefen", nil); err != nil {
		return err
	}
	if s.status != http.StatusOK || !strings.Contains(string(s.last), "einsatzbereit") {
		return fmt.Errorf("OpenRouter-Prüfung: %s", s.last)
	}
	return nil
}

func (s *Suite) disconnectRouter() error {
	if err := s.request(http.MethodDelete, "/api/settings/modellanbieter/openrouter", nil); err != nil {
		return err
	}
	if s.status != http.StatusOK {
		return fmt.Errorf("OpenRouter-Trennung HTTP %d: %s", s.status, s.last)
	}
	return nil
}

func (s *Suite) bothConnected() error {
	view, err := s.statusView()
	if err != nil {
		return err
	}
	return view.bothConnected()
}

func (view StatusView) bothConnected() error {
	codex, err := view.provider("codex-abo")
	if err != nil {
		return err
	}
	router, err := view.provider("openrouter")
	if err != nil {
		return err
	}
	if codex.Reference != "codex-chatgpt" || router.Reference != "openrouter-central" {
		return fmt.Errorf("Referenzen: %+v %+v", codex, router)
	}
	if codex.Status != "connected" || router.Status != "connected" || !codex.Ready || !router.Ready {
		return fmt.Errorf("Status: %+v %+v", codex, router)
	}
	return nil
}

func (s *Suite) codexAuth() error  { return s.auth("codex-abo", "device_code") }
func (s *Suite) routerAuth() error { return s.auth("openrouter", "api_key") }
func (s *Suite) auth(id, kind string) error {
	view, err := s.statusView()
	if err != nil {
		return err
	}
	provider, err := view.provider(id)
	if err != nil {
		return err
	}
	if provider.AuthType != kind {
		return fmt.Errorf("%s Authart %q statt %q", id, provider.AuthType, kind)
	}
	return nil
}

func (s *Suite) sameReferences() error {
	var before StatusView
	if err := json.Unmarshal(s.before, &before); err != nil {
		return err
	}
	after, err := s.statusView()
	if err != nil {
		return err
	}
	for _, id := range []string{"codex-abo", "openrouter"} {
		old, _ := before.provider(id)
		now, _ := after.provider(id)
		if old.Reference == "" || old.Reference != now.Reference {
			return fmt.Errorf("%s Referenz wechselte: %q → %q", id, old.Reference, now.Reference)
		}
	}
	return nil
}

func (s *Suite) routerDisconnected() error {
	view, err := s.statusView()
	if err != nil {
		return err
	}
	provider, err := view.provider("openrouter")
	if err != nil {
		return err
	}
	if provider.Status != "disconnected" || provider.Ready {
		return fmt.Errorf("OpenRouter weiter bereit: %+v", provider)
	}
	return nil
}

func (s *Suite) noRouterReference() error {
	view, err := s.statusView()
	if err != nil {
		return err
	}
	provider, err := view.provider("openrouter")
	if err != nil {
		return err
	}
	if provider.Reference != "" {
		return fmt.Errorf("getrennte Referenz bleibt sichtbar: %q", provider.Reference)
	}
	return nil
}

func (s *Suite) verifyDisconnectedAfterRestart() error { return s.getStatusAndCheckDisconnected() }
func (s *Suite) getStatusAndCheckDisconnected() error {
	if err := s.getStatus(); err != nil {
		return err
	}
	return s.routerDisconnected()
}

func (s *Suite) noStatusSecrets() error {
	for _, body := range s.snapshots {
		for _, secret := range []string{testKey, testAccess, testRefresh} {
			if strings.Contains(string(body), secret) {
				return fmt.Errorf("öffentliche Antwort enthält synthetisches Secret")
			}
		}
	}
	return nil
}
