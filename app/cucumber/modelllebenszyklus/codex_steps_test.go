package modelllebenszyklus

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"agentcontrolplane/app/internal/app/modellverbindung"
	"agentcontrolplane/app/internal/port/credentials"
	"github.com/cucumber/godog"
)

func (s *Suite) registerCodex(sc *godog.ScenarioContext) {
	sc.Step(`^eine gültige Codex-Abo-Sitzung liegt im geschützten Verbindungsspeicher$`, s.validCodex)
	sc.Step(`^ich den lokalen Serverprozess beende und mit demselben Speicher neu starte$`, s.restartProcess)
	sc.Step(`^ich die öffentliche Codex-Verbindungsansicht abrufe$`, s.publicCodex)
	sc.Step(`^zeigt sie die bestehende Verbindung ohne neuen Gerätecode als verbunden$`, s.connectedCodex)
	sc.Step(`^die serverseitige Zugangsgrenze den Zugang auflöst$`, s.resolveCodex)
	sc.Step(`^erhält sie das gespeicherte Token mit der zugehörigen Konto-ID$`, s.storedCodex)
	sc.Step(`^eine erneuerbare Codex-Abo-Sitzung läuft in Kürze ab$`, s.expiringCodex)
	sc.Step(`^der kontrollierte Codex-Anbieter hält die Erneuerungsantwort zurück$`, s.holdRefresh)
	sc.Step(`^zwei serverseitige Zugangsanforderungen gleichzeitig eintreffen$`, s.parallelAccess)
	sc.Step(`^ich die Erneuerungsantwort freigebe$`, s.releaseRefresh)
	sc.Step(`^erhält der Anbieter genau eine Erneuerungsanfrage$`, s.singleRefresh)
	sc.Step(`^beide Anforderungen erhalten dasselbe neue Token mit derselben Konto-ID$`, s.sameNewAccess)
	sc.Step(`^das neue Token ist nach einem Neustart weiterverwendbar$`, s.persistedNewAccess)
	sc.Step(`^der kontrollierte Codex-Anbieter verweigert die Erneuerung endgültig$`, s.revokeCodex)
	sc.Step(`^wird kein Token und keine Konto-ID ausgegeben$`, s.noAccess)
	sc.Step(`^die öffentliche Codex-Verbindungsansicht verlangt eine erneute Anmeldung$`, s.reauthentication)
	sc.Step(`^bleibt der Zugang bis zu einer neuen Gerätecode-Anmeldung gesperrt$`, s.stillRevoked)
}

func (s *Suite) validCodex() error {
	if err := s.setup(); err != nil {
		return err
	}
	return s.seedCodex(time.Hour)
}

func (s *Suite) expiringCodex() error {
	if err := s.setup(); err != nil {
		return err
	}
	return s.seedCodex(time.Minute)
}

func (s *Suite) publicCodex() error {
	if s.process != nil {
		if err := s.processCodexView(); err != nil {
			return err
		}
		return s.reopenFlow()
	}
	_, err := s.codexView()
	return err
}

func (s *Suite) connectedCodex() error {
	var view CodexView
	if err := s.decodeCodex(&view); err != nil {
		return err
	}
	if view.State != "connected" {
		return fmt.Errorf("Codex-Status %q", view.State)
	}
	return s.noSecrets()
}

func (s *Suite) decodeCodex(view *CodexView) error {
	return json.Unmarshal(s.response, view)
}

func (s *Suite) resolveCodex() error {
	var resolver credentials.AccessResolver = s.flow
	token, account, err := resolver.Access(context.Background())
	s.access[0] = AccessResult{token: token, account: account, err: err}
	return nil
}

func (s *Suite) storedCodex() error {
	result := s.access[0]
	if result.err != nil || result.token != priorToken || result.account != accountID {
		return fmt.Errorf("gespeicherter Codex-Zugang fehlt: %v", result.err)
	}
	return nil
}

func (s *Suite) holdRefresh() error {
	s.issuerState.entered = make(chan struct{}, 1)
	s.issuerState.release = make(chan struct{})
	return nil
}

func (s *Suite) parallelAccess() error {
	var wait sync.WaitGroup
	s.accessDone = make(chan struct{})
	wait.Add(2)
	for i := range s.access {
		go s.accessOne(i, &wait)
	}
	go func() { wait.Wait(); close(s.accessDone) }()
	select {
	case <-s.issuerState.entered:
		return nil
	case <-time.After(3 * time.Second):
		return fmt.Errorf("Erneuerungsanfrage blieb aus")
	}
}

func (s *Suite) accessOne(index int, wait *sync.WaitGroup) {
	defer wait.Done()
	token, account, err := s.flow.Access(context.Background())
	s.access[index] = AccessResult{token: token, account: account, err: err}
}

func (s *Suite) releaseRefresh() error {
	close(s.issuerState.release)
	select {
	case <-s.accessDone:
		return nil
	case <-time.After(3 * time.Second):
		return fmt.Errorf("parallele Auflösungen blieben aus")
	}
}

func (s *Suite) singleRefresh() error {
	if count := s.issuerState.count(); count != 1 {
		return fmt.Errorf("%d statt einer Erneuerungsanfrage", count)
	}
	return nil
}

func (s *Suite) sameNewAccess() error {
	for _, result := range s.access {
		if result.err != nil || result.token == "" || result.token == priorToken || result.account != accountID {
			return fmt.Errorf("uneinheitlicher erneuerter Zugang: %v", result.err)
		}
	}
	if s.access[0].token != s.access[1].token {
		return fmt.Errorf("verschiedene Tokens nach parallelem Refresh")
	}
	return nil
}

func (s *Suite) persistedNewAccess() error {
	if err := s.reopenFlow(); err != nil {
		return err
	}
	if err := s.resolveCodex(); err != nil {
		return err
	}
	if s.access[0].err != nil || s.access[0].token != s.access[1].token || s.issuerState.count() != 1 {
		return fmt.Errorf("erneuerter Zugang nach Neustart nicht wiederverwendbar")
	}
	return nil
}

func (s *Suite) revokeCodex() error { s.issuerState.revoked = true; return nil }

func (s *Suite) noAccess() error {
	if !errors.Is(s.access[0].err, modellverbindung.ErrReauthenticationRequired) || s.access[0].token != "" || s.access[0].account != "" {
		return fmt.Errorf("widerrufene Sitzung lieferte Zugang oder falschen Fehler")
	}
	return nil
}

func (s *Suite) reauthentication() error {
	view, err := s.codexView()
	if err != nil {
		return err
	}
	if view.State != "reauthentication_required" {
		return fmt.Errorf("Codex-Status %q", view.State)
	}
	return s.noSecrets()
}

func (s *Suite) stillRevoked() error {
	if err := s.reopenFlow(); err != nil {
		return err
	}
	if err := s.resolveCodex(); err != nil {
		return err
	}
	if s.access[0].err == nil || s.access[0].token != "" {
		return fmt.Errorf("widerrufener Zugang nach Neustart verfügbar")
	}
	return nil
}
