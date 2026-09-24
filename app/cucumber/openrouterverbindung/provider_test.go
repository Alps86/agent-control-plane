package openrouterverbindung

import (
	"net/http"
	"net/http/httptest"
)

func (s *Suite) startProvider() {
	s.provider = httptest.NewServer(http.HandlerFunc(s.mockProvider))
}

func (s *Suite) mockProvider(w http.ResponseWriter, r *http.Request) {
	authorization := r.Header.Get("Authorization")
	s.providerCalls = append(s.providerCalls, authorization)
	if s.probeEntered != nil && authorization == "Bearer "+validKey {
		s.probeOnce.Do(s.holdProvider)
	}

	if r.Method != http.MethodGet || r.URL.Path != "/api/v1/key" {
		http.NotFound(w, r)
		return
	}

	if r.Header.Get("Authorization") == "Bearer invalid-test-key" {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"data":{"label":"controlled-test-provider"}}`))
}

func (s *Suite) holdProvider() {
	s.probeEntered <- "Bearer " + validKey
	<-s.probeRelease
}
