package openrouterverbindung

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"time"
)

func (s *Suite) runBrowser() error {
	input, err := json.Marshal(map[string]string{
		"url": s.app.URL + htmlPath, "key": validKey, "replacement": replacementKey,
	})
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, "node", "browser.mjs")
	command.Stdin = bytes.NewReader(append(input, '\n'))
	output, err := command.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Browserlauf: %w: %s", err, output)
	}

	return s.browserResult(output)
}

func (s *Suite) browserResult(output []byte) error {
	var result struct {
		OK      bool     `json:"ok"`
		Error   string   `json:"error"`
		Origins []string `json:"origins"`
	}
	if json.Unmarshal(bytes.TrimSpace(output), &result) != nil || !result.OK {
		return fmt.Errorf("Browserpfad fehlgeschlagen: %s", output)
	}

	s.t.Logf("Chrome-Formular-Origin: %v", result.Origins)
	return nil
}
