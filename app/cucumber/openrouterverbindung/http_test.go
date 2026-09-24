package openrouterverbindung

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func (s *Suite) send(method, path, contentType string, body []byte) error {
	request, err := http.NewRequest(method, s.app.URL+path, bytes.NewReader(body))
	if err != nil {
		return err
	}

	if contentType != "" {
		request.Header.Set("Content-Type", contentType)
	}

	s.setRequestOrigin(request)
	response, err := s.app.Client().Do(request)
	if err != nil {
		return err
	}

	return s.readResponse(response)
}

func (s *Suite) readResponse(response *http.Response) error {
	defer response.Body.Close()
	s.responseStatus = response.StatusCode
	body, err := io.ReadAll(io.LimitReader(response.Body, 1<<20))
	s.responseBody = body
	return err
}

func (s *Suite) setRequestOrigin(request *http.Request) {
	if request.Method == http.MethodGet {
		return
	}

	if s.originMode != "missing" {
		request.Header.Set("Origin", s.app.URL)
	}

	if s.originMode == "foreign" {
		request.Header.Set("Origin", "https://foreign.example")
	}

	if s.hostOverride != "" {
		request.Host = s.hostOverride
	}

	s.originMode, s.hostOverride = "", ""
}

func (s *Suite) expectStatus(expected int) error {
	if s.responseStatus != expected {
		return fmt.Errorf("HTTP %d statt %d: %s", s.responseStatus, expected, s.responseBody)
	}

	return nil
}

func (s *Suite) sendForm(path, key string) error {
	values := url.Values{}
	if key != "" {
		values.Set("key", key)
	}

	return s.send(http.MethodPost, path, "application/x-www-form-urlencoded", []byte(values.Encode()))
}
