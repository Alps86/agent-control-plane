package modellzugang

import (
	"encoding/json"
	"fmt"
	"strings"
)

func (s *httpE2ESuite) frames() []e2eSSEFrame {
	return s.parseSSEFrames(s.probeBody)
}

func (s *httpE2ESuite) parseSSEFrames(body string) []e2eSSEFrame {
	var frames []e2eSSEFrame
	for _, block := range strings.Split(body, "\n\n") {
		frame := e2eSSEFrame{}
		for _, line := range strings.Split(block, "\n") {
			if strings.HasPrefix(line, "event: ") {
				frame.name = strings.TrimPrefix(line, "event: ")
			}
			if strings.HasPrefix(line, "data: ") {
				frame.data = []byte(strings.TrimPrefix(line, "data: "))
			}
		}
		if frame.name != "" && len(frame.data) != 0 {
			frames = append(frames, frame)
		}
	}
	return frames
}

func (s *httpE2ESuite) providerMatches(want bool) error {
	providers := 0
	for _, frame := range s.frames() {
		if frame.name != "provider" {
			continue
		}
		providers++
		if err := s.checkProviderFrame(frame, providers, want); err != nil {
			return err
		}
	}
	if providers != 2 {
		return fmt.Errorf("erwartet zwei redigierte Providerbeobachtungen, erhalten %d", providers)
	}
	return s.noProbeSecrets()
}

func (s *httpE2ESuite) checkProviderFrame(frame e2eSSEFrame, index int, want bool) error {
	var data e2eProviderEvent
	if json.Unmarshal(frame.data, &data) != nil || data.ProviderModelMatchesConfig == nil ||
		*data.ProviderModelMatchesConfig != want || data.Model != "local-e2e-model" ||
		data.RequestID != fmt.Sprintf("redacted-request-%d", index) {
		return fmt.Errorf("Providerbeobachtung %d hat falschen Modellvergleich oder unredigierte Metadaten", index)
	}
	return nil
}

func (s *httpE2ESuite) fakeModelMismatch() error {
	metadata := s.responses.Metadata()
	if len(metadata) != 2 {
		return fmt.Errorf("zwei Gegenstellen-Modelle fehlen")
	}
	if s.responses.mode == "missing" && (!metadata[0].modelPresent && metadata[0].model == "" &&
		metadata[1].modelPresent && metadata[1].model == "") {
		return nil
	}
	if s.responses.mode == "mismatch" && metadata[0].model == "other-local-model" &&
		metadata[1].model == "other-local-model" {
		return nil
	}
	if s.responses.mode == "echo" && metadata[0].model == s.issuer.accessToken &&
		metadata[1].model == s.issuer.accessToken {
		return nil
	}
	return fmt.Errorf("Gegenstelle sendete nicht das kontrollierte abweichende Modell")
}

func (s *httpE2ESuite) finalEvidence(answer, chunk bool) error {
	for _, frame := range s.frames() {
		if frame.name != "done" {
			continue
		}
		var data e2eFinalEvent
		if json.Unmarshal(frame.data, &data) != nil || data.AnswerMatchesExpected == nil ||
			data.ChunkBeforeResponseCompleted == nil || *data.AnswerMatchesExpected != answer ||
			*data.ChunkBeforeResponseCompleted != chunk || data.State != "completed" || data.Requests != 2 {
			return fmt.Errorf("redigierter Abschlussbeleg ist unvollständig")
		}
		return s.noProbeSecrets()
	}
	return fmt.Errorf("SSE-Abschluss fehlt")
}

func (s *httpE2ESuite) errorEvidence(answer, chunk bool) error {
	for _, frame := range s.frames() {
		if frame.name != "error" {
			continue
		}
		var data e2eFinalEvent
		if json.Unmarshal(frame.data, &data) != nil || data.AnswerMatchesExpected == nil ||
			data.ChunkBeforeResponseCompleted == nil || *data.AnswerMatchesExpected != answer ||
			*data.ChunkBeforeResponseCompleted != chunk || data.Requests != 2 {
			return fmt.Errorf("redigierter Fehlerbeleg ist unvollständig")
		}
		return s.noProbeSecrets()
	}
	return fmt.Errorf("SSE-Fehler fehlt")
}
