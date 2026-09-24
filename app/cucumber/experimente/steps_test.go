package experimente

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"

	webexperimente "agentcontrolplane/app/internal/adapter/web/experimente"
	"agentcontrolplane/ui/bridge"
	"github.com/cucumber/godog"
)

// NewSuite erzeugt den Kontext für öffentliche HTTP-Szenarien.
func NewSuite() *Suite {
	return &Suite{}
}

func (s *Suite) InitializeScenario(sc *godog.ScenarioContext) {
	sc.Before(s.before)
	sc.After(s.after)
	s.registerRequests(sc)
	s.registerInventory(sc)
	s.registerDecisions(sc)
}

func (s *Suite) registerRequests(sc *godog.ScenarioContext) {
	sc.Step(`^die Anwendung verwendet das freigegebene Inventar der Paperclip-Experimente$`, s.inventory)
	sc.Step(`^der Betreiber die Inventarseite über GET /experimente öffnet$`, s.openPage)
	sc.Step(`^der Betreiber die Inventaransicht über GET /experimente mit HX-Request öffnet$`, s.openFragment)
	sc.Step(`^antwortet die Anwendung mit Status 200 und einer deutschsprachigen HTML-Seite$`, s.pageResponse)
	sc.Step(`^antwortet die Anwendung mit Status 200 und einem HTML-Fragment ohne zweite Dokumenthülle$`, s.fragmentResponse)
	sc.Step(`^die Seite zeigt den Quellenstand des Inventars$`, s.sourceDate)
}

func (s *Suite) registerInventory(sc *godog.ScenarioContext) {
	sc.Step(`^sie zeigt die 14 experimentellen Bedien- und Laufzeitflächen sowie die 9 gesonderten API-Flächen jeweils genau einmal$`, s.groups)
	sc.Step(`^zu jedem Eintrag sind Name, sichere Primärquelle, tatsächlicher Paperclip-Reifegrad, Nutzen, ACP-Abhängigkeiten und begründete Entscheidung lesbar$`, s.entries)
	sc.Step(`^Plugin-Alpha, isolierte Workspaces, Cases, Chat-Style Tasks, Environments, External Objects und Status Cards sind enthalten$`, s.namedEntries)
	sc.Step(`^das Fragment zeigt dieselben 23 freigegebenen Einträge und Entscheidungen wie die Vollseite$`, s.sameFragment)
}

func (s *Suite) registerDecisions(sc *godog.ScenarioContext) {
	sc.Step(`^sind Zuordnen, Später evaluieren und Abgelöst als verschiedene Entscheidungen erkennbar$`, s.decisions)
	sc.Step(`^die später zu evaluierenden Funktionen bleiben mit Nutzen, Risiko und benötigtem Folgeentscheid sichtbar$`, s.laterDecisions)
	sc.Step(`^Alpha, Experiment, API-dokumentiert und zum Kern befördert bleiben als unterschiedliche Paperclip-Reifegrade erkennbar$`, s.maturity)
	sc.Step(`^jeder Eintrag nennt den ausdrücklich nachgewiesenen ACP-Implementierungsstatus oder kennzeichnet ihn als noch nicht nachgewiesen$`, s.statuses)
	sc.Step(`^keine zugeordnete oder experimentelle Funktion wird allein wegen der Inventarisierung als bereits in ACP verfügbar bezeichnet$`, s.noFalseClaim)
	sc.Step(`^führen alle externen Quellenlinks ausschließlich zu den freigegebenen HTTPS-Primärquellen$`, s.safeLinks)
	sc.Step(`^angezeigter Quellentext wird als Text statt als ausführbares HTML ausgegeben$`, s.escapedText)
}

func (s *Suite) before(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
	s.status, s.body, s.full = 0, "", ""
	s.header = nil
	s.rows = nil
	s.page = BrowserPage{}
	s.viewport = 1280
	return ctx, nil
}

func (s *Suite) after(ctx context.Context, _ *godog.Scenario, _ error) (context.Context, error) {
	if s.server != nil {
		s.server.Close()
		s.server = nil
	}

	return ctx, nil
}

func (s *Suite) inventory() error {
	rows, err := (&Inventory{}).read()
	if err != nil {
		return err
	}

	s.rows = rows
	renderer, err := bridge.New()
	if err != nil {
		return err
	}

	mux := http.NewServeMux()
	mux.Handle("/experimente", webexperimente.NewHandler(renderer))
	mux.Handle("/assets/", renderer.Assets())
	s.server = httptest.NewServer(mux)
	return nil
}

func (s *Suite) openPage() error {
	return s.get(false)
}

func (s *Suite) openFragment() error {
	return s.get(true)
}

func (s *Suite) get(fragment bool) error {
	request, err := http.NewRequest(http.MethodGet, s.server.URL+"/experimente", nil)
	if err != nil {
		return err
	}

	if fragment {
		request.Header.Set("HX-Request", "true")
	}

	return s.send(request)
}

func (s *Suite) send(request *http.Request) error {
	response, err := s.server.Client().Do(request)
	if err != nil {
		return err
	}

	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	s.status, s.header, s.body = response.StatusCode, response.Header, string(body)
	return err
}

func (s *Suite) pageResponse() error {
	if err := s.response(); err != nil {
		return err
	}

	if !strings.Contains(s.body, `<html lang="de">`) || !strings.Contains(s.body, "<!doctype html>") {
		return fmt.Errorf("deutsche HTML-Dokumenthülle fehlt")
	}

	return nil
}

func (s *Suite) fragmentResponse() error {
	if err := s.response(); err != nil {
		return err
	}

	if strings.Contains(s.body, "<!doctype") || strings.Contains(s.body, "<html") {
		return fmt.Errorf("HX-Antwort enthält eine Dokumenthülle")
	}

	return nil
}

func (s *Suite) response() error {
	if s.status != http.StatusOK {
		return fmt.Errorf("HTTP-Status %d statt 200: %s", s.status, s.body)
	}

	if !strings.HasPrefix(s.header.Get("Content-Type"), "text/html") {
		return fmt.Errorf("Content-Type %q statt HTML", s.header.Get("Content-Type"))
	}

	return nil
}
