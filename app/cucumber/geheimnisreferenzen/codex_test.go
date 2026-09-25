package geheimnisreferenzen

import (
	"fmt"
	"net/http"
)

func (s *Suite) shortCodex() error  { s.issuerState.short.Store(true); return nil }
func (s *Suite) revokeCodex() error { s.issuerState.revoked.Store(true); return nil }

func (s *Suite) requestRevokedCodex() error {
	if err := s.request(http.MethodGet, "/settings/modelle/codex/device/status", nil); err != nil {
		return err
	}
	if s.status != http.StatusOK {
		return fmt.Errorf("Codex-Status HTTP %d: %s", s.status, s.last)
	}
	return nil
}

func (s *Suite) codexReauthentication() error {
	view, err := s.statusView()
	if err != nil {
		return err
	}
	provider, err := view.provider("codex-abo")
	if err != nil {
		return err
	}
	if provider.Status != "reauthentication_required" || provider.Ready || provider.Reference != "" {
		return fmt.Errorf("widerrufene Codex-Sitzung: %+v", provider)
	}
	return nil
}
