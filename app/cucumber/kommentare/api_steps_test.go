package kommentare

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	appkommentar "agentcontrolplane/app/internal/app/kommentar"
	domainkommentar "agentcontrolplane/app/internal/domain/kommentar"
	portkommentar "agentcontrolplane/app/internal/port/kommentar"
	"github.com/cucumber/godog"
)

func (f *ArtifactFixture) ResolveArtifact(_ context.Context, _, _, _ string) (portkommentar.ArtifactMetadata, error) {
	if f.missing {
		return portkommentar.ArtifactMetadata{}, errors.New("not_found")
	}
	return f.metadata, nil
}

func (s *Suite) initialize(sc *godog.ScenarioContext) {
	sc.After(s.after)
	s.registerSetup(sc)
	s.registerActions(sc)
	s.registerAssertions(sc)
	s.registerBrowser(sc)
}

func (s *Suite) registerSetup(sc *godog.ScenarioContext) {
	sc.Step(`^ich starte den lokalen Server mit einer neuen SQLite-Datenbank als Betreiberin$`, s.fresh)
	sc.Step(`^die Organisation "([^"]+)" mit dem Projekt "([^"]+)" und der Aufgabe "([^"]+)" besteht$`, s.setupTask)
	sc.Step(`^"([^"]+)" ist ein Agent von "([^"]+)" mit Lese- und Schreibfreigabe für "([^"]+)"$`, s.setupScopedAgent)
	sc.Step(`^"([^"]+)" ist ein Agent von "([^"]+)" ohne Schreibfreigabe für "([^"]+)"$`, s.setupUnscopedAgent)
	sc.Step(`^"([^"]+)" ist ein Agent von "([^"]+)"$`, s.setupAgent)
	sc.Step(`^"([^"]+)" hat "([^"]+)" an "([^"]+)" kommentiert$`, s.agentPrecomment)
	sc.Step(`^ich habe "([^"]+)" an "([^"]+)" kommentiert$`, s.operatorPrecomment)
	sc.Step(`^die weitere Aufgabe "([^"]+)" besteht im Projekt "([^"]+)" von "([^"]+)"$`, s.setupAdditionalTask)
	sc.Step(`^der vertrauenswürdige Metadatenresolver kennt "([^"]+)" als Artefakt "([^"]+)" von "([^"]+)"$`, s.setupArtifact)
	sc.Step(`^der vertrauenswürdige Metadatenresolver meldet keine Metadaten für "([^"]+)"$`, s.setupMissingArtifact)
	sc.Step(`^der vertrauenswürdige Metadatenresolver meldet ein Artefakt von "([^"]+)" für "([^"]+)"$`, s.setupForeignArtifact)
}

func (s *Suite) registerActions(sc *godog.ScenarioContext) {
	s.registerActions1(sc)
	s.registerActions2(sc)
}

func (s *Suite) registerActions1(sc *godog.ScenarioContext) {
	sc.Step(`^ich als lokale Betreiberin "([^"]*)" über die öffentliche Aufgaben-API an "([^"]+)" kommentiere$`, s.operatorComment)
	sc.Step(`^ich als lokale Betreiberin "([^"]*)" an "([^"]+)" kommentiere$`, s.operatorComment)
	sc.Step(`^"([^"]+)" über die öffentliche Agenten-Fassade "([^"]+)" an "([^"]+)" kommentiert$`, s.agentComment)
	sc.Step(`^ein Client über die öffentliche Aufgaben-API eine Betreiberkennung als Autor für einen neuen Kommentar mitsendet$`, s.spoofAuthor)
	sc.Step(`^"([^"]+)" über die öffentliche Kommentar-Fassade einen Kommentar an "([^"]+)" aus "([^"]+)" senden möchte$`, s.foreignComment)
	sc.Step(`^"([^"]+)" über die öffentliche Kommentar-Fassade einen Kommentar an "([^"]+)" senden möchte$`, s.namedComment)
	sc.Step(`^"([^"]+)" über die öffentliche Kommentar-Fassade einen Kommentar an eine unbekannte Aufgabenkennung senden möchte$`, s.unknownComment)
	sc.Step(`^"([^"]+)" über die öffentliche Agenten-Fassade den Thread von "([^"]+)" abruft$`, s.agentThread)
	sc.Step(`^ich "([^"]+)" die Schreibfreigabe für "([^"]+)" entziehe$`, s.revokeWrite)
	sc.Step(`^"([^"]+)" ihren Kommentar über die öffentliche Agenten-Fassade in "([^"]+)" ändern möchte$`, s.agentUpdate)
}

func (s *Suite) registerActions2(sc *godog.ScenarioContext) {
	sc.Step(`^"([^"]+)" diesen Kommentar über die öffentliche Agenten-Fassade ändern möchte$`, s.agentUpdateOther)
	sc.Step(`^ich als lokale Betreiberin "([^"]+)" mit einem Aufgabenbezug auf "([^"]+)" an "([^"]+)" kommentiere$`, s.taskReferenceComment)
	sc.Step(`^ich als lokale Betreiberin "([^"]+)" mit der Artefaktkennung "([^"]+)" an "([^"]+)" kommentiere$`, s.artifactComment)
	sc.Step(`^ich den Server beende und mit derselben SQLite-Datenbank erneut starte$`, s.restart)
	sc.Step(`^ich den Server mit derselben SQLite-Datenbank neu starte$`, s.restart)
	sc.Step(`^ich das Projekt "([^"]+)" archiviere$`, s.archiveProject)
	sc.Step(`^ich das Projekt "([^"]+)" wiederherstelle$`, s.restoreProject)
	sc.Step(`^"([^"]+)" ihren Kommentar in "([^"]+)" ändern möchte$`, s.agentUpdate)
}

func (s *Suite) registerAssertions(sc *godog.ScenarioContext) {
	s.registerAssertions1(sc)
	s.registerAssertions2(sc)
	s.registerAssertions3(sc)
}

func (s *Suite) registerAssertions1(sc *godog.ScenarioContext) {
	sc.Step(`^liefert die API eine stabile Kommentarkennung und die Aufgabenkennung von "([^"]+)"$`, s.createdComment)
	sc.Step(`^zeigt der Aufgabenthread "([^"]+)" mit Quelle "([^"]+)" und einem serverseitigen Zeitpunkt$`, s.threadShows)
	sc.Step(`^der Aufgabenthread zeigt "([^"]+)" mit Quelle "([^"]+)" und einem serverseitigen Zeitpunkt$`, s.threadShows)
	sc.Step(`^zeigt derselbe Aufgabenthread denselben Kommentar mit Kennung, Quelle, Zeitpunkt und Aufgabenkennung genau einmal$`, s.threadAfterRestart)
	sc.Step(`^die Kommentarantwort enthält die stabile Kennung von "([^"]+)"$`, s.commentTaskID)
	sc.Step(`^wird die Autorenangabe abgewiesen$`, s.spoofRejected)
	sc.Step(`^es entsteht kein Kommentar mit gefälschter Quelle$`, s.noSpoofedComment)
	sc.Step(`^erhalte ich einen Feldfehler für den Kommentarinhalt$`, s.contentError)
	sc.Step(`^im Aufgabenthread entsteht kein neuer Kommentar$`, s.threadEmpty)
	sc.Step(`^wird der Schreibzugriff ohne Kommentar abgewiesen$`, s.writeDenied)
}

func (s *Suite) registerAssertions2(sc *godog.ScenarioContext) {
	sc.Step(`^die Antwort enthält weder Inhalt noch Kennung einer fremden Aufgabe$`, s.noForeignLeak)
	sc.Step(`^der Aufgabenthread von "([^"]+)" bleibt unverändert$`, s.namedThreadEmpty)
	sc.Step(`^der Aufgabenthread von "([^"]+)" aus "([^"]+)" bleibt unverändert$`, s.foreignThreadEmpty)
	sc.Step(`^der Aufgabenthread von eine unbekannte Aufgabenkennung bleibt unverändert$`, s.unknownThreadEmpty)
	sc.Step(`^wird der Lesezugriff ohne Kommentarinhalt und ohne Aufgabenangaben abgewiesen$`, s.readDenied)
	sc.Step(`^wird die Änderung verweigert$`, s.writeDenied)
	sc.Step(`^der Thread zeigt weiterhin "([^"]+)" mit ursprünglicher Quelle und Zeitpunkt$`, s.originalRemains)
	sc.Step(`^wird die Änderung ohne Überschreiben verweigert$`, s.writeDenied)
	sc.Step(`^der Thread zeigt weiterhin "([^"]+)" mit Quelle "([^"]+)"$`, s.originalSourceRemains)
	sc.Step(`^zeigt der Thread Kommentar, Quellaufgabenkennung und die stabile Kennung von "([^"]+)" als verifizierten Aufgabenlink$`, s.taskReferenceShown)
}

func (s *Suite) registerAssertions3(sc *godog.ScenarioContext) {
	sc.Step(`^bleibt der Aufgabenlink unverändert am selben Kommentar$`, s.taskReferenceShownAfterRestart)
	sc.Step(`^wird der gesamte Kommentar ohne fremde Aufgabendaten abgewiesen$`, s.referenceDenied)
	sc.Step(`^zeigt der Thread die aufgelöste Artefaktkennung "([^"]+)", den Anzeigenamen "([^"]+)" und einen internen Link am Kommentar$`, s.artifactShown)
	sc.Step(`^bleibt diese typisierte Referenz unverändert am selben Kommentar$`, s.artifactAfterRestart)
	sc.Step(`^der Kommentaraufruf hat kein Artefakt angelegt oder seinen Inhalt ausgeliefert$`, s.noArtifactContent)
	sc.Step(`^wird der gesamte Kommentar ohne fremde Artefaktdaten abgewiesen$`, s.referenceDenied)
	sc.Step(`^kann ich den bestehenden Kommentar weiter im Thread lesen$`, s.archivedReadable)
	sc.Step(`^wird die Änderung ohne neuen Kommentar abgewiesen$`, s.writeDenied)
	sc.Step(`^bleibt "([^"]+)" unverändert$`, s.originalRemainsOnly)
	sc.Step(`^erscheint der neue Kommentar im Thread$`, s.newCommentShown)
}

func (s *Suite) setupTask(org, project, title string) error { return s.ensureTask(org, project, title) }
func (s *Suite) setupAgent(name, org string) error          { return s.ensureAgent(org, name) }
func (s *Suite) setupScopedAgent(name, org, project string) error {
	if err := s.ensureAgent(org, name); err != nil {
		return err
	}
	return s.scope(org, project, name, true, true)
}
func (s *Suite) setupUnscopedAgent(name, org, project string) error {
	if err := s.ensureAgent(org, name); err != nil {
		return err
	}
	return s.scope(org, project, name, true, false)
}
func (s *Suite) setupAdditionalTask(title, project, org string) error {
	return s.ensureTask(org, project, title)
}

func (s *Suite) operatorComment(content, title string) error {
	input, _ := json.Marshal(map[string]string{"content": content})
	if err := s.request("POST", s.commentPath("Nordstern", "Website", title), string(input)); err != nil {
		return err
	}
	s.content = content
	return s.captureComment()
}

func (s *Suite) captureComment() error {
	if s.response.Status != http.StatusCreated {
		return nil
	}
	if err := json.Unmarshal(s.response.Body, &s.lastComment); err != nil {
		return err
	}
	s.commentID, s.createdAt, s.sourceKind, s.sourceID = s.lastComment.ID, s.lastComment.CreatedAt, s.lastComment.SourceKind, s.lastComment.SourceID
	return nil
}

func (s *Suite) agentComment(agent, content, title string) error {
	if err := s.agentService("Nordstern", agent); err != nil {
		return err
	}
	comment, err := s.service.CreateForAgent(context.Background(), s.orgID("Nordstern"), s.taskID("Nordstern", title), appkommentar.ActionTaskComment, appkommentar.CreateInput{Content: content})
	s.lastError, s.content = err, content
	if err == nil {
		s.lastComment = s.fromDomain(comment)
		s.commentID, s.createdAt = comment.ID, comment.CreatedAt
	}
	return nil
}

func (s *Suite) fromDomain(value domainkommentar.Comment) Comment {
	data, _ := json.Marshal(value)
	var comment Comment
	_ = json.Unmarshal(data, &comment)
	return comment
}

func (s *Suite) agentPrecomment(agent, content, title string) error {
	if err := s.ensureTask("Nordstern", "Website", title); err != nil {
		return err
	}
	if err := s.setupScopedAgent(agent, "Nordstern", "Website"); err != nil {
		return err
	}
	if err := s.agentComment(agent, content, title); err != nil {
		return err
	}
	if s.lastError != nil {
		return s.lastError
	}
	return nil
}

func (s *Suite) operatorPrecomment(content, title string) error {
	if err := s.operatorComment(content, title); err != nil {
		return err
	}
	return s.status(http.StatusCreated)
}

func (s *Suite) spoofAuthor() error {
	value := `{"content":"Falsche Herkunft","source_kind":"operator","source_id":"local-operator"}`
	return s.request("POST", s.commentPath("Nordstern", "Website", "Texte prüfen"), value)
}

func (s *Suite) foreignComment(agent, title, org string) error {
	if err := s.agentService("Nordstern", agent); err != nil {
		return err
	}
	_, s.lastError = s.service.CreateForAgent(context.Background(), s.orgID("Nordstern"), s.taskID(org, title), appkommentar.ActionTaskComment, appkommentar.CreateInput{Content: "Unerlaubt"})
	return nil
}

func (s *Suite) namedComment(agent, title string) error {
	org := "Nordstern"
	if agent == "Fremd" {
		org = "Suedstern"
	}
	if err := s.agentService(org, agent); err != nil {
		return err
	}
	_, s.lastError = s.service.CreateForAgent(context.Background(), s.orgID("Nordstern"), s.taskID("Nordstern", title), appkommentar.ActionTaskComment, appkommentar.CreateInput{Content: "Unerlaubt"})
	return nil
}

func (s *Suite) unknownComment(agent string) error {
	if err := s.agentService("Nordstern", agent); err != nil {
		return err
	}
	_, s.lastError = s.service.CreateForAgent(context.Background(), s.orgID("Nordstern"), "unknown", appkommentar.ActionTaskComment, appkommentar.CreateInput{Content: "Unerlaubt"})
	return nil
}

func (s *Suite) agentThread(agent, title string) error {
	if err := s.agentService("Suedstern", agent); err != nil {
		return err
	}
	_, s.lastError = s.service.ListForAgent(context.Background(), s.orgID("Nordstern"), s.taskID("Nordstern", title))
	return nil
}

func (s *Suite) revokeWrite(agent, project string) error {
	return s.scope("Nordstern", project, agent, true, false)
}

func (s *Suite) agentUpdate(agent, content string) error {
	if err := s.agentService("Nordstern", agent); err != nil {
		return err
	}
	_, s.lastError = s.service.UpdateForAgent(context.Background(), s.orgID("Nordstern"), s.taskID("Nordstern", "Texte prüfen"), s.commentID, appkommentar.ActionTaskComment, content)
	return nil
}
func (s *Suite) agentUpdateOther(agent string) error { return s.agentUpdate(agent, "Fremde Änderung") }

func (s *Suite) taskReferenceComment(content, target, title string) error {
	ref := Reference{Type: "task", ID: s.taskID("Nordstern", target), OrganizationID: s.orgID("Nordstern")}
	if target == "Fremdaufgabe" {
		ref.ID = s.taskID("Suedstern", target)
	}
	return s.operatorCommentWithReference(content, title, ref)
}

func (s *Suite) operatorCommentWithReference(content, title string, ref Reference) error {
	reference := map[string]string{"type": ref.Type, "id": ref.ID, "organization_id": ref.OrganizationID}
	input, _ := json.Marshal(map[string]any{"content": content, "reference": reference})
	if err := s.request("POST", s.commentPath("Nordstern", "Website", title), string(input)); err != nil {
		return err
	}
	s.content = content
	return s.captureComment()
}

func (s *Suite) setupArtifact(id, name, org string) error {
	if err := s.ensureTask("Nordstern", "Website", "Texte prüfen"); err != nil {
		return err
	}
	if err := s.ensureOrganization(org); err != nil {
		return err
	}
	s.resolver = &ArtifactFixture{metadata: portkommentar.ArtifactMetadata{ID: id, TaskID: s.taskID("Nordstern", "Texte prüfen"), OrganizationID: s.orgID(org), DisplayName: name, Link: "/artefakte/" + id}}
	return nil
}
func (s *Suite) setupMissingArtifact(id string) error {
	s.resolver = &ArtifactFixture{missing: true}
	return nil
}
func (s *Suite) setupForeignArtifact(org, id string) error { return s.setupArtifact(id, "Fremd", org) }

func (s *Suite) artifactComment(content, id, title string) error {
	if err := s.agentService("Nordstern", "Mira"); err != nil {
		return err
	}
	ref := &Reference{Type: "artifact", ID: id, OrganizationID: s.orgID("Nordstern")}
	value, err := s.service.Create(context.Background(), s.orgID("Nordstern"), s.taskID("Nordstern", title), appkommentar.CreateInput{Content: content, Reference: s.toDomain(ref)})
	s.lastError, s.content = err, content
	if err == nil {
		s.lastComment = s.fromDomain(value)
		s.commentID = value.ID
	}
	return nil
}

func (s *Suite) toDomain(ref *Reference) *domainkommentar.Reference {
	return &domainkommentar.Reference{Type: ref.Type, ID: ref.ID, OrganizationID: ref.OrganizationID}
}

func (s *Suite) comments() ([]Comment, error) {
	if err := s.request("GET", s.commentPath("Nordstern", "Website", "Texte prüfen"), ""); err != nil {
		return nil, err
	}
	if err := s.status(http.StatusOK); err != nil {
		return nil, err
	}
	var list CommentList
	if err := json.Unmarshal(s.response.Body, &list); err != nil {
		return nil, err
	}
	return list.Comments, nil
}

func (s *Suite) createdComment(title string) error {
	if err := s.status(http.StatusCreated); err != nil {
		return err
	}
	if s.commentID == "" || s.lastComment.TaskID != s.taskID("Nordstern", title) {
		return fmt.Errorf("Kommentar ohne stabile Kennungen: %+v", s.lastComment)
	}
	return nil
}

func (s *Suite) threadShows(content, source string) error {
	comments, err := s.comments()
	if err != nil {
		return err
	}
	wantKind, wantID := "operator", "local-operator"
	wantName := "Betreiberin"
	if source != "Betreiberin" {
		wantKind, wantID = "agent", s.agents[s.key("Nordstern", source)]
		wantName = source
	}
	for _, item := range comments {
		if item.Content == content && item.SourceKind == wantKind && item.SourceID == wantID && item.SourceName == wantName && item.CreatedAt != "" {
			return nil
		}
	}
	return fmt.Errorf("Kommentar/Quelle/Zeit fehlt: %+v", comments)
}

func (s *Suite) threadAfterRestart() error {
	comments, err := s.comments()
	if err != nil {
		return err
	}
	count := 0
	for _, item := range comments {
		if item.ID == s.commentID && item.TaskID == s.taskID("Nordstern", "Texte prüfen") && item.CreatedAt == s.createdAt && item.SourceKind == s.sourceKind && item.SourceID == s.sourceID {
			count++
		}
	}
	if count != 1 {
		return fmt.Errorf("Kommentar nach Neustart %d-mal: %+v", count, comments)
	}
	return nil
}

func (s *Suite) commentTaskID(title string) error {
	if s.lastError != nil {
		return s.lastError
	}
	if s.lastComment.ID == "" || s.lastComment.TaskID != s.taskID("Nordstern", title) {
		return fmt.Errorf("Aufgabenbezug fehlt: %+v", s.lastComment)
	}
	return nil
}

func (s *Suite) spoofRejected() error { return s.status(http.StatusBadRequest) }

func (s *Suite) noSpoofedComment() error {
	comments, err := s.comments()
	if err != nil {
		return err
	}
	for _, item := range comments {
		if item.SourceKind != "agent" || item.SourceID != s.agents[s.key("Nordstern", "Mira")] {
			return fmt.Errorf("Falsche Herkunft: %+v", item)
		}
	}
	return nil
}

func (s *Suite) contentError() error {
	if err := s.status(http.StatusUnprocessableEntity); err != nil {
		return err
	}
	var result FieldError
	if err := json.Unmarshal(s.response.Body, &result); err != nil {
		return err
	}
	if result.FieldErrors["content"] == "" {
		return fmt.Errorf("Inhaltsfehler fehlt: %s", s.response.Body)
	}
	return nil
}

func (s *Suite) threadEmpty() error {
	comments, err := s.comments()
	if err != nil {
		return err
	}
	if len(comments) != 0 {
		return fmt.Errorf("Unerwartete Kommentare: %+v", comments)
	}
	return nil
}

func (s *Suite) namedThreadEmpty(_ string) error { return s.threadEmpty() }

func (s *Suite) foreignThreadEmpty(title, org string) error {
	if err := s.request("GET", s.commentPath(org, "Fremd", title), ""); err != nil {
		return err
	}
	if err := s.status(http.StatusOK); err != nil {
		return err
	}
	var list CommentList
	if err := json.Unmarshal(s.response.Body, &list); err != nil {
		return err
	}
	if len(list.Comments) != 0 {
		return fmt.Errorf("Fremder Thread verändert: %+v", list.Comments)
	}
	return nil
}

func (s *Suite) unknownThreadEmpty() error { return s.threadEmpty() }

func (s *Suite) writeDenied() error {
	if s.lastError != nil {
		if errors.Is(s.lastError, appkommentar.ErrNotFound) || errors.Is(s.lastError, appkommentar.ErrActionDenied) {
			return nil
		}
		return fmt.Errorf("Falscher Verweigerungsfehler: %w", s.lastError)
	}
	if s.response.Status == http.StatusForbidden || s.response.Status == http.StatusNotFound {
		return nil
	}
	return fmt.Errorf("Schreiben nicht verweigert: HTTP %d %s", s.response.Status, s.response.Body)
}

func (s *Suite) noForeignLeak() error {
	if s.lastError != nil && strings.Contains(s.lastError.Error(), "Fremdaufgabe") {
		return fmt.Errorf("Fremde Daten im Fehler: %w", s.lastError)
	}
	data := string(s.response.Body)
	if strings.Contains(data, "Fremdaufgabe") || strings.Contains(data, s.taskID("Suedstern", "Fremdaufgabe")) {
		return fmt.Errorf("Fremde Daten: %s", data)
	}
	return nil
}

func (s *Suite) readDenied() error {
	if !errors.Is(s.lastError, appkommentar.ErrNotFound) {
		return fmt.Errorf("Fremder Thread nicht datenfrei verweigert: %v", s.lastError)
	}
	return nil
}

func (s *Suite) originalRemains(content string) error {
	comments, err := s.comments()
	if err != nil {
		return err
	}
	for _, item := range comments {
		if item.ID == s.commentID && item.Content == content && item.CreatedAt == s.createdAt {
			return nil
		}
	}
	return fmt.Errorf("Ursprünglicher Kommentar verändert: %+v", comments)
}

func (s *Suite) originalSourceRemains(content, source string) error {
	if err := s.originalRemains(content); err != nil {
		return err
	}
	return s.threadShows(content, source)
}

func (s *Suite) taskReferenceShown(title string) error {
	comments, err := s.comments()
	if err != nil {
		return err
	}
	for _, item := range comments {
		if item.ID == s.commentID && item.TaskID == s.taskID("Nordstern", "Texte prüfen") && item.Reference != nil && item.Reference.Type == "task" && item.Reference.ID == s.taskID("Nordstern", title) {
			return nil
		}
	}
	return fmt.Errorf("Aufgabenlink fehlt: %+v", comments)
}

func (s *Suite) taskReferenceShownAfterRestart() error { return s.taskReferenceShown("Freigabe") }

func (s *Suite) referenceDenied() error {
	if s.lastError != nil {
		if errors.Is(s.lastError, appkommentar.ErrInvalidReference) {
			return nil
		}
		return fmt.Errorf("Falscher Referenzfehler: %w", s.lastError)
	}
	if s.response.Status == http.StatusUnprocessableEntity || s.response.Status == http.StatusNotFound {
		return s.noForeignLeak()
	}
	return fmt.Errorf("Referenz angenommen: HTTP %d %s", s.response.Status, s.response.Body)
}

func (s *Suite) artifactShown(id, name string) error {
	if s.lastError != nil {
		return s.lastError
	}
	comments, err := s.comments()
	if err != nil {
		return err
	}
	for _, item := range comments {
		if item.ID == s.commentID && item.Reference != nil && item.Reference.ID == id && item.Reference.DisplayName == name && strings.HasPrefix(item.Reference.Link, "/") {
			return nil
		}
	}
	return fmt.Errorf("Aufgelöste Artefaktmetadaten fehlen: %+v", comments)
}

func (s *Suite) artifactAfterRestart() error { return s.artifactShown("entwurf-7", "Textentwurf") }

func (s *Suite) noArtifactContent() error {
	data := string(s.response.Body)
	if strings.Contains(data, "download") || strings.Contains(data, "file_content") {
		return fmt.Errorf("Artefaktinhalt ausgeliefert: %s", data)
	}
	return nil
}

func (s *Suite) archivedReadable() error                  { return s.originalRemains("Erster Entwurf") }
func (s *Suite) originalRemainsOnly(content string) error { return s.originalRemains(content) }

func (s *Suite) newCommentShown() error {
	comments, err := s.comments()
	if err != nil {
		return err
	}
	for _, item := range comments {
		if item.Content == "Noch eine Prüfung" {
			return nil
		}
	}
	return fmt.Errorf("Kommentar nach Restore fehlt: %+v", comments)
}

func (s *Suite) archiveProject(project string) error {
	return s.projectArchiveAction(project, "archivieren", "archived")
}
func (s *Suite) restoreProject(project string) error {
	return s.projectArchiveAction(project, "wiederherstellen", "active")
}

func (s *Suite) projectArchiveAction(project, action, want string) error {
	path := "/api/organisationen/" + s.orgID("Nordstern") + "/projekte/" + s.projects[s.key("Nordstern", project)] + "/" + action
	if err := s.request("POST", path, ""); err != nil {
		return err
	}
	if err := s.status(http.StatusOK); err != nil {
		return err
	}
	var response struct {
		Status string `json:"status"`
	}
	if err := json.Unmarshal(s.response.Body, &response); err != nil {
		return err
	}
	if response.Status != want {
		return fmt.Errorf("Projektstatus %q statt %q: %s", response.Status, want, s.response.Body)
	}
	return nil
}
