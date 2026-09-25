package modellwahl

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"time"
)

func (s *Suite) startProvider() error {
	s.stopProvider()
	s.providerCalls.Store(0)
	s.blockedExternal.Store(0)
	s.providerServer = httptest.NewServer(http.HandlerFunc(s.serveProvider))
	if err := s.requireLoopbackProvider(); err != nil {
		s.stopProvider()
		return err
	}
	if err := s.verifyProxyBarrier(); err != nil {
		s.stopProvider()
		return err
	}
	return nil
}

func (s *Suite) verifyProxyBarrier() error {
	proxyURL, err := url.Parse(s.providerServer.URL)
	if err != nil {
		return err
	}
	client := &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(proxyURL)}, Timeout: 2 * time.Second}
	request, err := http.NewRequest(http.MethodGet, "http://fremd.example/api/v1/key", nil)
	if err != nil {
		return err
	}
	request.Header.Set("Authorization", "Bearer "+s.secret)
	response, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("lokale Proxy-Sperrprüfung: %w", err)
	}
	defer response.Body.Close()
	return s.checkProxyResponse(response)
}

func (s *Suite) checkProxyResponse(response *http.Response) error {
	if response.StatusCode != http.StatusBadGateway || s.providerCalls.Load() != 0 || s.blockedExternal.Load() != 1 {
		return fmt.Errorf("externe HTTP-Proxy-Anfrage nicht gesperrt: HTTP %d, Probe %d, Sperren %d", response.StatusCode, s.providerCalls.Load(), s.blockedExternal.Load())
	}
	s.blockedExternal.Store(0)
	return nil
}

func (s *Suite) requireLoopbackProvider() error {
	if s.providerServer == nil {
		return fmt.Errorf("kontrollierter Anbieter fehlt")
	}
	address, err := url.Parse(s.providerServer.URL)
	if err != nil || address.Scheme != "http" || address.User != nil || address.RawQuery != "" || address.Fragment != "" {
		return fmt.Errorf("Probe-URL ist nicht lokal")
	}
	host, _, err := net.SplitHostPort(address.Host)
	if err != nil {
		return err
	}
	ip := net.ParseIP(host)
	if ip == nil || !ip.IsLoopback() {
		return fmt.Errorf("Probe-URL ist kein numerischer Loopback")
	}
	return nil
}

func (s *Suite) stopProvider() {
	if s.providerServer == nil {
		return
	}
	s.providerServer.Close()
	s.providerServer = nil
}

func (s *Suite) serveProvider(w http.ResponseWriter, r *http.Request) {
	if s.externalProviderRequest(r) {
		s.blockedExternal.Add(1)
		http.Error(w, "external provider blocked", http.StatusBadGateway)
		return
	}
	if r.Method != http.MethodGet || r.URL.Path != "/api/v1/key" || r.URL.RawQuery != "" {
		http.NotFound(w, r)
		return
	}
	s.providerCalls.Add(1)
	if r.Header.Get("Authorization") != "Bearer "+s.secret {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = io.WriteString(w, `{"data":{"label":"story25-controlled-provider"}}`)
}

func (s *Suite) externalProviderRequest(r *http.Request) bool {
	if s.providerServer == nil {
		return true
	}
	if r.Method == http.MethodConnect || r.URL.IsAbs() || r.URL.Host != "" {
		return true
	}
	return r.Host != s.providerServer.Listener.Addr().String()
}
