package openrouterverbindung

import "net/http"

func (s *Suite) noConnection() error {
	if err := s.getSettings(); err != nil {
		return err
	}

	return s.expectStatusText("nicht eingerichtet")
}

func (s *Suite) validConnection() error {
	if err := s.saveKey(validKey); err != nil {
		return err
	}

	return s.rememberReference()
}

func (s *Suite) saveValid() error {
	return s.saveKey(validKey)
}

func (s *Suite) saveReplacement() error {
	return s.saveKey(replacementKey)
}

func (s *Suite) saveEmpty() error {
	return s.saveKey("")
}

func (s *Suite) saveInvalid() error {
	return s.saveKey(invalidKey)
}

func (s *Suite) checkedReady() error {
	if err := s.checkConnection(); err != nil {
		return err
	}

	return s.readyStatus()
}

func (s *Suite) fetchBothViews() error {
	if err := s.getHTML(); err != nil {
		return err
	}

	s.htmlBody = append([]byte(nil), s.responseBody...)
	if err := s.getSettings(); err != nil {
		return err
	}

	s.jsonBody = append([]byte(nil), s.responseBody...)
	return nil
}

func (s *Suite) rejectedSave(origin string) error {
	if origin == "fremdem Origin" {
		s.originMode = "foreign"
	}

	if origin == "ohne Origin" {
		s.originMode = "missing"
	}

	return s.send(http.MethodPost, apiPath, "application/json", []byte(`{"key":"`+replacementKey+`"}`))
}

func (s *Suite) rejectedCheck() error {
	s.hostOverride = "rebind.example"
	return s.send(http.MethodPost, apiPath+"/pruefen", "application/json", nil)
}
