package modellwahl

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/cucumber/godog"
	"net/http"
	"strings"
)

func (s *Suite) initialize(sc *godog.ScenarioContext) {
	sc.After(s.after)
	s.registerBase(sc)
	s.registerSelection(sc)
	s.registerChecks(sc)
	s.registerEvidence(sc)
	s.registerInvalidCatalog(sc)
	s.registerBrowser(sc)
}

func (s *Suite) registerBase(sc *godog.ScenarioContext) {
	sc.Step(`^ein lokaler Server mit kontrollierten Modellkatalogen läuft$`, s.fresh)
	sc.Step(`^die Organisation "Nord" mit dem Eino-Agenten "Mira" besteht$`, s.existing)
	sc.Step(`^die zentrale OpenRouter-Verbindung ist einsatzbereit und für "Nord" und "Mira" freigegeben$`, s.openRouterGranted)
	sc.Step(`^die zentrale OpenRouter-Verbindung ist einsatzbereit$`, s.openRouterReady)
	sc.Step(`^ihre Freigabe für (Organisation|Agent) fehlt$`, s.missingGrant)
}

func (s *Suite) registerSelection(sc *godog.ScenarioContext) {
	sc.Step(`^ich Miras Modellwahl über HTTP abrufe$`, s.readChoice)
	sc.Step(`^ich für "Mira" die OpenRouter-Verbindung und das Modell "openai/gpt-4" über HTTP speichere$`, s.chooseOpenRouter)
	sc.Step(`^ich für "Mira" den unbekannten Anbieter "fremd" über HTTP speichere$`, s.chooseUnknownProvider)
	sc.Step(`^ich für "Mira" die fremde Verbindung "andere-verbindung" über HTTP speichere$`, s.chooseUnknownConnection)
	sc.Step(`^ich für "Mira" das nicht katalogisierte Modell "fremd/neu" über HTTP speichere$`, s.chooseUnknownModel)
	sc.Step(`^ich Miras Modellwahl aus einer fremden Organisation über HTTP lese und ändere$`, s.crossOrganization)
}

func (s *Suite) registerChecks(sc *godog.ScenarioContext) {
	sc.Step(`^ist "Codex-Abo" als Standardanbieter ausgewiesen$`, s.codexDefault)
	sc.Step(`^die Ausführungsart von "Mira" ist "Eino"$`, s.agentEino)
	sc.Step(`^Verbindungskennungen und Modellkennungen stammen aus dem registrierten Katalog$`, s.catalogOnly)
	sc.Step(`^das Codex-Modell ist ohne Konto-Nachweis als ungeprüft markiert$`, s.codexUnverified)
	sc.Step(`^der Katalog zeigt nur nicht geheime Verbindungsreferenzen$`, s.catalogReferencesOnly)
	sc.Step(`^zeigt Miras öffentliche Modellwahlansicht "OpenRouter" und "openai/gpt-4"$`, s.agentChoice)
	sc.Step(`^die Ausführungsart von "Mira" bleibt "Eino"$`, s.agentEino)
	sc.Step(`^die öffentliche Modellwahlantwort enthält den hinterlegten Schlüssel nicht$`, s.noSecrets)
	sc.Step(`^die Modellwahl bleibt nach einem Serverneustart erhalten$`, s.choicePersists)
	sc.Step(`^wird die Modellwahl mit einem Hinweis auf die fehlende (Organisation|Agent) abgelehnt$`, s.grantDenied)
	sc.Step(`^die öffentliche Antwort hat HTTP-Status (\d+)$`, s.status)
	sc.Step(`^Miras bisherige Modellwahl bleibt erhalten$`, s.choiceUnchanged)
	sc.Step(`^wird die Modellwahl mit einem verständlichen Kataloghinweis abgelehnt$`, s.catalogDenied)
	sc.Step(`^erhalte ich keine Modellwahl oder Katalogdaten von "Mira"$`, s.crossHidden)
	sc.Step(`^die fremde Antwort entspricht einer unbekannten Agentenkennung mit HTTP 404$`, s.sameUnknown)
}

func (s *Suite) registerEvidence(sc *godog.ScenarioContext) {
	sc.Step(`^nennt der Katalog für OpenRouter seine konfigurierte Authentifizierungsart$`, s.authentication)
	sc.Step(`^er zeigt für "openai/gpt-4" Modellquelle, Stand und Prüfstatus$`, s.modelEvidence)
	sc.Step(`^er zeigt für Textfähigkeit Quelle, Prüfzeit und Status$`, s.textCapabilityEvidence)
	sc.Step(`^er führt Text, Tools, Audioeingabe, Audioausgabe und Echtzeit einzeln auf$`, s.capabilityRows)
	sc.Step(`^nicht belegte Echtzeit-Sprache wird als "Nicht nachgewiesen" gekennzeichnet$`, s.realtimeUnverified)
}

func (s *Suite) existing() error {
	if s.orgID != "" {
		return nil
	}
	org := Organization{}
	if err := s.create(`{"name":"Nord","description":"Test"}`, "/api/organisationen", &org); err != nil {
		return err
	}
	s.orgID = org.ID
	agent := Agent{}
	path := "/api/organisationen/" + s.orgID + "/agenten"
	if err := s.create(`{"name":"Mira","template_id":"recherche","execution_kind":"eino"}`, path, &agent); err != nil {
		return err
	}
	s.agentID = agent.ID
	return s.captureInitial()
}

func (s *Suite) readChoice() error { return s.request("GET", s.choicePath(), "") }
func (s *Suite) choose(provider, connection, model string) error {
	body, _ := json.Marshal(map[string]string{"provider_id": provider, "connection_id": connection, "model_id": model})
	return s.request("PUT", s.choicePath(), string(body))
}
func (s *Suite) chooseOpenRouter() error {
	return s.choose("openrouter", "openrouter-central", "openai/gpt-4")
}
func (s *Suite) chooseUnknownProvider() error {
	return s.choose("fremd", "openrouter-central", "openai/gpt-4")
}
func (s *Suite) chooseUnknownConnection() error {
	return s.choose("openrouter", "andere-verbindung", "openai/gpt-4")
}
func (s *Suite) chooseUnknownModel() error {
	return s.choose("openrouter", "openrouter-central", "fremd/neu")
}
func (s *Suite) status(code int) error {
	if s.last.Status != code {
		return fmt.Errorf("HTTP %d statt %d: %s", s.last.Status, code, s.last.Body)
	}
	return nil
}

func (s *Suite) bodyContains(parts ...string) error {
	for _, part := range parts {
		if !bytes.Contains(bytes.ToLower(s.last.Body), []byte(strings.ToLower(part))) {
			return fmt.Errorf("%q fehlt in HTTP %d: %s", part, s.last.Status, s.last.Body)
		}
	}
	return nil
}
func (s *Suite) view() (ChoiceView, error) {
	view := ChoiceView{}
	if s.last.Status != 200 {
		return view, fmt.Errorf("Modellwahl HTTP %d: %s", s.last.Status, s.last.Body)
	}
	err := json.Unmarshal(s.last.Body, &view)
	return view, err
}
func (s *Suite) provider(id string) (Provider, error) {
	view, err := s.view()
	if err != nil {
		return Provider{}, err
	}
	for _, provider := range view.Providers {
		if provider.ID == id {
			return provider, nil
		}
	}
	return Provider{}, fmt.Errorf("Anbieter %q fehlt", id)
}
func (s *Suite) codexDefault() error {
	view, err := s.view()
	if err != nil {
		return err
	}
	if view.Selection.Provider != "codex-abo" {
		return fmt.Errorf("Standardanbieter %q", view.Selection.Provider)
	}
	return nil
}
func (s *Suite) catalogOnly() error {
	provider, err := s.provider("codex-abo")
	if err != nil {
		return err
	}
	if len(provider.Connections) == 0 || provider.Connections[0].Reference != "codex-chatgpt" {
		return fmt.Errorf("Codex-Referenz fehlt")
	}
	if len(provider.Models) == 0 || provider.Models[0].ID != "gpt-5.3-codex" {
		return fmt.Errorf("Codex-Katalogmodell fehlt")
	}
	openRouter, err := s.provider("openrouter")
	if err != nil {
		return err
	}
	if len(openRouter.Connections) != 1 || openRouter.Connections[0].Reference != "openrouter-central" {
		return fmt.Errorf("explizite OpenRouter-Referenz fehlt: %+v", openRouter.Connections)
	}
	return nil
}
func (s *Suite) codexUnverified() error {
	provider, err := s.provider("codex-abo")
	if err != nil {
		return err
	}
	if len(provider.Models) == 0 || provider.Models[0].CheckStatus != "unverified" || provider.Models[0].Source != "https://developers.openai.com/api/docs/models/all" || provider.Models[0].ObservedAt != "2026-09-24T23:55:13Z" {
		return fmt.Errorf("Codex-Modell als geprüft behauptet")
	}
	return nil
}
func (s *Suite) catalogReferencesOnly() error {
	view, err := s.view()
	if err != nil {
		return err
	}
	for _, provider := range view.Providers {
		for _, connection := range provider.Connections {
			if connection.Reference == "" {
				return fmt.Errorf("leere Verbindungsreferenz")
			}
		}
	}
	for _, field := range []string{`"secret":`, `"key":`, `"access_token":`, `"refresh_token":`} {
		if bytes.Contains(bytes.ToLower(s.last.Body), []byte(field)) {
			return fmt.Errorf("geheimes Katalogfeld %s", field)
		}
	}
	return nil
}
func (s *Suite) noSecrets() error {
	if !s.secretStored {
		return fmt.Errorf("Testschlüssel wurde nicht öffentlich hinterlegt; Redaction unbelegt")
	}
	if bytes.Contains(s.last.Body, []byte(s.secret)) {
		return fmt.Errorf("Zugangsdaten in öffentlicher Antwort")
	}
	return nil
}
func (s *Suite) agentEino() error {
	if err := s.readChoice(); err != nil {
		return err
	}
	view, err := s.view()
	if err != nil {
		return err
	}
	if view.ExecutionKind != "eino" {
		return fmt.Errorf("Ausführungsart %q", view.ExecutionKind)
	}
	return nil
}
func (s *Suite) agentChoice() error {
	if err := s.readChoice(); err != nil {
		return err
	}
	view, err := s.view()
	if err != nil {
		return err
	}
	if view.Selection != (Selection{Provider: "openrouter", Connection: "openrouter-central", Model: "openai/gpt-4"}) {
		return fmt.Errorf("gewählte Modellroute falsch: %+v", view.Selection)
	}
	return nil
}
func (s *Suite) choicePersists() error {
	if err := s.restart(); err != nil {
		return err
	}
	if err := s.readChoice(); err != nil {
		return err
	}
	return s.agentChoice()
}
func (s *Suite) choiceUnchanged() error {
	if err := s.readChoice(); err != nil {
		return err
	}
	if s.last.Status != 200 {
		return fmt.Errorf("Modellwahl HTTP %d", s.last.Status)
	}
	initial, current := ChoiceView{}, ChoiceView{}
	if err := json.Unmarshal(s.initial, &initial); err != nil {
		return err
	}
	if err := json.Unmarshal(s.last.Body, &current); err != nil {
		return err
	}
	if initial.Selection != current.Selection {
		return fmt.Errorf("Modellwahl verändert: %s", s.last.Body)
	}
	return nil
}
func (s *Suite) grantDenied(level string) error {
	if s.last.Status != http.StatusForbidden {
		return fmt.Errorf("%s-Freigabe: HTTP %d: %s", level, s.last.Status, s.last.Body)
	}
	return s.bodyContains(strings.ToLower(level))
}
func (s *Suite) catalogDenied() error {
	if s.last.Status != 422 {
		return fmt.Errorf("Katalog-Ablehnung HTTP %d: %s", s.last.Status, s.last.Body)
	}
	return s.bodyContains("model")
}
func (s *Suite) authentication() error {
	provider, err := s.provider("openrouter")
	if err != nil {
		return err
	}
	if provider.AuthType != "api_key" {
		return fmt.Errorf("Authentifizierung %q", provider.AuthType)
	}
	return nil
}
func (s *Suite) catalogModel() (Model, error) {
	provider, err := s.provider("openrouter")
	if err != nil {
		return Model{}, err
	}
	for _, model := range provider.Models {
		if model.ID == "openai/gpt-4" {
			return model, nil
		}
	}
	return Model{}, fmt.Errorf("Katalogmodell openai/gpt-4 fehlt")
}
func (s *Suite) modelEvidence() error {
	model, err := s.catalogModel()
	if err != nil {
		return err
	}
	if model.Source != "https://openrouter.ai/docs/api/api-reference/models/get-models" || model.ObservedAt != "2026-09-24T23:55:13Z" || model.CheckStatus != "unverified" {
		return fmt.Errorf("Modellbeleg unvollständig: %+v", model)
	}
	return nil
}
func (s *Suite) textCapabilityEvidence() error {
	model, err := s.catalogModel()
	if err != nil {
		return err
	}
	text := model.Capabilities["text"]
	if text.Status != "unverified" || text.Source != "https://openrouter.ai/docs/api/api-reference/models/get-models" || text.CheckedAt != "2026-09-24T23:55:13Z" {
		return fmt.Errorf("Textbeleg unvollständig: %+v", text)
	}
	return nil
}
func (s *Suite) capabilityRows() error {
	model, err := s.catalogModel()
	if err != nil {
		return err
	}
	for _, name := range []string{"text", "tools", "audio_input", "audio_output", "realtime"} {
		item, present := model.Capabilities[name]
		if !present || item.Status != "unverified" || item.Source != "https://openrouter.ai/docs/api/api-reference/models/get-models" || item.CheckedAt != "2026-09-24T23:55:13Z" {
			return fmt.Errorf("Fähigkeit %s ohne Status/Quelle/Zeit: %+v", name, item)
		}
	}
	return nil
}
func (s *Suite) realtimeUnverified() error {
	model, err := s.catalogModel()
	if err != nil {
		return err
	}
	if model.Capabilities["realtime"].Status != "unverified" {
		return fmt.Errorf("Echtzeit fälschlich belegt")
	}
	return nil
}

func (s *Suite) crossOrganization() error {
	other := Organization{}
	if err := s.create(`{"name":"Sued","description":"Test"}`, "/api/organisationen", &other); err != nil {
		return err
	}
	path := "/api/organisationen/" + other.ID + "/agenten/" + s.agentID + "/modellwahl"
	if err := s.request("GET", path, ""); err != nil {
		return err
	}
	s.crossRead = s.last
	if err := s.request("PUT", path, `{"provider_id":"openrouter","connection_id":"openrouter-central","model_id":"openai/gpt-4"}`); err != nil {
		return err
	}
	s.crossWrite = s.last
	return s.request("GET", "/api/organisationen/"+s.orgID+"/agenten/unknown-agent/modellwahl", "")
}

func (s *Suite) crossHidden() error {
	if s.crossRead.Status != 404 || s.crossWrite.Status != 404 {
		return fmt.Errorf("fremde Modellwahl GET %d, PUT %d", s.crossRead.Status, s.crossWrite.Status)
	}
	for _, response := range []Response{s.crossRead, s.crossWrite} {
		for _, private := range []string{"Mira", "openai/gpt-4", "codex-chatgpt"} {
			if bytes.Contains(response.Body, []byte(private)) {
				return fmt.Errorf("fremde Antwort verrät %q", private)
			}
		}
	}
	return nil
}

func (s *Suite) sameUnknown() error {
	if s.last.Status != 404 {
		return fmt.Errorf("unbekannter Agent HTTP %d", s.last.Status)
	}
	if !bytes.Equal(s.crossRead.Body, s.last.Body) {
		return fmt.Errorf("fremde und unbekannte Antwort unterscheiden sich")
	}
	return nil
}

func (s *Suite) openRouterReady() error {
	return fmt.Errorf("Story-24-Prozesssetup wartet auf öffentliche Verbindungs-/Freigabe-API; kein Anbieteraufruf erfolgt")
}

func (s *Suite) openRouterGranted() error    { return s.openRouterReady() }
func (s *Suite) missingGrant(_ string) error { return s.openRouterReady() }
