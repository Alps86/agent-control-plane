package uipreview

import (
	"fmt"
	"net/http"
	"strings"
)

func (s *Suite) unknown() error {
	s.responses = nil
	for _, path := range []string{"/?view=missing", "/?view=projects&fixture=missing"} {
		result, err := s.fetch(path, false)
		if err != nil {
			return err
		}

		s.responses = append(s.responses, result)
	}

	return nil
}

func (s *Suite) unknownDenied() error {
	return s.allDenied()
}

func (s *Suite) otherPath() error {
	s.responses = nil
	for _, path := range []string{"/fixtures/organization.json", "/go.mod", "/private", "/fragments/missing.html"} {
		result, err := s.fetch(path, false)
		if err != nil {
			return err
		}

		s.responses = append(s.responses, result)
	}

	return nil
}

func (s *Suite) otherPathDenied() error {
	return s.allDenied()
}

func (s *Suite) traversal() error {
	s.responses = nil
	for _, path := range []string{"/?view=projects&fixture=..%2Forganization", "/assets/%2e%2e%2ffiiixtures%2forganization.json"} {
		result, err := s.fetch(path, false)
		if err != nil {
			return err
		}

		s.responses = append(s.responses, result)
	}

	return nil
}

func (s *Suite) traversalDenied() error {
	return s.allDenied()
}

func (s *Suite) allDenied() error {
	if len(s.responses) == 0 {
		return fmt.Errorf("no requests checked")
	}

	for _, result := range s.responses {
		if err := s.denied(result); err != nil {
			return err
		}
	}

	return nil
}

func (s *Suite) denied(result Response) error {
	if result.Status != http.StatusNotFound {
		return fmt.Errorf("expected 404, got %d", result.Status)
	}

	for _, leaked := range []string{"Atelier Nord", "Kundenportal", "module agentcontrolplane", "PageTitle"} {
		if strings.Contains(result.Body, leaked) {
			return fmt.Errorf("response leaks %q", leaked)
		}
	}

	return nil
}
