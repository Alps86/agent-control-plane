package prozess

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
)

func (s *Suite) openBrowser() error       { return s.browser("read") }
func (s *Suite) browserDisconnect() error { return s.browser("disconnect") }

func (s *Suite) browser(action string) error {
	input, _ := json.Marshal(map[string]string{"base": "http://" + s.address, "action": action})
	command := exec.Command("node", "browser.mjs")
	command.Stdin = bytes.NewReader(input)
	output, err := command.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Chrome-Prozess: %w: %s", err, output)
	}
	if err := json.Unmarshal(output, &s.page); err != nil {
		return fmt.Errorf("Chrome-Ausgabe: %w: %s", err, output)
	}
	if !s.page.OK {
		return fmt.Errorf("Chrome-Seite: %s", s.page.Error)
	}
	return nil
}
