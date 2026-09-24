package experimente

import (
	"encoding/json"
	"fmt"
	"os/exec"
)

func (s *Suite) browse(address string, width int) (BrowserPage, error) {
	output, err := exec.Command("node", "browser.mjs", address, fmt.Sprint(width)).CombinedOutput()
	if err != nil {
		return BrowserPage{}, fmt.Errorf("Chrome-Browserweg: %w: %s", err, output)
	}

	var page BrowserPage
	if err := json.Unmarshal(output, &page); err != nil {
		return BrowserPage{}, fmt.Errorf("Chrome-Ergebnis: %w: %s", err, output)
	}

	return page, nil
}
