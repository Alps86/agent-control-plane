package aktivitaet

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"syscall"

	"github.com/cucumber/godog"
)

func (s *Story66) initialize66(sc *godog.ScenarioContext) {
	s.Suite.initialize(sc)
	s.query = nil
	s.pageEvents, s.allEvents, s.csvRows, s.pageBefore, s.rowsBefore = nil, nil, nil, nil, nil
	s.register66Setup(sc)
	s.register66API(sc)
	s.register66Browser(sc)
}

func (s *Story66) register66Setup(sc *godog.ScenarioContext) {
	sc.Step(`^in "([^"]+)" besteht der Agent "([^"]+)"$`, s.ensureAgent)
	sc.Step(`^mehrere Aufgabenereignisse mit verschiedenen Agenten, Aktionen und Zeitpunkten bestehen in "([^"]+)"$`, s.manyEvents)
	sc.Step(`^ein Aufgabenereignis in "([^"]+)" enthält Unicode, Komma, Anführungszeichen und Zeilenumbruch in einem sichtbaren Feld$`, s.quotedEvent)
	sc.Step(`^in "([^"]+)" besteht eine Aufgabe mit Aktivität$`, s.foreignEvent)
}

func (s *Story66) register66API(sc *godog.ScenarioContext) {
	sc.Step(`^ich den Aktivitätsverlauf von "([^"]+)" nach Agent "([^"]+)", Aktion "([^"]+)", dem Zeitpunkt des mittleren Ereignisses und dessen Aufgabe filtere$`, s.combinedFilter)
	sc.Step(`^enthält die gefilterte API-Antwort ausschließlich das passende Ereignis$`, s.oneMatch)
	sc.Step(`^jedes Ereignis enthält Kennung, Zeitpunkt, Aktion, Agent und Deep Link zur Aufgabe$`, s.completeEvents)
	sc.Step(`^ich den Aktivitätsverlauf von "([^"]+)" seitenweise mit zwei Ereignissen je Seite abrufe$`, s.pagedEvents)
	sc.Step(`^sind die Ereignisse über alle Seiten vollständig, überschneidungsfrei und stabil sortiert$`, s.stablePages)
	sc.Step(`^ein erneuter Abruf liefert dieselben Seiten in derselben Reihenfolge$`, s.repeatPages)
	sc.Step(`^ich eine gefilterte Seite des Aktivitätsverlaufs von "([^"]+)" abrufe und als CSV exportiere$`, s.filteredPageAndCSV)
	sc.Step(`^entsprechen die CSV-Datensätze in Inhalt und Reihenfolge exakt der API-Seite$`, s.csvMatchesPage)
	sc.Step(`^die CSV-Antwort ist als CSV-Download gekennzeichnet$`, s.csvDownload)
	sc.Step(`^ich den Aktivitätsverlauf von "([^"]+)" als CSV exportiere$`, s.exportAll)
	sc.Step(`^liest ein standardkonformer CSV-Parser den ursprünglichen Feldwert unverändert$`, s.quotedValue)
	sc.Step(`^die CSV enthält keinen zusätzlichen Ereignisdatensatz$`, s.oneCSVRecord)
	sc.Step(`^ich den Aktivitätsverlauf von "([^"]+)" nach einem nicht passenden Objekt filtere$`, s.emptyFilter)
	sc.Step(`^ist die gefilterte API-Antwort leer$`, s.emptyList)
	sc.Step(`^der Export enthält nur die CSV-Kopfzeile$`, s.headerOnly)
	sc.Step(`^ich den Aktivitätsverlauf von "([^"]+)" filtere und als CSV exportiere$`, s.ownCSV)
	sc.Step(`^enthält der Export keine Ereignisse aus "([^"]+)"$`, s.noForeignCSV)
	sc.Step(`^enthält weder Token noch Geheimnisse oder vollständige Prompts$`, s.noSecrets)
	sc.Step(`^ich einen ungültigen Zeitraum oder eine ungültige Seitengröße an die Aktivitäts-API sende$`, s.invalidFilters)
	sc.Step(`^werden Liste und Export mit einem verständlichen Clientfehler abgewiesen$`, s.badFilterErrors)
	sc.Step(`^liefern dieselben Filter dieselbe API-Seite und dieselben CSV-Datensätze$`, s.sameAfterRestart)
}

func (s *Story66) register66Browser(sc *godog.ScenarioContext) {
	sc.Step(`^ich Agent "([^"]+)", Aktion "([^"]+)", Zeitraum und Aufgabe als Aktivitätsfilter wähle$`, s.browserFilter)
	sc.Step(`^sehe ich nur die passenden Ereignisse in stabiler Reihenfolge$`, s.browserMatches)
	sc.Step(`^ich die sichtbare Aktivitätsseite als CSV herunterlade$`, s.browserCSV)
	sc.Step(`^enthält der Download genau die sichtbaren Ereignisse in derselben Reihenfolge$`, s.csvMatchesPage)
	sc.Step(`^ich einen Filter ohne Treffer wähle$`, s.browserEmptyFilter)
	sc.Step(`^sehe ich einen verständlichen Leerzustand$`, s.browserEmpty)
	sc.Step(`^der CSV-Download enthält nur die Kopfzeile und keine fremde Aktivität$`, s.browserEmptyCSV)
}

func (s *Story66) manyEvents(org string) error {
	for _, item := range []struct{ title, agent string }{{"Alpha", "Mira"}, {"Beta", "Noah"}, {"Gamma", "Mira"}} {
		if err := s.createTask(item.title, "Website", item.agent); err != nil {
			return err
		}
	}

	events, err := s.events(org)
	if err != nil {
		return err
	}

	if len(events) != 6 {
		return fmt.Errorf("sechs Ereignisse erwartet: %+v", events)
	}

	s.allEvents = events
	for _, event := range events {
		if event.Kind == "assigned" && event.AssigneeID == s.agents[s.key(org, "Mira")] {
			s.selected = event
			break
		}
	}

	return nil
}

func (s *Story66) quotedEvent(org string) error {
	if org != "Nordstern" {
		return fmt.Errorf("unbekannte Testorganisation %q", org)
	}

	return s.createTask("Übergabe, \"Mira\"\nzweite Zeile", "Website", "Mira")
}

func (s *Story66) foreignEvent(org string) error {
	if err := s.ensureProjectAndAgent(org, "Fremdprojekt", "Fremdagent"); err != nil {
		return err
	}

	title := "fremd-token-PROMPT-GEHEIMNIS"
	data, _ := json.Marshal(map[string]string{"title": title, "description": "Fremd", "priority": "normal", "assignee_id": s.agents[s.key(org, "Fremdagent")]})
	if err := s.request("POST", s.taskPath(org, "Fremdprojekt"), string(data)); err != nil {
		return err
	}

	if s.status != http.StatusCreated {
		return fmt.Errorf("Fremdaufgabe HTTP %d: %s", s.status, s.body)
	}

	return nil
}

func (s *Story66) activityPath(org string, export bool) string {
	path := "/api/organisationen/" + s.organizations[org] + "/aktivitaet"
	if export {
		path += "/export.csv"
	}

	if len(s.query) != 0 {
		path += "?" + s.query.Encode()
	}

	return path
}

func (s *Story66) fetchPage(org string) error {
	if err := s.request("GET", s.activityPath(org, false), ""); err != nil {
		return err
	}

	if s.status != http.StatusOK {
		return fmt.Errorf("Aktivität HTTP %d: %s", s.status, s.body)
	}

	var list EventList
	if err := json.Unmarshal(s.body, &list); err != nil {
		return err
	}

	s.pageEvents = list.Events
	return nil
}

func (s *Story66) fetchCSV(org string) error {
	response, err := s.client.Get(s.url(s.activityPath(org, true)))
	if err != nil {
		return err
	}

	defer response.Body.Close()
	s.status = response.StatusCode
	s.contentType = response.Header.Get("Content-Type")
	s.disposition = response.Header.Get("Content-Disposition")
	s.csvBody, err = io.ReadAll(response.Body)
	if err != nil {
		return err
	}

	if s.status != http.StatusOK {
		return fmt.Errorf("CSV HTTP %d: %s", s.status, s.csvBody)
	}

	s.csvRows, err = csv.NewReader(strings.NewReader(string(s.csvBody))).ReadAll()
	return err
}

func (s *Story66) combinedFilter(org, agent, action string) error {
	s.query = url.Values{"agent": {s.agents[s.key(org, agent)]}, "action": {action},
		"from": {s.selected.OccurredAt}, "to": {s.selected.OccurredAt}, "object": {s.selected.TaskID}}
	return s.fetchPage(org)
}

func (s *Story66) oneMatch() error {
	if len(s.pageEvents) != 1 || s.pageEvents[0].ID != s.selected.ID {
		return fmt.Errorf("kombinierter Filter: %+v, erwartet %+v", s.pageEvents, s.selected)
	}

	return nil
}

func (s *Story66) completeEvents() error {
	for _, event := range s.pageEvents {
		if event.ID == "" || event.OccurredAt == "" || event.Kind == "" || event.AssigneeID == "" || event.DeepLink == "" {
			return fmt.Errorf("Ereignis unvollständig: %+v", event)
		}
	}

	return nil
}

func (s *Story66) pagedEvents(org string) error {
	s.query = url.Values{"limit": {"2"}, "offset": {"0"}}
	for offset := 0; offset < len(s.allEvents); offset += 2 {
		s.query.Set("offset", fmt.Sprint(offset))
		if err := s.fetchPage(org); err != nil {
			return err
		}

		s.pageBefore = append(s.pageBefore, s.pageEvents...)
	}

	return nil
}

func (s *Story66) stablePages() error {
	if len(s.pageBefore) != len(s.allEvents) {
		return fmt.Errorf("Paging verlor Ereignisse: %+v", s.pageBefore)
	}

	seen := map[string]bool{}
	for _, event := range s.pageBefore {
		if seen[event.ID] {
			return fmt.Errorf("doppeltes Ereignis %s", event.ID)
		}
		seen[event.ID] = true
	}

	if !reflect.DeepEqual(s.pageBefore, s.allEvents) {
		return fmt.Errorf("Paging-Reihenfolge weicht ab")
	}

	return nil
}

func (s *Story66) repeatPages() error {
	org := "Nordstern"
	for offset := 0; offset < len(s.allEvents); offset += 2 {
		s.query.Set("offset", fmt.Sprint(offset))
		if err := s.fetchPage(org); err != nil {
			return err
		}

		end := offset + len(s.pageEvents)
		if !reflect.DeepEqual(s.pageEvents, s.pageBefore[offset:end]) {
			return fmt.Errorf("Seite %d änderte sich", offset/2)
		}
	}

	return nil
}

func (s *Story66) filteredPageAndCSV(org string) error {
	s.query = url.Values{"agent": {s.agents[s.key(org, "Mira")]}, "action": {"assigned"}, "limit": {"1"}, "offset": {"1"}}
	if err := s.fetchPage(org); err != nil {
		return err
	}

	if err := s.fetchCSV(org); err != nil {
		return err
	}

	s.pageBefore = append([]Event(nil), s.pageEvents...)
	s.rowsBefore = append([][]string(nil), s.csvRows...)
	return nil
}

func (s *Story66) csvMatchesPage() error {
	if len(s.csvRows) != len(s.pageEvents)+1 {
		return fmt.Errorf("CSV-Zeilenzahl %d, Ereignisse %d", len(s.csvRows), len(s.pageEvents))
	}

	header := []string{"id", "organization_id", "project_id", "task_id", "kind", "actor", "source", "occurred_at", "object_title", "deep_link", "assignee_id", "assignee_name"}
	if !reflect.DeepEqual(s.csvRows[0], header) {
		return fmt.Errorf("CSV-Kopfzeile: %+v", s.csvRows[0])
	}

	for index, event := range s.pageEvents {
		expected := []string{event.ID, event.OrganizationID, event.ProjectID, event.TaskID, event.Kind, event.Actor,
			event.Source, event.OccurredAt, event.ObjectTitle, event.DeepLink, event.AssigneeID, event.AssigneeName}
		if !reflect.DeepEqual(s.csvRows[index+1], expected) {
			return fmt.Errorf("CSV-Zeile %d: %+v, erwartet %+v", index, s.csvRows[index+1], expected)
		}
	}

	return nil
}

func (s *Story66) csvDownload() error {
	if !strings.Contains(s.contentType, "text/csv") || !strings.Contains(s.disposition, "attachment") {
		return fmt.Errorf("CSV-Kopfzeilen: %s / %s", s.contentType, s.disposition)
	}

	return nil
}

func (s *Story66) exportAll(org string) error {
	s.query = nil
	return s.fetchCSV(org)
}

func (s *Story66) quotedValue() error {
	value := "Übergabe, \"Mira\"\nzweite Zeile"
	for _, row := range s.csvRows[1:] {
		for _, field := range row {
			if field == value {
				return nil
			}
		}
	}

	return fmt.Errorf("CSV-Parser erhielt sichtbaren Feldwert nicht unverändert: %+v", s.csvRows)
}

func (s *Story66) oneCSVRecord() error {
	if len(s.csvRows) != 3 {
		return fmt.Errorf("Aufgabe muss genau zwei Aktivitätszeilen ergeben: %+v", s.csvRows)
	}

	return nil
}

func (s *Story66) emptyFilter(org string) error {
	s.query = url.Values{"object": {"nicht-vorhandene-aufgabe"}}
	if err := s.fetchPage(org); err != nil {
		return err
	}

	return s.fetchCSV(org)
}

func (s *Story66) emptyList() error {
	if len(s.pageEvents) != 0 {
		return fmt.Errorf("Filter ohne Treffer: %+v", s.pageEvents)
	}

	return nil
}

func (s *Story66) headerOnly() error {
	if len(s.csvRows) != 1 || len(s.csvRows[0]) == 0 {
		return fmt.Errorf("CSV ohne reine Kopfzeile: %+v", s.csvRows)
	}

	return nil
}

func (s *Story66) ownCSV(org string) error {
	if err := s.postTask("Eigene Aufgabe", "Website", "Mira", map[string]string{
		"description": "sk-test-secret VOLLPROMPT_MARKER"}); err != nil {
		return err
	}

	if s.status != http.StatusCreated {
		return fmt.Errorf("Eigene Aufgabe HTTP %d: %s", s.status, s.body)
	}

	s.query = url.Values{"action": {"created"}}
	return s.fetchCSV(org)
}

func (s *Story66) noForeignCSV(org string) error {
	foreign, err := s.events(org)
	if err != nil {
		return err
	}

	for _, event := range foreign {
		if strings.Contains(string(s.csvBody), event.ID) || strings.Contains(string(s.csvBody), event.ObjectTitle) {
			return fmt.Errorf("fremdes Ereignis im Export: %s", event.ID)
		}
	}

	return nil
}

func (s *Story66) noSecrets() error {
	for _, marker := range []string{"fremd-token", "PROMPT-GEHEIMNIS", "sk-test-secret", "VOLLPROMPT_MARKER"} {
		if strings.Contains(string(s.csvBody), marker) {
			return fmt.Errorf("Geheimnismarker im Export: %s", marker)
		}
	}

	return nil
}

func (s *Story66) invalidFilters() error {
	return nil
}

func (s *Story66) badFilterErrors() error {
	for _, query := range []string{"from=kein-datum", "limit=-1"} {
		for _, export := range []bool{false, true} {
			s.query, _ = url.ParseQuery(query)
			if err := s.request("GET", s.activityPath("Nordstern", export), ""); err != nil {
				return err
			}

			if s.status < 400 || s.status >= 500 || len(s.body) == 0 {
				return fmt.Errorf("Filter %s Export %v HTTP %d: %s", query, export, s.status, s.body)
			}
		}
	}

	return nil
}

func (s *Story66) sameAfterRestart() error {
	if err := s.fetchPage("Nordstern"); err != nil {
		return err
	}

	if !reflect.DeepEqual(s.pageEvents, s.pageBefore) {
		return fmt.Errorf("API nach Neustart verändert")
	}

	if err := s.fetchCSV("Nordstern"); err != nil {
		return err
	}

	if !reflect.DeepEqual(s.csvRows, s.rowsBefore) {
		return fmt.Errorf("CSV nach Neustart verändert")
	}

	return nil
}

func (s *Story66) browserFilter(agent, action string) error {
	s.query = url.Values{"agent": {s.agents[s.key("Nordstern", agent)]}, "action": {action},
		"from": {s.selected.OccurredAt}, "to": {s.selected.OccurredAt}, "object": {s.selected.TaskID}}
	if err := s.fetchPage("Nordstern"); err != nil {
		return err
	}

	if err := s.startBrowser66(); err != nil {
		return err
	}

	values := map[string]string{"agent": s.query.Get("agent"), "action": action,
		"from": s.query.Get("from"), "to": s.query.Get("to"), "object": s.query.Get("object")}
	return s.browser66Command(map[string]any{"filters": values, "mode": "activity"})
}

func (s *Story66) browserMatches() error {
	if err := s.oneMatch(); err != nil {
		return err
	}

	if !strings.Contains(s.page.Text, s.selected.ObjectTitle) {
		return fmt.Errorf("passendes Ereignis fehlt im Browser: %+v", s.page)
	}

	for _, event := range s.allEvents {
		if event.TaskID != s.selected.TaskID && strings.Contains(strings.Join(s.page.LinkTexts, "\x00"), event.ObjectTitle) {
			return fmt.Errorf("fremdes Filterereignis sichtbar: %s", event.ObjectTitle)
		}
	}

	return nil
}

func (s *Story66) browserCSV() error {
	return s.browser66Command(map[string]any{"download": true, "mode": "activity"})
}

func (s *Story66) browserEmptyFilter() error {
	s.query = url.Values{"object": {"nicht-vorhandene-aufgabe"}}
	if err := s.startBrowser66(); err != nil {
		return err
	}

	return s.browser66Command(map[string]any{"filters": map[string]string{"object": s.query.Get("object")}, "mode": "activity"})
}

func (s *Story66) browserEmpty() error {
	if !strings.Contains(strings.ToLower(s.page.Text), "keine aktivität") {
		return fmt.Errorf("Leerzustand fehlt: %+v", s.page)
	}

	return nil
}

func (s *Story66) browserEmptyCSV() error {
	if err := s.browserCSV(); err != nil {
		return err
	}

	return s.headerOnly()
}

func (s *Story66) startBrowser66() error {
	s.stopBrowser()
	script, err := filepath.Abs("story66_browser.mjs")
	if err != nil {
		return err
	}

	s.browser = exec.Command("node", script)
	s.browser.Stderr = os.Stderr
	s.browser.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if err := s.browserPipes(); err != nil {
		return err
	}

	path := "/organisationen/" + s.organizations["Nordstern"] + "/aktivitaet"
	return s.browserNavigate(path, "activity")
}

func (s *Story66) browser66Command(command map[string]any) error {
	data, _ := json.Marshal(command)
	if _, err := s.input.Write(append(data, '\n')); err != nil {
		return err
	}

	if err := s.input.Flush(); err != nil {
		return err
	}

	if !s.output.Scan() {
		return fmt.Errorf("Chrome endete: %v", s.output.Err())
	}

	var reply BrowserReply
	if err := json.Unmarshal(s.output.Bytes(), &reply); err != nil {
		return err
	}

	if !reply.OK {
		return fmt.Errorf("Chrome: %s", reply.Error)
	}

	s.page = reply.Page
	if reply.Download == "" {
		return nil
	}

	s.csvBody = []byte(reply.Download)
	rows, err := csv.NewReader(strings.NewReader(reply.Download)).ReadAll()
	if err != nil {
		return err
	}

	s.csvRows = rows
	return nil
}
