package modellzugang

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"agentcontrolplane/app/internal/app/modellpruefung"
)

func (s *Suite) localReportSetup() error {
	s.local = true
	s.t.Setenv("MS01_LIVE_TOKEN", "offline-secret-token")
	s.t.Setenv("MS01_LIVE_ACCOUNT_ID", "offline-secret-account")
	s.answer = "geheimer Prompt aus dem Test"
	s.reportPath = filepath.Join(s.t.TempDir(), "report.json")
	s.nachweis = modellpruefung.NewNachweis()
	s.transportBeleg = modellpruefung.Transportbeleg{Anbieter: "lokaler Testanbieter", Modell: "testmodell",
		Anfragekennung: "offline-request-1", Antwort: true, Stream: true}
	s.reportRequests = []reportRequest{{RequestID: "offline-request-1", Model: "testmodell", HTTPStatus: 200}}
	return nil
}

func (s *Suite) localReportWrite() error {
	return s.persistReport()
}

func (s *Suite) localReportRedacted() error {
	data, err := os.ReadFile(s.reportPath)
	if err != nil {
		return err
	}

	if !bytes.Contains(data, []byte("offline-request-1")) || !bytes.Contains(data, []byte("testmodell")) ||
		bytes.Contains(data, []byte("offline-secret-token")) || bytes.Contains(data, []byte("offline-secret-account")) ||
		bytes.Contains(data, []byte("geheimer Prompt")) {
		return fmt.Errorf("lokaler Bericht enthält fehlende oder unredigierte Daten")
	}

	return nil
}

func (s *Suite) localReportMode() error {
	info, err := os.Stat(s.reportPath)
	if err != nil {
		return err
	}

	if info.Mode().Perm() != 0600 {
		return fmt.Errorf("Prüfberichtrechte %04o statt 0600", info.Mode().Perm())
	}

	return nil
}

func (s *Suite) localExistingReport() error {
	if err := s.localReportSetup(); err != nil {
		return err
	}
	return os.WriteFile(s.reportPath, []byte("alter Bericht"), 0600)
}

func (s *Suite) localReplacedReport() error {
	data, err := os.ReadFile(s.reportPath)
	if err != nil {
		return err
	}
	if !bytes.Contains(data, []byte("offline-request-1")) || bytes.Contains(data, []byte("alter Bericht")) {
		return fmt.Errorf("Bericht wurde nicht vollständig ersetzt")
	}
	return s.localReportMode()
}

func (s *Suite) localSymlinkReport() error {
	if err := s.localReportSetup(); err != nil {
		return err
	}
	s.reportVictim = filepath.Join(filepath.Dir(s.reportPath), "victim.json")
	if err := os.WriteFile(s.reportVictim, []byte("unverändert"), 0600); err != nil {
		return err
	}
	return os.Symlink(s.reportVictim, s.reportPath)
}

func (s *Suite) localRejectedReport() error {
	s.reportErr = s.persistReport()
	return nil
}

func (s *Suite) localSymlinkUnchanged() error {
	info, err := os.Lstat(s.reportPath)
	if err != nil || info.Mode()&os.ModeSymlink == 0 || s.reportErr == nil {
		return fmt.Errorf("Symlink-Ziel wurde nicht abgelehnt: %v", s.reportErr)
	}
	data, err := os.ReadFile(s.reportVictim)
	if err != nil || string(data) != "unverändert" {
		return fmt.Errorf("Symlink-Ziel wurde verändert: %v", err)
	}
	return nil
}

func (s *Suite) localNoReportTemp() error {
	entries, err := os.ReadDir(filepath.Dir(s.reportPath))
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if len(entry.Name()) >= len(".live-report-") && entry.Name()[:len(".live-report-")] == ".live-report-" {
			return fmt.Errorf("temporäre Berichtsdatei blieb zurück")
		}
	}
	return nil
}
