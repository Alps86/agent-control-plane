package kommentar

import (
	"regexp"
	"strings"
)

var artifactIDPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.:-]{0,127}$`)

// ValidatedContent entfernt nur äußeren Leerraum und verwirft leere Eingaben.
func (c Comment) ValidatedContent() (string, error) {
	content := strings.TrimSpace(c.Content)
	if content == "" {
		return "", ErrContentRequired
	}

	return content, nil
}

// Validate prüft Metadatenform; Aufgabenauflösung erfolgt im Store.
func (r *Reference) Validate() (*Reference, error) {
	if r == nil {
		return nil, nil
	}

	copy := *r
	copy.ID = strings.TrimSpace(copy.ID)
	copy.DisplayName = ""
	copy.Link = ""
	if copy.Type == ReferenceTask && artifactIDPattern.MatchString(copy.ID) {
		return &copy, nil
	}

	if copy.Type == ReferenceArtifact && artifactIDPattern.MatchString(copy.ID) {
		return &copy, nil
	}

	return nil, ErrInvalidReference
}
