package uibridge

import (
	"bytes"
	"fmt"
	"strings"
)

func (s *Suite) neutralTemplatesBuilt() error {
	for _, name := range []string{"page.html", "content.html"} {
		if err := s.verifyCopy("templates/bridge-check/"+name, false); err != nil {
			return err
		}
	}

	return nil
}

func (s *Suite) sameViewMap() error {
	data, err := s.bridge.LoadFixture("organization.json")
	if err != nil {
		return err
	}

	data["Notice"] = "Öffentliche Bridge prüft alle Vorlagen"
	s.data = data
	return nil
}

func (s *Suite) renderNeutral() error {
	var page, content bytes.Buffer
	if err := s.bridge.Render(&page, "bridge-check/page", s.data); err != nil {
		return err
	}

	if err := s.bridge.Render(&content, "bridge-check/content", s.data); err != nil {
		return err
	}

	s.check, s.checkPart = page.String(), content.String()
	return nil
}

func (s *Suite) neutralValues() error {
	for _, value := range []string{"Übersicht · Agent Control Plane", "Öffentliche Bridge prüft alle Vorlagen"} {
		if !strings.Contains(s.check, value) || !strings.Contains(s.checkPart, value) {
			return fmt.Errorf("neutral template missing %q", value)
		}
	}

	if !strings.Contains(s.check, "<!doctype html>") || strings.Contains(s.checkPart, "<!doctype html>") {
		return fmt.Errorf("neutral page and content frames differ incorrectly")
	}

	return nil
}

func (s *Suite) rootTemplatesStillWork() error {
	var page, content bytes.Buffer
	if err := s.bridge.Render(&page, "page", s.data); err != nil {
		return err
	}

	if err := s.bridge.Render(&content, "content", s.data); err != nil {
		return err
	}

	if !strings.Contains(page.String(), "Atelier Nord") || !strings.Contains(content.String(), "Atelier Nord") {
		return fmt.Errorf("root templates lost shared map")
	}

	return nil
}
