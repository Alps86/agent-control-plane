package modellwahl

import (
	"context"
	"fmt"
	"github.com/cucumber/godog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func (s *Suite) registerInvalidCatalog(sc *godog.ScenarioContext) {
	sc.Step(`^der kontrollierte Katalog enthält OpenRouter ohne Verbindungsreferenz$`, s.catalogWithoutConnection)
	sc.Step(`^ich einen Serverprozess mit diesem Katalog starte$`, s.startInvalidCatalog)
	sc.Step(`^endet der Start mit einem verständlichen Hinweis auf die fehlende OpenRouter-Verbindung$`, s.invalidCatalogReason)
	sc.Step(`^es entsteht keine öffentliche Modellwahl-Grenze aus dieser Konfiguration$`, s.invalidCatalogNoServer)
}

func (s *Suite) catalogWithoutConnection() error {
	declaration := `,"connections":[{"reference":"openrouter-central","label":"OpenRouter"}]`
	if !strings.Contains(catalogFixture, declaration) {
		return fmt.Errorf("Testkatalog enthält keine explizite OpenRouter-Verbindung")
	}
	s.invalidCatalogPath = filepath.Join(s.t.TempDir(), "catalog-without-openrouter-connection.json")
	invalid := strings.Replace(catalogFixture, declaration, "", 1)
	return os.WriteFile(s.invalidCatalogPath, []byte(invalid), 0600)
}

func (s *Suite) startInvalidCatalog() error {
	s.stopServer()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, s.binary)
	cmd.Env = append(s.serverEnv(), "APP_ADDR="+s.address, "APP_DB_PATH="+s.dbPath, "APP_MODEL_CATALOG_PATH="+s.invalidCatalogPath,
		"APP_CREDENTIALS_PATH="+filepath.Join(s.credentialDir, "credentials.enc"), "APP_CREDENTIAL_KEY_PATH="+filepath.Join(s.credentialDir, "master.key"),
		"APP_OPENROUTER_PROBE_URL="+s.providerServer.URL+"/api/v1/key", "HTTPS_PROXY="+s.providerServer.URL, "HTTP_PROXY="+s.providerServer.URL,
		"NO_PROXY=localhost,127.0.0.1")
	output, err := cmd.CombinedOutput()
	if ctx.Err() != nil {
		return fmt.Errorf("ungültiger Katalog startete Server statt abzubrechen: %s", output)
	}
	s.invalidOutput, s.invalidExited = output, err != nil
	return nil
}

func (s *Suite) invalidCatalogReason() error {
	if !s.invalidExited {
		return fmt.Errorf("Serverstart trotz fehlender OpenRouter-Verbindung erfolgreich")
	}
	if !strings.Contains(string(s.invalidOutput), "OpenRouter-Katalog: explizite Verbindungsreferenz openrouter-central fehlt oder ist ungültig") {
		return fmt.Errorf("unklarer Startfehler: %s", s.invalidOutput)
	}
	return nil
}

func (s *Suite) invalidCatalogNoServer() error {
	response, err := s.client.Get(s.url(s.choicePath()))
	if response != nil {
		response.Body.Close()
	}
	if err == nil {
		return fmt.Errorf("Modellwahl-Server trotz ungültigem Katalog erreichbar")
	}
	return nil
}
