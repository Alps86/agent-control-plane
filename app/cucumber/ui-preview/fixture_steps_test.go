package uipreview

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func (s *Suite) alternateName(want, absent string) error {
	if !strings.Contains(s.page.Text, want) || strings.Contains(s.page.Text, absent) {
		return fmt.Errorf("alternate fixture mismatch: %s", s.page.Text)
	}

	return nil
}

func (s *Suite) clickProjects() error {
	page, err := s.browserCommand(map[string]any{"click": `nav a[href*="view=projects"]`})
	s.page = page
	return err
}

func (s *Suite) fixtureAddress(fixture string) error {
	if !strings.Contains(s.page.URL, "fixture="+fixture) {
		return fmt.Errorf("fixture disappeared from URL: %s", s.page.URL)
	}

	return nil
}

func (s *Suite) alternateProjects(name string) error {
	if name != "projects-alternate.json" {
		return fmt.Errorf("alternate project fixture not rendered: %s", s.page.Text)
	}

	result, err := s.fetch("/?view=projects&fixture=alternate", false)
	if err != nil || !strings.Contains(result.Body, "Küstenwerk") {
		return fmt.Errorf("alternate project page: %v, %s", err, result.Body)
	}

	return s.contains("Kundenportal")
}

func (s *Suite) copyFixture() error {
	s.copyDir = filepath.Join(s.t.TempDir(), "ui")
	copy := exec.Command("cp", "-a", "../../../ui", s.copyDir)
	if output, err := copy.CombinedOutput(); err != nil {
		return fmt.Errorf("copy UI: %w: %s", err, output)
	}

	s.changedName = "Nur JSON geändert 07"
	file := filepath.Join(s.copyDir, "web/fixtures/organization.json")
	content, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	return os.WriteFile(file, []byte(strings.Replace(string(content), "Atelier Nord", s.changedName, 1)), 0644)
}

func (s *Suite) rebuildFixture() error {
	build := exec.Command("npm", "run", "build")
	build.Dir = filepath.Join(s.copyDir, "web")
	if output, err := build.CombinedOutput(); err != nil {
		return fmt.Errorf("rebuild UI: %w: %s", err, output)
	}

	if err := s.build(s.copyDir); err != nil {
		return err
	}

	s.stopServer()
	return s.startServer(s.copyDir)
}

func (s *Suite) changedFixture() error {
	if err := s.browse("/?view=organization"); err != nil {
		return err
	}

	return s.contains(s.changedName)
}

func (s *Suite) emptyProjects() error {
	return s.contains("Noch keine Projekte")
}

func (s *Suite) tasksError() error {
	if err := s.contains("Aufgaben nicht verfügbar"); err != nil {
		return err
	}

	return s.contains("Zur Übersicht")
}
