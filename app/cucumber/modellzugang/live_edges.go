package modellzugang

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"agentcontrolplane/app/internal/adapter/model/codexabo"
	"agentcontrolplane/app/internal/app/modellpruefung"
	"github.com/cloudwego/eino/schema"
	"github.com/cucumber/godog"
)

func (s *Suite) pendingStream(agent string) error {
	if agent != "Mira" {
		return fmt.Errorf("unerwarteter Agent")
	}

	s.pruefung.Erlaube(agent, "Website")
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	s.cancel = cancel
	stream, err := s.agent.Stream(ctx, []*schema.Message{schema.UserMessage("Lies den Status von Website mit der Fachaktion.")})
	if err != nil {
		cancel()
		return err
	}

	s.streamDone = make(chan struct{})
	s.firstChunk = make(chan struct{}, 1)
	go s.consumeAndClose(stream)
	return s.awaitFirstChunk()
}

func (s *Suite) consumeAndClose(stream *schema.StreamReader[*schema.Message]) {
	defer close(s.streamDone)
	defer stream.Close()
	s.consume(stream)
}

func (s *Suite) awaitFirstChunk() error {
	select {
	case <-s.firstChunk:
		return nil
	case <-s.streamDone:
		return fmt.Errorf("Providerstream endete vor dem Abbruchfenster")
	case <-time.After(60 * time.Second):
		s.cancel()
		return fmt.Errorf("Providerstream lieferte keinen Antwortteil")
	}
}

func (s *Suite) pendingTool(aktion string) error {
	if aktion != modellpruefung.StatusAktion || s.fixture.Accesses() != 0 {
		return fmt.Errorf("Fachschritt bereits ausgeführt oder falsch")
	}

	return nil
}

func (s *Suite) cancelStream() error {
	s.priorRequests = len(s.trace.Requests())
	s.cancel()
	select {
	case <-s.streamDone:
		s.cancelled = true
		return nil
	case <-time.After(5 * time.Second):
		return fmt.Errorf("Modellstream reagiert nicht auf Abbruch")
	}
}

func (s *Suite) cancelReport() error {
	if !s.cancelled || s.modelErr == nil || s.chunks == 0 || !s.trace.BodyClosed() {
		return fmt.Errorf("tatsächlicher Abbruchzustand fehlt")
	}

	for _, response := range s.observations.Entries() {
		if response.Status == "completed" {
			return fmt.Errorf("Providerstream war vor dem Abbruch abgeschlossen")
		}
	}

	return nil
}

func (s *Suite) noToolAfterCancel() error {
	if s.fixture.Accesses() != 0 {
		return fmt.Errorf("ausstehender Fachschritt wurde ausgeführt")
	}

	return nil
}

func (s *Suite) noRestart() error {
	if len(s.trace.Requests()) != s.priorRequests {
		return fmt.Errorf("Modellaufruf nach Abbruch gestartet")
	}

	s.transportBeleg.Abbruch = true
	return s.syncEvidence()
}

func (s *Suite) limitProof() error {
	path := os.Getenv("MS01_LIMIT_EVIDENCE_FILE")
	if path == "" {
		return godog.ErrPending
	}

	data, err := os.ReadFile(path)
	if err != nil || json.Unmarshal(data, &s.evidence) != nil || !s.validEvidence() {
		return godog.ErrPending
	}

	s.proof = path
	return nil
}

func (s *Suite) validEvidence() bool {
	observed, err := time.Parse(time.RFC3339, s.evidence.ObservedAt)
	if err != nil || observed.After(time.Now()) || time.Since(observed) > 30*time.Minute {
		return false
	}

	accountID := os.Getenv("MS01_LIVE_ACCOUNT_ID")
	accountHash := sha256.Sum256([]byte(accountID))
	return s.evidence.Provider == "codex-abo" && s.evidence.Source == "provider_limit" && accountID != "" &&
		s.evidence.AccountRef == hex.EncodeToString(accountHash[:]) &&
		s.evidence.HTTPStatus == 429 && s.evidence.RequestID != ""
}

func (s *Suite) proofRecorded() error {
	if s.proof == "" {
		return fmt.Errorf("Limit-Vorabnachweis fehlt")
	}

	return nil
}

func (s *Suite) limitCall() error {
	return s.runLive("Antworte kurz auf diesen Prüfaufruf.")
}

func (s *Suite) providerLimited() error {
	var providerErr *codexabo.ProviderError
	if !errors.As(s.modelErr, &providerErr) || providerErr.Kind != "limit" {
		return fmt.Errorf("kein echter Anbieterlimitfehler: %v", s.modelErr)
	}

	return nil
}

func (s *Suite) limitReport() error {
	if err := s.providerLimited(); err != nil {
		return err
	}

	if s.failedObservation() == nil {
		return fmt.Errorf("redigierter Anbieterlimitbericht fehlt")
	}

	return nil
}

func (s *Suite) noAPIFallback() error {
	requests := s.trace.Requests()
	if len(requests) == 0 {
		return fmt.Errorf("kein Abo-Modellaufruf beobachtet")
	}

	for _, request := range requests {
		if request.host != "chatgpt.com" || request.path != "/backend-api/codex/responses" {
			return fmt.Errorf("unerwartete Modellroute")
		}
	}

	return nil
}

func (s *Suite) pocGateMatches() error {
	s.transportBeleg.Limitfehler = true
	if err := s.syncEvidence(); err != nil {
		return err
	}

	complete := s.transportBeleg.Antwort && s.transportBeleg.Stream && s.transportBeleg.Toolrunde &&
		s.transportBeleg.Folgeaufruf && s.transportBeleg.Abbruch && s.transportBeleg.Limitfehler
	if s.nachweis.Status().Offen == complete {
		return fmt.Errorf("App-Gate entspricht dem technischen Gesamtbeleg nicht")
	}

	return nil
}

func (s *Suite) codexOnly() error {
	s.agentOnly = modellpruefung.NewNachweis()
	s.agentOnly.ErfasseAgentenlauf()
	return nil
}

func (s *Suite) checkGate() error {
	status := s.agentOnly.Status()
	if status.Agentenlaeufe != 1 || status.Modellbelege != 0 {
		return fmt.Errorf("Agentenlauf oder Modellbeleg falsch bewertet")
	}

	return nil
}

func (s *Suite) gateStillOpen(name string) error {
	if name != "MS-01" || !s.agentOnly.Status().Offen {
		return fmt.Errorf("technisches Gate wurde falsch bewertet")
	}

	return nil
}

func (s *Suite) noAgentSubstitute() error {
	if s.agentOnly.Status().Modellbelege != 0 {
		return fmt.Errorf("Codex-Agentenlauf als Modellaufruf gezählt")
	}

	return nil
}
