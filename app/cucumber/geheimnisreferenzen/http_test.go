package geheimnisreferenzen

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func (s *Suite) request(method, path string, body any) error {
	encoded, err := s.encode(body)
	if err != nil {
		return err
	}
	request, err := http.NewRequest(method, s.app.URL+path, bytes.NewReader(encoded))
	if err != nil {
		return err
	}
	s.requestHeaders(request, method, body)
	response, err := s.app.Client().Do(request)
	if err != nil {
		return err
	}
	return s.record(response)
}

func (s *Suite) requestHeaders(request *http.Request, method string, body any) {
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	if method != http.MethodGet {
		request.Header.Set("Origin", s.app.URL)
	}
}

func (s *Suite) encode(body any) ([]byte, error) {
	if body == nil {
		return nil, nil
	}
	return json.Marshal(body)
}

func (s *Suite) record(response *http.Response) error {
	defer response.Body.Close()
	s.status = response.StatusCode
	last, err := io.ReadAll(io.LimitReader(response.Body, 1<<20))
	s.last = last
	if err == nil {
		s.snapshots = append(s.snapshots, append([]byte(nil), s.last...))
	}
	return err
}

func (s *Suite) statusView() (StatusView, error) {
	var view StatusView
	if s.status != http.StatusOK {
		return view, fmt.Errorf("Status HTTP %d: %s", s.status, s.last)
	}
	if err := json.Unmarshal(s.last, &view); err != nil {
		return view, err
	}
	return view, nil
}

func (v StatusView) provider(id string) (ProviderView, error) {
	for _, candidate := range v.Providers {
		if candidate.ID == id {
			return candidate, nil
		}
	}
	return ProviderView{}, fmt.Errorf("Anbieter %s fehlt", id)
}

func (s *Suite) getStatus() error {
	if len(s.before) == 0 && len(s.last) > 0 && strings.Contains(string(s.last), `"providers"`) {
		s.before = append([]byte(nil), s.last...)
	}
	return s.request(http.MethodGet, "/api/settings/modellanbieter/status", nil)
}

func (s *Suite) controlled() error {
	if s.issuer == nil || s.provider == nil {
		return fmt.Errorf("kontrollierte Anbieter fehlen")
	}
	return nil
}
