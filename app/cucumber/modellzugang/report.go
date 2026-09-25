package modellzugang

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"agentcontrolplane/app/internal/adapter/model/codexabo"
)

func (s *Suite) syncEvidence() error {
	s.nachweis.ErfasseTransport(s.transportBeleg)
	s.appendReportRequests()
	return s.persistReport()
}

func (s *Suite) appendReportRequests() {
	responses := s.observations.Entries()
	for _, request := range s.trace.Requests() {
		entry := reportRequest{RequestID: request.requestID, Model: os.Getenv("MS01_LIVE_MODEL"), HTTPStatus: request.status}
		for _, response := range responses {
			if response.RequestID == request.requestID {
				entry.Model, entry.ProviderStatus, entry.Usage = response.Model, response.Status, response.Usage
			}
		}

		s.reportRequests = append(s.reportRequests, entry)
	}
}

func (s *Suite) persistReport() error {
	accountID := os.Getenv("MS01_LIVE_ACCOUNT_ID")
	hash := sha256.Sum256([]byte(accountID))
	artifact := reportArtifact{Story: "Story-08/MS-01", Execution: "Eino", Provider: "Codex-Abo",
		AccountSHA256: hex.EncodeToString(hash[:]), ObservedUTC: time.Now().UTC().Format(time.RFC3339),
		GateOpen: s.nachweis.Status().Offen, TransportEvidence: s.transportBeleg, Requests: s.reportRequests}
	data, err := json.MarshalIndent(artifact, "", "  ")
	if err != nil {
		return err
	}

	return s.writeRedacted(data, accountID)
}

func (s *Suite) writeRedacted(data []byte, accountID string) error {
	token := os.Getenv("MS01_LIVE_TOKEN")
	if token == "" || accountID == "" || bytes.Contains(data, []byte(token)) || bytes.Contains(data, []byte(accountID)) {
		return fmt.Errorf("Prüfbericht enthält Zugangsdaten oder hat keine Live-Verbindung")
	}

	path := s.reportPath
	if path == "" {
		path = filepath.Join("..", "..", "..", "spec", "nachweise", "story-08", "live-report.json")
	}

	return s.replaceReport(path, data)
}

func (s *Suite) replaceReport(path string, data []byte) error {
	dir := filepath.Dir(path)
	file, err := os.CreateTemp(dir, ".live-report-*")
	if err != nil {
		return err
	}

	defer os.Remove(file.Name())
	defer file.Close()
	if err := s.writeReportTemp(file, data); err != nil {
		return err
	}
	if err := s.checkReportTarget(path); err != nil {
		return err
	}
	if err := os.Rename(file.Name(), path); err != nil {
		return err
	}
	return s.syncReportDirectory(dir)
}

func (s *Suite) writeReportTemp(file *os.File, data []byte) error {
	if err := file.Chmod(0600); err != nil {
		return err
	}
	if _, err := file.Write(data); err != nil {
		return err
	}
	if err := file.Sync(); err != nil {
		return err
	}
	return file.Close()
}

func (s *Suite) checkReportTarget(path string) error {
	info, err := os.Lstat(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("Prüfberichtziel ist keine reguläre Datei")
	}
	return nil
}

func (s *Suite) syncReportDirectory(dir string) error {
	directory, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer directory.Close()
	return directory.Sync()
}

func (s *Suite) responseID() string {
	for _, response := range s.observations.Entries() {
		if response.RequestID != "" {
			return response.RequestID
		}
	}

	return ""
}

func (s *Suite) observedModel() string {
	for _, response := range s.observations.Entries() {
		if response.Model != "" {
			return response.Model
		}
	}

	return ""
}

func (s *Suite) failedObservation() *codexabo.Observation {
	for _, response := range s.observations.Entries() {
		if response.Status != "completed" {
			copy := response
			return &copy
		}
	}

	return nil
}
