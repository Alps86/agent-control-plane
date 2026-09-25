package modellzugang

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

func (s *httpE2ESuite) timedProbe() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	defer s.responses.ReleaseCompletion()
	request, err := http.NewRequestWithContext(ctx, http.MethodPost,
		s.appServer.URL+"/api/settings/modelle/codex/probe", bytes.NewReader([]byte("{}")))
	if err != nil {
		return err
	}
	request.Header.Set("Origin", s.appServer.URL)
	request.Header.Set("Content-Type", "application/json")
	return s.readTimedProbe(request)
}

func (s *httpE2ESuite) readTimedProbe(request *http.Request) error {
	response, err := s.client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	reader := bufio.NewReader(response.Body)
	part, err := s.readUntilProgress(reader)
	if err != nil {
		return err
	}
	if err := s.progressBeforeCompletion(part); err != nil {
		return err
	}
	s.responses.ReleaseCompletion()
	rest, err := io.ReadAll(reader)
	s.probeStatus, s.probeBody = response.StatusCode, part+string(rest)
	s.captureLocalEvidence()
	return err
}

func (s *httpE2ESuite) progressBeforeCompletion(part string) error {
	select {
	case <-s.responses.deltaFlushed:
	default:
		return fmt.Errorf("öffentlicher Fortschritt ging Provider-Antwortteil voraus")
	}
	select {
	case <-s.responses.completionSent:
		return fmt.Errorf("Providerabschluss ging öffentlichem Fortschritt voraus")
	default:
		return s.checkProgressFrame(part)
	}
}

func (s *httpE2ESuite) checkProgressFrame(part string) error {
	progress := 0
	for _, frame := range s.parseSSEFrames(part) {
		if frame.name != "provider_progress" {
			continue
		}
		progress++
		var data map[string]any
		if json.Unmarshal(frame.data, &data) != nil || len(data) != 2 ||
			data["state"] != "provider_delta_received" || data["requests"] != float64(2) {
			return fmt.Errorf("öffentlicher Provider-Fortschritt ist nicht redigiert")
		}
	}
	if progress != 1 || strings.Contains(part, s.issuer.accessToken) ||
		strings.Contains(part, s.issuer.accountID) || strings.Contains(part, "MS01-PROBE-OK") {
		return fmt.Errorf("öffentlicher Provider-Fortschritt fehlt oder enthält Rohdaten")
	}
	return nil
}

func (s *httpE2ESuite) readUntilProgress(reader *bufio.Reader) (string, error) {
	var body strings.Builder
	progress := false
	for {
		line, err := reader.ReadString('\n')
		body.WriteString(line)
		if err != nil {
			return "", fmt.Errorf("öffentlicher Provider-Fortschritt fehlt: %w", err)
		}
		if strings.TrimSpace(line) == "event: provider_progress" {
			progress = true
		}
		if progress && strings.TrimSpace(line) == "" {
			return body.String(), nil
		}
	}
}
