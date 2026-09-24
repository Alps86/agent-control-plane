package modelllebenszyklus

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func (s *Suite) request(method, path, key string) error {
	body, err := json.Marshal(map[string]string{"key": key})
	if err != nil {
		return err
	}
	request, err := http.NewRequest(method, s.app.URL+path, bytes.NewReader(body))
	if err != nil {
		return err
	}
	if method != http.MethodGet {
		request.Header.Set("Origin", s.app.URL)
		request.Header.Set("Content-Type", "application/json")
	}

	response, err := s.app.Client().Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	s.status = response.StatusCode
	s.response, err = io.ReadAll(io.LimitReader(response.Body, 1<<20))
	return err
}

func (s *Suite) codexView() (CodexView, error) {
	if err := s.request(http.MethodGet, "/settings/modelle/codex/device/status", ""); err != nil {
		return CodexView{}, err
	}
	if s.status != http.StatusOK {
		return CodexView{}, fmt.Errorf("Codex HTTP %d: %s", s.status, s.response)
	}
	var view CodexView
	err := json.Unmarshal(s.response, &view)
	return view, err
}

func (s *Suite) routerView() (OpenRouterView, error) {
	if err := s.request(http.MethodGet, "/api/settings/modellanbieter/openrouter", ""); err != nil {
		return OpenRouterView{}, err
	}
	if s.status != http.StatusOK {
		return OpenRouterView{}, fmt.Errorf("OpenRouter HTTP %d: %s", s.status, s.response)
	}
	var view OpenRouterView
	err := json.Unmarshal(s.response, &view)
	return view, err
}

func (s *Suite) routerAction(method, path, key string) error {
	if err := s.request(method, path, key); err != nil {
		return err
	}
	if s.status != http.StatusOK {
		return fmt.Errorf("OpenRouter HTTP %d: %s", s.status, s.response)
	}
	return nil
}
