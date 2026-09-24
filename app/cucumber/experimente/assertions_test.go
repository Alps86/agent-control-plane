package experimente

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

func (s *Suite) sourceDate() error {
	if !strings.Contains(s.text(s.body), "Quellenstand: 24.09.2026") {
		return fmt.Errorf("freigegebener Quellenstand fehlt")
	}

	return nil
}

func (s *Suite) groups() error {
	if len(s.cards()) != 23 {
		return fmt.Errorf("%d statt 23 Quellzeilen sichtbar", len(s.cards()))
	}

	for group, count := range map[string]int{"experimentell": 14, "api": 9} {
		if err := s.group(group, count); err != nil {
			return err
		}
	}

	return nil
}

func (s *Suite) group(group string, expected int) error {
	marker := `data-group-key="` + group + `"`
	if strings.Count(s.body, marker) != 1 {
		return fmt.Errorf("Gruppe %s erscheint nicht genau einmal", group)
	}

	count := 0
	for key := range s.cards() {
		if strings.HasPrefix(key, group+"-") {
			count++
		}
	}

	if count != expected {
		return fmt.Errorf("Gruppe %s hat %d statt %d Zeilen", group, count, expected)
	}

	return nil
}

func (s *Suite) entries() error {
	if err := s.groups(); err != nil {
		return err
	}

	for _, row := range s.rows {
		if err := s.entry(row); err != nil {
			return err
		}
	}

	return nil
}

func (s *Suite) entry(row SourceRow) error {
	card, err := s.cardFor(row)
	if err != nil {
		return err
	}

	if err := s.fields(row, card); err != nil {
		return err
	}

	for _, link := range row.links {
		if !strings.Contains(card, `href="`+link+`"`) {
			return fmt.Errorf("%s: Primärquelle %s fehlt", row.name, link)
		}
	}

	return nil
}

func (s *Suite) fields(row SourceRow, card string) error {
	for _, label := range []string{"Paperclip-Reifegrad", "Nutzen", "ACP-Abhängigkeiten", "Entscheidung und Begründung", "Primärquellen"} {
		if !strings.Contains(card, "<dt>"+label+"</dt>") {
			return fmt.Errorf("%s: Feld %s fehlt", row.name, label)
		}
	}

	if !strings.Contains(s.text(card), "In ACP noch nicht nachgewiesen") {
		return fmt.Errorf("%s: ACP-Status fehlt", row.name)
	}

	return s.sourceCells(row, card)
}

func (s *Suite) sourceCells(row SourceRow, card string) error {
	visible := s.text(card)
	start := 1
	if row.group == "api" {
		start = 2
	}

	for _, cell := range row.cells[start:] {
		if cell != "" && !strings.Contains(visible, cell) {
			return fmt.Errorf("%s: freigegebener Inhalt fehlt: %q", row.name, cell)
		}
	}

	return nil
}

func (s *Suite) namedEntries() error {
	for _, name := range []string{"Plugin SDK und Manager", "Isolated Workspaces", "Cases API", "Chat-Style Tasks", "Environments", "External Objects", "Status Cards"} {
		if !strings.Contains(s.text(s.body), name) {
			return fmt.Errorf("benannter Inventareintrag %q fehlt", name)
		}
	}

	return nil
}

func (s *Suite) sameFragment() error {
	fragment := s.cards()
	if len(fragment) != 23 {
		return fmt.Errorf("Fragment hat %d statt 23 Zeilen", len(fragment))
	}

	if err := s.entries(); err != nil {
		return err
	}

	if err := s.openPage(); err != nil {
		return err
	}

	return s.compareCards(fragment)
}

func (s *Suite) compareCards(fragment map[string]string) error {
	if err := s.pageResponse(); err != nil {
		return err
	}

	for key, card := range s.cards() {
		if card != fragment[key] {
			return fmt.Errorf("Vollseite und Fragment unterscheiden sich bei %s", key)
		}
	}

	return s.groups()
}

func (s *Suite) decisions() error {
	for _, kind := range []string{"zuordnen", "spaeter-evaluieren", "abgeloest"} {
		if !strings.Contains(s.body, `data-kind="`+kind+`"`) {
			return fmt.Errorf("Entscheidung %s fehlt", kind)
		}
	}

	return nil
}

func (s *Suite) laterDecisions() error {
	for _, name := range []string{"Plugin SDK und Manager", "Environments", "Isolated Workspaces", "Cases API", "Status Cards"} {
		if err := s.later(name); err != nil {
			return err
		}
	}

	return nil
}

func (s *Suite) later(name string) error {
	for _, row := range s.rows {
		if row.name != name {
			continue
		}

		card, err := s.cardFor(row)
		if err != nil {
			return err
		}

		if !strings.Contains(card, `data-kind="spaeter-evaluieren"`) || !strings.Contains(s.text(card), "Später evaluieren") {
			return fmt.Errorf("%s ohne sichtbare spätere Evaluation", name)
		}

		return nil
	}

	return fmt.Errorf("freigegebene Quellzeile %s fehlt", name)
}

func (s *Suite) maturity() error {
	for _, value := range []string{"Alpha", "Experiment", "API-dokumentiert", "Zum Kern befördert"} {
		if !strings.Contains(s.text(s.body), value) {
			return fmt.Errorf("Paperclip-Reifegrad %q fehlt", value)
		}
	}

	return nil
}

func (s *Suite) statuses() error {
	for key, card := range s.cards() {
		if !strings.Contains(s.text(card), "ACP: In ACP noch nicht nachgewiesen") {
			return fmt.Errorf("%s ohne belegten ACP-Status", key)
		}
	}

	return s.groups()
}

func (s *Suite) noFalseClaim() error {
	if strings.Contains(s.text(s.body), "Bereits in ACP verfügbar") {
		return fmt.Errorf("unbelegte ACP-Verfügbarkeit behauptet")
	}

	return s.statuses()
}

func (s *Suite) safeLinks() error {
	allowed := map[string]bool{}
	for _, row := range s.rows {
		for _, link := range row.links {
			allowed[link] = true
		}
	}

	return s.checkLinks(allowed)
}

func (s *Suite) checkLinks(allowed map[string]bool) error {
	pattern := regexp.MustCompile(`<a\b[^>]*href="([^"]+)"[^>]*>`)
	for _, match := range pattern.FindAllStringSubmatch(s.body, -1) {
		if err := s.checkLink(match[1], allowed); err != nil {
			return err
		}
	}

	return nil
}

func (s *Suite) checkLink(raw string, allowed map[string]bool) error {
	if strings.HasPrefix(raw, "#") || strings.HasPrefix(raw, "/") {
		return nil
	}

	parsed, err := url.Parse(raw)
	if err != nil || parsed.Scheme != "https" || parsed.Host != "docs.paperclip.ing" || !allowed[raw] {
		return fmt.Errorf("nicht freigegebener externer Quellenlink %q", raw)
	}

	return nil
}

func (s *Suite) escapedText() error {
	for _, tag := range []string{"<script", "<iframe", "<object", "<embed"} {
		for _, card := range s.cards() {
			if strings.Contains(strings.ToLower(card), tag) {
				return fmt.Errorf("ausführbares HTML im Quellentext: %s", tag)
			}
		}
	}

	return s.entries()
}
