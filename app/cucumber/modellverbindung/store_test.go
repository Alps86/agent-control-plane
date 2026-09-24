package modellverbindung

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"

	"agentcontrolplane/app/internal/adapter/credentials"
	"github.com/cucumber/godog"
)

const syntheticCredential = "synthetic-codex-session-private-value"

func (s *Suite) registerStoreSteps(sc *godog.ScenarioContext) {
	sc.Step(`^ein geschützter lokaler Verbindungsspeicher ist geöffnet$`, s.openFreshStore)
	sc.Step(`^ich habe Codex-Anmeldedaten geschützt gespeichert$`, s.savedCredential)
	sc.Step(`^ich die Codex-Anmeldedaten speichere$`, s.saveCredential)
	sc.Step(`^ich den Speicher mit demselben Masterschlüssel erneut öffne$`, s.reopenStore)
	sc.Step(`^ich den Verbindungsspeicher erneut öffne$`, s.reopenStore)
	sc.Step(`^der Masterschlüssel fehlt$`, s.removeKey)
	sc.Step(`^die Berechtigungen der "([^"]*)" zu weit geöffnet werden$`, s.permissiveFile)
	sc.Step(`^stehen die Anmeldedaten nicht im Klartext in der Secret-Datei$`, s.encryptedOnDisk)
	sc.Step(`^die Secret-Datei enthält keine Anmeldedaten im Klartext$`, s.encryptedOnDisk)
	sc.Step(`^erhalte ich die gespeicherten Anmeldedaten zurück$`, s.credentialSurvived)
	sc.Step(`^wird der Verbindungsspeicher nicht geöffnet$`, s.storeRejected)
}

func (s *Suite) openFreshStore() error {
	directory := filepath.Join(s.t.TempDir(), "private")
	if err := os.Mkdir(directory, 0700); err != nil {
		return err
	}

	s.secretPath = filepath.Join(directory, "connections.enc")
	s.keyPath = filepath.Join(directory, "master.key")
	opened, err := credentials.NewStore(s.secretPath, s.keyPath)
	s.store = opened
	s.openErr = err
	return s.openErr
}

func (s *Suite) savedCredential() error {
	if err := s.openFreshStore(); err != nil {
		return err
	}

	return s.saveCredential()
}

func (s *Suite) saveCredential() error {
	return s.store.Save(context.Background(), "codex-subscription", []byte(syntheticCredential))
}

func (s *Suite) reopenStore() error {
	opened, err := credentials.NewStore(s.secretPath, s.keyPath)
	s.openErr = err
	s.store = nil
	if err == nil {
		s.store = opened
	}

	return nil
}

func (s *Suite) removeKey() error {
	return os.Remove(s.keyPath)
}

func (s *Suite) permissiveFile(file string) error {
	path := s.secretPath
	if file == "Masterschlüssel" {
		path = s.keyPath
	}

	return os.Chmod(path, 0644)
}

func (s *Suite) encryptedOnDisk() error {
	data, err := os.ReadFile(s.secretPath)
	if err != nil {
		return err
	}

	if bytes.Contains(data, []byte(syntheticCredential)) {
		return fmt.Errorf("Anmeldedaten stehen im Klartext in der Secret-Datei")
	}

	return nil
}

func (s *Suite) credentialSurvived() error {
	if s.openErr != nil {
		return fmt.Errorf("geschützter Speicher ließ sich nicht erneut öffnen")
	}

	actual, err := s.store.Load(context.Background(), "codex-subscription")
	if err != nil || !bytes.Equal(actual, []byte(syntheticCredential)) {
		return fmt.Errorf("gespeicherte Anmeldedaten fehlen nach erneutem Öffnen")
	}

	return nil
}

func (s *Suite) storeRejected() error {
	if s.openErr == nil || s.store != nil {
		return fmt.Errorf("unsicherer Verbindungsspeicher wurde geöffnet")
	}

	return nil
}
