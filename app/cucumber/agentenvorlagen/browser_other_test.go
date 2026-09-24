package agentenvorlagen

import (
	"fmt"
	"strings"
)

func (s *Suite) browserOpenUnconnectedProfile(name string) error {
	if err := s.browserOpenApplication("Nordstern"); err != nil {
		return err
	}
	if err := s.agentExists(name, "Nordstern", "Recherche"); err != nil {
		return err
	}
	return s.browserNavigate("/organisationen/"+s.orgIDs["Nordstern"]+"/agenten/"+s.agentIDs[name], name)
}

func (s *Suite) browserStart(_ string) error {
	return s.browserCommand(map[string]any{"submit": `main form[action$="/start"]`, "wait": map[string]string{"text": "Modellverbindung fehlt"}})
}

func (s *Suite) browserStartReason() error {
	if !strings.Contains(s.page.Text, "Modellverbindung fehlt") {
		return fmt.Errorf("Startgrund fehlt: %s", s.page.Text)
	}
	return nil
}

func (s *Suite) browserNoStartConfirmation() error {
	if strings.Contains(s.page.URL, "/laeufe/") || strings.Contains(s.page.Text, "Lauf gestartet") {
		return fmt.Errorf("unerwartete Startbestätigung: %+v", s.page)
	}
	return nil
}

func (s *Suite) browserStillNotReady(_ string) error { return s.browserNotReady() }

func (s *Suite) browserSelectTeam(name string) error {
	s.selectedTeam = name
	return s.browserCommand(map[string]any{"click": `main a[href*="team=` + strings.ToLower(name) + `"]`, "wait": map[string]string{"text": "Teamvorlage " + name}})
}

func (s *Suite) browserTeamPreview(name string) error {
	if !strings.Contains(s.page.URL, "team="+strings.ToLower(s.selectedTeam)) {
		return fmt.Errorf("keine gewählte Teamvorlage: %s", s.page.URL)
	}
	template, err := s.loadTemplate(s.selectedOrg, name)
	if err != nil {
		return err
	}
	expected := name + " · " + template.Role + " · " + template.Instructions
	for _, entry := range s.page.TeamEntries {
		if strings.Contains(entry, expected) {
			return s.browserCommand(map[string]any{"click": `main a[href*="vorlage=` + strings.ToLower(name) + `"]`, "wait": map[string]string{"text": "Vorschau: " + name}})
		}
	}
	return fmt.Errorf("Rolle oder Auftrag fehlen in Teamvorschau: %+v", s.page.TeamEntries)
}

func (s *Suite) browserCreateFromTeam(name string) error { return s.browserCreate(name) }

func (s *Suite) browserListOnly(name, org string) error {
	if err := s.browserListContains(name, org); err != nil {
		return err
	}
	if len(s.page.Cards) != 1 {
		return fmt.Errorf("Teamimport statt Einzelagent: %+v", s.page.Cards)
	}
	return nil
}

func (s *Suite) browserNameError() error {
	if !strings.Contains(s.page.Alert, "Name") && !strings.Contains(s.page.Alert, "Namen") {
		return fmt.Errorf("Namenshinweis fehlt: %+v", s.page)
	}
	return nil
}

func (s *Suite) browserListEmpty(org string) error {
	if err := s.browserOpenList(org); err != nil {
		return err
	}
	return s.browserEmpty()
}

func (s *Suite) browserDuplicate(template, name string) error {
	if err := s.browserNavigate("/organisationen/"+s.orgIDs["Nordstern"]+"/agenten", "Agentenübersicht"); err != nil {
		return err
	}
	if err := s.browserOpenForm(); err != nil {
		return err
	}
	if err := s.browserSelectTemplate(template); err != nil {
		return err
	}
	if err := s.browserFillName(name); err != nil {
		return err
	}
	return s.browserSubmit()
}

func (s *Suite) browserNameConflict() error {
	if !strings.Contains(strings.ToLower(s.page.Alert), "vergeben") {
		return fmt.Errorf("Konflikthinweis fehlt: %+v", s.page)
	}
	return nil
}

func (s *Suite) browserListOnce(org, agent string) error {
	if err := s.browserOpenList(org); err != nil {
		return err
	}
	return s.browserCount(agent, 1)
}
