package modellverbindung

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	appflow "agentcontrolplane/app/internal/app/modellverbindung"
	credentialport "agentcontrolplane/app/internal/port/credentials"
	"github.com/cucumber/godog"
)

func (s *Suite) registerTokenSteps(sc *godog.ScenarioContext) {
	sc.Step("^nur der serverseitige Zugangs-Port kann Token und Konto-ID zusammen auflösen$", s.onlyServerAccess)
	sc.Step("^das Tokenbündel liegt verschlüsselt im Verbindungsspeicher$", s.tokenBundleEncrypted)
	sc.Step("^das Zugangstoken kurz vor Ablauf steht und der Anbieter neue Tokens liefert$", s.tokenExpired)
	sc.Step("^der serverseitige Zugangs-Port Token und Konto-ID anfordert$", s.requestAccess)
	sc.Step("^erhält nur der serverseitige Port das erneuerte Token mit derselben Konto-ID$", s.rotatedAccess)
	sc.Step("^der Anbieter die Token-Erneuerung mit \"invalid_grant\" verweigert$", s.invalidGrant)
	sc.Step("^enthalten Statusantwort Protokoll und UI-Fixtures keine Zugangstokens$", s.noTokenArtifacts)
	sc.Step("^ich erhalte keinen neuen Gerätecode vom Anbieter$", s.noSecondChallenge)
	sc.Step("^der serverseitige Zugangs-Port liefert weiterhin Token und Konto-ID$", s.accessStillAvailable)
	sc.Step("^der Anbieter bei der Erneuerung kein frisches Zugangstoken liefert$", s.missingFreshToken)
	sc.Step("^der Anbieter bei der Erneuerung ein Zugangstoken für ein anderes Konto liefert$", s.foreignAccountToken)
	sc.Step("^der Anbieter ein neues Zugangstoken für dasselbe Konto ohne ID-Token liefert$", s.matchingTokenNoID)
	sc.Step("^liefert der Port weder Token noch Konto-ID$", s.noAccessPair)
	sc.Step("^das bisherige Tokenbündel wird nicht durch ein neues ersetzt$", s.bundleNotReplaced)
	sc.Step("^das erneuerte Tokenbündel wird geschützt gespeichert$", s.renewedBundleStored)
}

func (s *Suite) requestAccess() error {
	var resolver credentialport.AccessResolver = s.service
	s.accessToken, s.accountID, s.accessErr = resolver.Access(context.Background())
	return nil
}

func (s *Suite) onlyServerAccess() error {
	if err := s.requestAccess(); err != nil {
		return err
	}
	if s.accessErr != nil || s.accessToken == "" || s.accountID != mockAccountID {
		return fmt.Errorf("serverseitige Token-Konto-Auflösung fehlgeschlagen")
	}
	return s.noSecrets()
}

func (s *Suite) tokenBundleEncrypted() error {
	saved, err := s.store.Load(context.Background(), appflow.CredentialKey)
	if err != nil || !bytes.Contains(saved, []byte(s.accessToken)) {
		return fmt.Errorf("Tokenbündel wurde nicht gespeichert")
	}
	encrypted, err := os.ReadFile(s.secretPath)
	if err != nil {
		return err
	}
	for _, secret := range []string{s.issuer.issuedToken, s.issuer.rotatedToken, mockRefreshToken, mockRotatedRefreshToken, mockAccountID} {
		if secret != "" && bytes.Contains(encrypted, []byte(secret)) {
			return fmt.Errorf("Tokenbündel steht im Klartext auf Disk")
		}
	}
	return nil
}

func (s *Suite) tokenExpired() error {
	if s.issuer.issuedToken == "" {
		return fmt.Errorf("kein ausgegebenes Token vorhanden")
	}
	return nil
}

func (s *Suite) rotatedAccess() error {
	if s.accessErr != nil || s.issuer.refreshes == 0 || s.accessToken != s.issuer.rotatedToken || s.accountID != mockAccountID {
		return fmt.Errorf("rotierter Zugang oder Konto-ID fehlt")
	}
	return s.noSecrets()
}

func (s *Suite) invalidGrant() error {
	s.issuer.refreshOutcome = "invalid_grant"
	return nil
}

func (s *Suite) noSecondChallenge() error {
	if err := s.expectState("connected", ""); err != nil {
		return err
	}

	if s.response.data.UserCode != "" || s.issuer.starts != 1 {
		return fmt.Errorf("verbundene Sitzung löste neue Gerätecode-Anmeldung aus")
	}

	return nil
}

func (s *Suite) accessStillAvailable() error {
	if err := s.requestAccess(); err != nil {
		return err
	}

	if s.accessErr != nil || s.accessToken == "" || s.accountID != mockAccountID {
		return fmt.Errorf("verbundene Sitzung verlor ihren Modellzugang")
	}

	return nil
}

func (s *Suite) missingFreshToken() error {
	s.issuer.refreshOutcome = "missing_access"
	return s.rememberBundle()
}

func (s *Suite) foreignAccountToken() error {
	s.issuer.refreshOutcome = "foreign_account"
	return s.rememberBundle()
}

func (s *Suite) matchingTokenNoID() error {
	s.issuer.refreshOutcome = "matching_no_id"
	return nil
}

func (s *Suite) rememberBundle() error {
	var err error
	s.priorBundle, err = s.store.Load(context.Background(), appflow.CredentialKey)
	return err
}

func (s *Suite) noAccessPair() error {
	if s.accessErr == nil || s.accessToken != "" || s.accountID != "" {
		return fmt.Errorf("fehlgeschlagene Erneuerung lieferte Token oder Konto-ID")
	}

	return nil
}

func (s *Suite) bundleNotReplaced() error {
	current, err := s.store.Load(context.Background(), appflow.CredentialKey)
	if err != nil && !errors.Is(err, credentialport.ErrNotFound) {
		return err
	}

	if err == nil && !bytes.Equal(current, s.priorBundle) {
		return fmt.Errorf("fehlgeschlagene Erneuerung ersetzte Tokenbündel")
	}

	return nil
}

func (s *Suite) renewedBundleStored() error {
	return s.tokenBundleEncrypted()
}

func (s *Suite) noTokenArtifacts() error {
	if err := s.noSecrets(); err != nil {
		return err
	}
	if err := s.noTokenInLogs(); err != nil {
		return err
	}
	return s.noTokenInFixtures()
}

func (s *Suite) noTokenInLogs() error {
	for _, secret := range []string{s.issuer.issuedToken, s.issuer.rotatedToken, mockRefreshToken, mockRotatedRefreshToken} {
		if secret != "" && strings.Contains(s.logs.String(), secret) {
			return fmt.Errorf("Token in Protokoll")
		}
	}
	return nil
}

func (s *Suite) noTokenInFixtures() error {
	root := "../../../ui/web/fixtures"
	err := filepath.WalkDir(root, s.checkFixture)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

func (s *Suite) checkFixture(path string, entry fs.DirEntry, err error) error {
	if err != nil {
		return err
	}
	if entry.IsDir() {
		return nil
	}
	data, readErr := os.ReadFile(path)
	if readErr != nil {
		return readErr
	}
	for _, secret := range []string{s.issuer.issuedToken, s.issuer.rotatedToken, mockRefreshToken, mockRotatedRefreshToken} {
		if secret != "" && bytes.Contains(data, []byte(secret)) {
			return fmt.Errorf("Token in UI-Fixture")
		}
	}
	return nil
}
