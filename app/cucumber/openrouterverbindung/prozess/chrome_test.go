package prozess

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
)

func (s *Suite) chromePage() error {
	input, err := json.Marshal(map[string]string{"url": "http://" + s.address + htmlPath, "key": syntheticKey})
	if err != nil {
		return err
	}
	command := exec.Command("node", "browser.mjs")
	command.Stdin = bytes.NewReader(input)
	output, err := command.CombinedOutput()
	if err != nil {
		return fmt.Errorf("echter Chrome-Prozess: %w: %s", err, output)
	}
	var result ChromeResult
	if err := json.Unmarshal(output, &result); err != nil {
		return fmt.Errorf("Chrome-Ausgabe: %w: %s", err, output)
	}
	if !result.OK {
		return fmt.Errorf("Chrome-Seitenprüfung: %s", result.Error)
	}
	return nil
}
