package experimente

import (
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/url"
	"strings"
)

//go:embed inventory.json
var embeddedInventory embed.FS

func (inventoryLoader) load() (inventory, error) {
	data, err := embeddedInventory.ReadFile("inventory.json")
	if err != nil {
		return inventory{}, err
	}

	var result inventory
	if err = json.Unmarshal(data, &result); err != nil {
		return inventory{}, err
	}

	return result, result.validate()
}

func (i inventory) validate() error {
	if i.SourceAsOf != "24.09.2026" || len(i.SourceSHA256) != sha256.Size*2 || len(i.Groups) != 2 {
		return errors.New("Quellenstand oder Gruppen fehlen")
	}

	if _, err := hex.DecodeString(i.SourceSHA256); err != nil {
		return err
	}

	if i.Groups[0].Key != "experimentell" || len(i.Groups[0].Entries) != 14 || i.Groups[1].Key != "api" || len(i.Groups[1].Entries) != 9 {
		return errors.New("Inventargruppen unvollständig")
	}

	return i.validateEntries()
}

func (i inventory) validateEntries() error {
	seen := make(map[string]bool)
	for _, group := range i.Groups {
		if err := group.validateEntries(seen); err != nil {
			return err
		}
	}

	return i.validateRelations(seen)
}

func (g group) validateEntries(seen map[string]bool) error {
	if strings.TrimSpace(g.Heading) == "" {
		return errors.New("Gruppentitel fehlt")
	}

	for _, entry := range g.Entries {
		if seen[entry.SourceKey] || !strings.HasPrefix(entry.SourceKey, g.Key+"-") {
			return errors.New("Quellkennung doppelt oder falsch")
		}

		seen[entry.SourceKey] = true
		if err := entry.validate(); err != nil {
			return err
		}
	}

	return nil
}

func (i inventory) validateRelations(seen map[string]bool) error {
	for _, group := range i.Groups {
		for _, entry := range group.Entries {
			if entry.RelatedSourceKey != "" && !seen[entry.RelatedSourceKey] {
				return errors.New("Querverweis ohne Quelle")
			}
		}
	}

	return nil
}

func (e entry) validate() error {
	if strings.TrimSpace(e.Name) == "" || strings.TrimSpace(e.PaperclipMaturity) == "" || strings.TrimSpace(e.Benefit) == "" || strings.TrimSpace(e.ACPDependencies) == "" || strings.TrimSpace(e.ACPImplementationStatus) == "" || len(e.SourceLinks) == 0 || len(e.DecisionParts) == 0 {
		return errors.New("Inventarzeile unvollständig")
	}

	for _, link := range e.SourceLinks {
		if err := link.validate(); err != nil {
			return err
		}
	}

	for _, part := range e.DecisionParts {
		if part.Kind == "" || strings.TrimSpace(part.Reason) == "" {
			return errors.New("Entscheidung unvollständig")
		}
	}

	return nil
}

func (l sourceLink) validate() error {
	u, err := url.Parse(l.URL)
	if err != nil || strings.TrimSpace(l.Label) == "" {
		return errors.New("Quellenlink ungültig")
	}

	if u.Scheme != "https" || u.Host != "docs.paperclip.ing" || u.User != nil || u.Fragment != "" {
		return errors.New("Quellenlink nicht freigegeben")
	}

	return nil
}
