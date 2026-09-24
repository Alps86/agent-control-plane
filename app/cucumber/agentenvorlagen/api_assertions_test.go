package agentenvorlagen

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"
)

func (s *Suite) decodeProfile() (Profile, error) {
	var profile Profile
	err := json.Unmarshal(s.response.Body, &profile)
	return profile, err
}

func (s *Suite) getProfile(org, agent string) (Profile, error) {
	if err := s.request("GET", s.agentPath(org, agent), "", ""); err != nil {
		return Profile{}, err
	}
	if err := s.statusCode(http.StatusOK); err != nil {
		return Profile{}, err
	}
	return s.decodeProfile()
}

func (s *Suite) getAgentList(org string) (AgentList, error) {
	if err := s.request("GET", s.orgPath(org), "", ""); err != nil {
		return AgentList{}, err
	}
	if err := s.statusCode(http.StatusOK); err != nil {
		return AgentList{}, err
	}
	var list AgentList
	err := json.Unmarshal(s.response.Body, &list)
	return list, err
}

func (s *Suite) agentBelongsToOrg(agent, org string) error {
	profile, err := s.decodeProfile()
	if err != nil {
		return err
	}
	if profile.Name != agent || profile.OrganizationID != s.orgIDs[org] || profile.ExecutionKind != "eino" {
		return fmt.Errorf("Agentenkontext falsch: %+v", profile)
	}
	return nil
}

func (s *Suite) agentDetailProfile() error {
	profile, err := s.getProfile(s.orgName(s.initialProfile.OrganizationID), s.initialProfile.Name)
	if err != nil {
		return err
	}
	if profile.Role == "" || profile.Instructions == "" || len(profile.Capabilities) == 0 {
		return fmt.Errorf("unvollständiges Profil: %+v", profile)
	}
	return nil
}

func (s *Suite) agentNotReady() error {
	profile, err := s.getProfile(s.orgName(s.initialProfile.OrganizationID), s.initialProfile.Name)
	if err != nil {
		return err
	}
	if profile.Readiness.Ready || profile.Readiness.Code != "model_unconfigured" {
		return fmt.Errorf("Readiness falsch: %+v", profile.Readiness)
	}
	return nil
}

func (s *Suite) agentSameAfterRestart(name, org string) error {
	list, err := s.getAgentList(org)
	if err != nil {
		return err
	}
	count := 0
	for _, profile := range list.Agents {
		if profile.Name == name && profile.ID == s.initialID {
			count++
		}
	}
	if count != 1 {
		return fmt.Errorf("Agent nach Neustart: %d passende Einträge: %+v", count, list.Agents)
	}
	return nil
}

func (s *Suite) agentUnchanged() error {
	profile, err := s.getProfile(s.orgName(s.initialProfile.OrganizationID), s.initialProfile.Name)
	if err != nil {
		return err
	}
	if !reflect.DeepEqual(profile, s.initialProfile) {
		return fmt.Errorf("Profil nach Neustart verändert: vorher %+v, nachher %+v", s.initialProfile, profile)
	}
	return nil
}

func (s *Suite) getAgentFromOrg(agent, org string) error {
	return s.request("GET", s.orgPath(org)+"/"+s.agentIDs[agent], "", "")
}

func (s *Suite) notFoundWithoutData() error {
	if err := s.statusCode(http.StatusNotFound); err != nil {
		return err
	}
	body := string(s.response.Body)
	if strings.Contains(body, s.initialID) || strings.Contains(body, s.initialProfile.Name) || strings.Contains(body, s.initialProfile.Role) {
		return fmt.Errorf("404 verrät Agentendaten: %s", body)
	}
	return nil
}

func (s *Suite) agentAbsent(org, name string) error {
	list, err := s.getAgentList(org)
	if err != nil {
		return err
	}
	for _, profile := range list.Agents {
		if profile.Name == name || profile.ID == s.agentIDs[name] {
			return fmt.Errorf("fremder Agent sichtbar: %+v", profile)
		}
	}
	return nil
}

func (s *Suite) getSameAgent(org string) error {
	return s.request("GET", s.orgPath(org)+"/"+s.initialID, "", "")
}

func (s *Suite) sameAgentID(name string) error {
	if err := s.statusCode(http.StatusOK); err != nil {
		return err
	}
	profile, err := s.decodeProfile()
	if err != nil {
		return err
	}
	if profile.Name != name || profile.ID != s.initialID {
		return fmt.Errorf("andere Agentenkennung: %+v", profile)
	}
	return nil
}

func (s *Suite) noModelConnection(agent string) error {
	org := s.orgName(s.initialProfile.OrganizationID)
	if err := s.request("GET", s.agentPath(org, agent)+"/bereitschaft", "", ""); err != nil {
		return err
	}
	if err := s.statusCode(http.StatusOK); err != nil {
		return err
	}
	var readiness Readiness
	if err := json.Unmarshal(s.response.Body, &readiness); err != nil {
		return err
	}
	if readiness.Ready || readiness.Code != "model_unconfigured" {
		return fmt.Errorf("Modellstatus falsch: %+v", readiness)
	}
	return nil
}

func (s *Suite) requestStart(agent string) error {
	org := s.orgName(s.initialProfile.OrganizationID)
	return s.request("POST", s.agentPath(org, agent)+"/start", "", "application/json")
}

func (s *Suite) startDenied() error {
	if err := s.statusCode(http.StatusConflict); err != nil {
		return err
	}
	var failure struct {
		Error  string `json:"error"`
		Reason string `json:"reason"`
	}
	if err := json.Unmarshal(s.response.Body, &failure); err != nil {
		return err
	}
	if failure.Error != "model_unconfigured" || failure.Reason != "Modellverbindung fehlt" {
		return fmt.Errorf("Startfehler unklar: %s", s.response.Body)
	}
	return nil
}

func (s *Suite) noRunConfirmation() error {
	if s.response.Location != "" || strings.Contains(string(s.response.Body), `"run_id"`) || strings.Contains(string(s.response.Body), `"started"`) {
		return fmt.Errorf("Startantwort bestätigt einen Lauf: %+v", s.response)
	}
	return nil
}

func (s *Suite) stillNotReady(agent string) error { return s.noModelConnection(agent) }
