package prozess

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

func (s *Suite) openNavigation() error {
	if err := s.get("/organisationen"); err != nil {
		return err
	}
	if err := s.pageLinks("/settings"); err != nil {
		return err
	}
	if err := s.get("/settings"); err != nil {
		return err
	}
	if err := s.pageLinks(htmlPath); err != nil {
		return err
	}
	if err := s.get(htmlPath); err != nil {
		return err
	}
	s.html = bytes.Clone(s.body)
	return s.expectStatus(http.StatusOK)
}

func (s *Suite) openSettings() error {
	if err := s.get("/settings"); err != nil {
		return err
	}
	if err := s.pageLinks(htmlPath); err != nil {
		return err
	}
	if err := s.get(htmlPath); err != nil {
		return err
	}
	s.html = bytes.Clone(s.body)
	return s.expectStatus(http.StatusOK)
}

func (s *Suite) pageLinks(path string) error {
	if err := s.expectStatus(http.StatusOK); err != nil {
		return err
	}
	if !bytes.Contains(s.body, []byte(`href="`+path+`"`)) {
		return fmt.Errorf("öffentlicher Navigationslink %s fehlt", path)
	}
	return nil
}

func (s *Suite) checkSettingsAssets() error {
	if !bytes.Contains(s.html, []byte(`href="/assets/index.css"`)) {
		return fmt.Errorf("OpenRouter-Seite bindet Stylesheet nicht ein")
	}
	if err := s.get("/assets/index.css"); err != nil {
		return err
	}
	if err := s.expectStatus(http.StatusOK); err != nil {
		return err
	}
	if !strings.HasPrefix(s.contentType, "text/css") || len(bytes.TrimSpace(s.body)) == 0 {
		return fmt.Errorf("Stylesheet ohne CSS-Inhalt: %q, %d Bytes", s.contentType, len(s.body))
	}
	return nil
}

func (s *Suite) notConnected() error {
	view, err := s.connection()
	if err != nil {
		return err
	}
	if view.Connected || view.Status != "nicht eingerichtet" || view.Reference != "" {
		return fmt.Errorf("unerwarteter öffentlicher Trennstatus: %+v", view)
	}
	return nil
}

func (s *Suite) saveSyntheticKey() error {
	if err := s.postForm(htmlPath, url.Values{"key": {syntheticKey}}); err != nil {
		return err
	}
	if err := s.expectStatus(http.StatusOK); err != nil {
		return err
	}
	s.html = bytes.Clone(s.body)
	return s.secretAbsent(s.html, "HTML-Antwort")
}

func (s *Suite) connectionNoSecret() error {
	view, err := s.connection()
	if err != nil {
		return err
	}
	if !view.Connected || view.Reference != "openrouter-central" || view.Status != "nicht geprüft" {
		return fmt.Errorf("unerwartete öffentliche Verbindung: %+v", view)
	}
	s.reference = view.Reference
	if !bytes.Contains(s.html, []byte(s.reference)) {
		return fmt.Errorf("HTML und JSON zeigen unterschiedliche Referenzen")
	}
	return s.secretAbsent(s.html, "HTML-Antwort")
}

func (s *Suite) encryptedFile() error {
	if s.secretPath == s.keyPath || s.secretPath == s.dbPath || s.keyPath == s.dbPath {
		return fmt.Errorf("Daten- und Secret-Pfade sind nicht getrennt")
	}
	for _, path := range []string{s.secretPath, s.keyPath} {
		if err := s.secureFile(path); err != nil {
			return err
		}
	}
	if err := s.fileWithoutSecret(s.dbPath); err != nil {
		return err
	}
	return s.scanDataDirectory()
}

func (s *Suite) secureFile(path string) error {
	if err := s.fileWithoutSecret(path); err != nil {
		return err
	}
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if info.Mode().Perm()&0077 != 0 {
		return fmt.Errorf("Datei %s ist für andere Benutzer lesbar", path)
	}
	return nil
}

func (s *Suite) fileWithoutSecret(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return s.secretAbsent(data, path)
}

func (s *Suite) scanDataDirectory() error {
	return filepath.WalkDir(filepath.Dir(s.dbPath), func(path string, entry os.DirEntry, err error) error {
		if err != nil || entry.IsDir() || path == s.binary {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return s.secretAbsent(data, path)
	})
}

func (s *Suite) secretAbsent(data []byte, source string) error {
	if bytes.Contains(data, []byte(syntheticKey)) {
		return fmt.Errorf("synthetischer Schlüssel im Klartext in %s", source)
	}
	return nil
}

func (s *Suite) sameAfterRestart() error {
	view, err := s.connection()
	if err != nil {
		return err
	}
	if !view.Connected || view.Reference != s.reference || view.Status != "nicht geprüft" {
		return fmt.Errorf("Verbindung nach Neustart nicht stabil: %+v", view)
	}
	return nil
}

func (s *Suite) disconnect() error {
	if err := s.delete(apiPath); err != nil {
		return err
	}
	return s.expectStatus(http.StatusOK)
}

func (s *Suite) noProbe() error {
	for _, request := range s.requested {
		if strings.Contains(request, "/pruefen") {
			return fmt.Errorf("öffentliche Anbieterprüfung aufgerufen: %s", request)
		}
	}
	if count := s.proxyCalls.Load(); count != 0 {
		return fmt.Errorf("%d ausgehende Anbieteranfragen am Sperrproxy beobachtet", count)
	}
	return nil
}
