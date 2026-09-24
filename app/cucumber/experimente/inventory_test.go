package experimente

import (
	"bufio"
	"fmt"
	"html"
	"os"
	"regexp"
	"strings"
)

func (i *Inventory) read() ([]SourceRow, error) {
	file, err := os.Open("../../../spec/experimente/inventar.md")
	if err != nil {
		return nil, err
	}

	defer file.Close()
	scanner := bufio.NewScanner(file)
	group := ""
	for scanner.Scan() {
		group = i.line(scanner.Text(), group)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return i.rows, i.validate()
}

func (i *Inventory) line(line, group string) string {
	if line == "## Experimentelle Bedienflächen und Laufzeitfunktionen" {
		return "experimentell"
	}

	if line == "## Weitere in der API-Navigation sichtbare Flächen" {
		return "api"
	}

	if strings.HasPrefix(line, "## ") {
		return ""
	}

	if group != "" && strings.HasPrefix(line, "| [") {
		i.rows = append(i.rows, i.row(line, group))
	}

	return group
}

func (i *Inventory) row(line, group string) SourceRow {
	parts := strings.Split(strings.Trim(line, "|"), "|")
	row := SourceRow{group: group}
	for _, part := range parts {
		row.cells = append(row.cells, i.plain(part))
		for _, match := range regexp.MustCompile(`\[[^]]+\]\((https://[^)]+)\)`).FindAllStringSubmatch(part, -1) {
			row.links = append(row.links, match[1])
		}
	}

	row.name = strings.TrimSpace(regexp.MustCompile(`\[([^]]+)\]`).FindStringSubmatch(parts[0])[1])
	return row
}

func (i *Inventory) plain(value string) string {
	value = regexp.MustCompile(`\[([^]]+)\]\(https://[^)]+\)`).ReplaceAllString(value, "$1")
	value = strings.ReplaceAll(strings.ReplaceAll(value, "**", ""), "`", "")
	return strings.Join(strings.Fields(value), " ")
}

func (i *Inventory) validate() error {
	counts := map[string]int{}
	for _, row := range i.rows {
		counts[row.group]++
		if row.name == "" || len(row.links) == 0 {
			return fmt.Errorf("unvollständige Quellzeile im freigegebenen Inventar")
		}
	}

	if counts["experimentell"] != 14 || counts["api"] != 9 {
		return fmt.Errorf("freigegebenes Inventar hat %d/%d statt 14/9 Zeilen", counts["experimentell"], counts["api"])
	}

	return nil
}

func (s *Suite) cards() map[string]string {
	cards := map[string]string{}
	pattern := regexp.MustCompile(`(?s)<article\b[^>]*data-source-key="([^"]+)"[^>]*>.*?</article>`)
	for _, match := range pattern.FindAllStringSubmatch(s.body, -1) {
		cards[match[1]] = match[0]
	}

	return cards
}

func (s *Suite) text(value string) string {
	value = regexp.MustCompile(`(?s)<[^>]*>`).ReplaceAllString(value, " ")
	return strings.Join(strings.Fields(html.UnescapeString(value)), " ")
}

func (s *Suite) cardFor(row SourceRow) (string, error) {
	for key, card := range s.cards() {
		if !strings.HasPrefix(key, row.group+"-") || !strings.Contains(card, `href="`+row.links[0]+`"`) {
			continue
		}

		if !strings.Contains(s.text(card), row.name) {
			return "", fmt.Errorf("Quellzeile %s ohne Namen %q", key, row.name)
		}

		return card, nil
	}

	return "", fmt.Errorf("Quellzeile für %s fehlt", row.name)
}
