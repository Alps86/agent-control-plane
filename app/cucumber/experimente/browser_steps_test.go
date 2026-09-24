package experimente

import (
	"fmt"

	"github.com/cucumber/godog"
)

func (s *Suite) InitializeBrowserScenario(sc *godog.ScenarioContext) {
	sc.Before(s.before)
	sc.After(s.after)
	sc.Step(`^die Anwendung verwendet das freigegebene Paperclip-Inventar$`, s.inventory)
	sc.Step(`^ich betrachte die Anwendung auf einem schmalen Bildschirm$`, s.narrowScreen)
	sc.Step(`^ich die öffentliche Seite /experimente öffne$`, s.openBrowser)
	sc.Step(`^sehe ich den Quellenstand vom 24.09.2026$`, s.browserDate)
	sc.Step(`^ich sehe zwei getrennte Gruppen mit 14 experimentellen und 9 weiteren API-Quellzeilen$`, s.browserGroups)
	sc.Step(`^ich kann alle 23 Einträge ohne horizontales Scrollen lesen$`, s.browserMobile)
	sc.Step(`^jede Quellzeile nennt Reifegrad, Nutzen, ACP-Abhängigkeiten, Entscheidung und ACP-Implementierungsstand$`, s.browserFields)
	sc.Step(`^Plugin-Alpha, isolierte Workspaces, Cases, Chat-Style Tasks, Environments, External Objects und Status Cards sind lesbar$`, s.browserNamed)
	s.registerBrowserDecisions(sc)
}

func (s *Suite) registerBrowserDecisions(sc *godog.ScenarioContext) {
	sc.Step(`^erklärt die Seite Zuordnen, Später evaluieren, Abgelöst und Verweis getrennt$`, s.browserLegend)
	sc.Step(`^ein später zu evaluierender Eintrag nennt Nutzen, Risiko und erforderlichen Folgeentscheid$`, s.browserLater)
	sc.Step(`^eine Story-Zuordnung wird nicht als implementierte ACP-Funktion dargestellt$`, s.browserNoFalseClaim)
	sc.Step(`^die zweite Quellzeile zu Status Cards verweist sichtbar auf die zugehörige Entscheidung$`, s.browserRelated)
	sc.Step(`^führen die Quellenlinks zu den freigegebenen HTTPS-Seiten unter docs.paperclip.ing$`, s.browserLinks)
	sc.Step(`^die Seite hat eine Hauptüberschrift und erreichbare Gruppenüberschriften$`, s.browserHeadings)
	sc.Step(`^ich denselben Weg als HTMX-Fragment öffne$`, s.browserFragment)
	sc.Step(`^sehe ich dieselben 23 Quellzeilen ohne zweite Dokumenthülle$`, s.browserSameFragment)
}

func (s *Suite) narrowScreen() error {
	s.viewport = 375
	return nil
}

func (s *Suite) openBrowser() error {
	page, err := s.browse(s.server.URL+"/experimente", s.viewport)
	if err != nil {
		return err
	}

	s.page = page
	return s.browserAssert(page.AssetsLoaded && page.Title != "", "Browserseite oder Assets fehlen")
}

func (s *Suite) browserDate() error {
	return s.browserAssert(s.page.Source == "Quellenstand: 24.09.2026", "Quellenstand fehlt")
}

func (s *Suite) browserGroups() error {
	return s.browserAssert(s.page.Groups == 2 && s.page.Entries == 23, "Gruppen oder Quellzeilen fehlen")
}

func (s *Suite) browserMobile() error {
	return s.browserAssert(s.page.Entries == 23 && !s.page.MobileWide, "mobile Seite überläuft oder ist unvollständig")
}

func (s *Suite) browserFields() error {
	return s.browserAssert(s.page.FieldsComplete && s.page.StatusesComplete, "Kartenfelder oder ACP-Status fehlen")
}

func (s *Suite) browserNamed() error {
	return s.browserAssert(s.page.NamedComplete, "benannte Experimente fehlen")
}

func (s *Suite) browserLegend() error {
	return s.browserAssert(s.page.LegendComplete && s.page.DecisionsComplete, "Entscheidungslegende fehlt")
}

func (s *Suite) browserLater() error {
	return s.browserAssert(s.page.LaterComplete, "spätere Evaluation nicht verständlich")
}

func (s *Suite) browserNoFalseClaim() error {
	return s.browserAssert(s.page.NoFalseClaim, "unbelegte ACP-Verfügbarkeit behauptet")
}

func (s *Suite) browserRelated() error {
	return s.browserAssert(s.page.RelatedVisible, "Status-Cards-Querverweis fehlt")
}

func (s *Suite) browserLinks() error {
	return s.browserAssert(s.page.SourcesSafe, "Quellenlinks sind nicht freigegeben")
}

func (s *Suite) browserHeadings() error {
	return s.browserAssert(s.page.HeadingsComplete, "erreichbare Überschriften fehlen")
}

func (s *Suite) browserFragment() error {
	return s.browserAssert(s.page.FragmentSame, "Browser-Fragment weicht ab")
}

func (s *Suite) browserSameFragment() error {
	return s.browserAssert(s.page.FragmentSame && s.page.FragmentNoShell, "Fragmentinhalt oder Dokumenthülle falsch")
}

func (s *Suite) browserAssert(condition bool, message string) error {
	if !condition {
		return fmt.Errorf("%s: %+v", message, s.page)
	}

	return nil
}
