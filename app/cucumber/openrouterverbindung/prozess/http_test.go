package prozess

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

func (s *Suite) get(path string) error {
	return s.send(http.MethodGet, path, "", "")
}

func (s *Suite) postForm(path string, values url.Values) error {
	return s.send(http.MethodPost, path, values.Encode(), "application/x-www-form-urlencoded")
}

func (s *Suite) delete(path string) error {
	return s.send(http.MethodDelete, path, "", "")
}

func (s *Suite) send(method, path, body, contentType string) error {
	request, err := http.NewRequest(method, "http://"+s.address+path, strings.NewReader(body))
	if err != nil {
		return err
	}

	request.Header.Set("Content-Type", contentType)
	if method != http.MethodGet {
		request.Header.Set("Origin", "http://"+s.address)
	}

	s.requested = append(s.requested, method+" "+path)
	return s.perform(request)
}

func (s *Suite) perform(request *http.Request) error {
	response, err := s.client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	s.status = response.StatusCode
	s.contentType = response.Header.Get("Content-Type")
	s.body, err = io.ReadAll(response.Body)
	return err
}

func (s *Suite) expectStatus(want int) error {
	if s.status != want {
		return fmt.Errorf("HTTP %d statt %d: %s", s.status, want, s.body)
	}

	return nil
}

func (s *Suite) connection() (PublicConnection, error) {
	if err := s.get(apiPath); err != nil {
		return PublicConnection{}, err
	}
	if err := s.expectStatus(http.StatusOK); err != nil {
		return PublicConnection{}, err
	}

	if bytes.Contains(s.body, []byte(syntheticKey)) {
		return PublicConnection{}, fmt.Errorf("Schlüssel in öffentlicher JSON-Antwort")
	}

	var result PublicConnection
	err := json.Unmarshal(s.body, &result)
	return result, err
}
