package modellwahl

import (
	"fmt"
	"github.com/cucumber/godog"
	"net/http"
)

func (s *Suite) registerNoCatalog(sc *godog.ScenarioContext) {
	sc.Step(`^ich starte den lokalen Server ohne Modellkatalog-Konfiguration neu$`, s.noCatalogRestart)
	sc.Step(`^ich öffentliche OpenRouter-Settings und Miras Modellwahl über HTTP abrufe$`, s.noCatalogViews)
	sc.Step(`^bleiben Settings und Modellwahl mit HTTP 200 erreichbar$`, s.noCatalogReachable)
	sc.Step(`^der Katalog enthält keinen auswählbaren Modellkandidaten$`, s.noCatalogCandidates)
}

func (s *Suite) noCatalogRestart() error {
	s.catalogPath = ""
	return s.restart()
}

func (s *Suite) noCatalogViews() error {
	if err := s.request(http.MethodGet, "/api/settings/modellanbieter/openrouter", ""); err != nil {
		return err
	}
	s.settingsStatus = s.last.Status
	return s.readChoice()
}

func (s *Suite) noCatalogReachable() error {
	if s.settingsStatus != 200 || s.last.Status != 200 {
		return fmt.Errorf("Settings HTTP %d, Modellwahl HTTP %d", s.settingsStatus, s.last.Status)
	}
	return nil
}

func (s *Suite) noCatalogCandidates() error {
	view, err := s.view()
	if err != nil {
		return err
	}
	if view.Selection.Model != "" {
		return fmt.Errorf("Modell ohne Katalog vorausgewählt: %+v", view.Selection)
	}
	for _, provider := range view.Providers {
		if provider.Selectable || len(provider.Models) != 0 {
			return fmt.Errorf("Anbieter ohne Katalog auswählbar: %+v", provider)
		}
	}
	return nil
}
