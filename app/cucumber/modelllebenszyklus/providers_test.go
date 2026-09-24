package modelllebenszyklus

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

func (p *IssuerState) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/oauth/token" {
		http.NotFound(w, r)
		return
	}
	var request map[string]string
	if json.NewDecoder(r.Body).Decode(&request) != nil || request["refresh_token"] != priorRefresh {
		http.Error(w, `{"error":"invalid_request"}`, http.StatusBadRequest)
		return
	}

	p.mu.Lock()
	p.calls++
	revoked, entered, release := p.revoked, p.entered, p.release
	p.mu.Unlock()
	p.reply(w, revoked, entered, release)
}

func (p *IssuerState) reply(w http.ResponseWriter, revoked bool, entered chan struct{}, release chan struct{}) {
	if entered != nil {
		select {
		case entered <- struct{}{}:
		default:
		}
		<-release
	}
	if revoked {
		http.Error(w, `{"error":"invalid_grant"}`, http.StatusBadRequest)
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]string{"access_token": p.token(), "refresh_token": nextRefresh})
}

func (p *IssuerState) token() string {
	payload := fmt.Sprintf(`{"exp":%d,"https://api.openai.com/auth":{"chatgpt_account_id":%q}}`, time.Now().Add(time.Hour).Unix(), accountID)
	return "header." + base64.RawURLEncoding.EncodeToString([]byte(payload)) + ".signature"
}

func (p *IssuerState) idToken() string {
	payload := fmt.Sprintf(`{"https://api.openai.com/auth":{"chatgpt_account_id":%q}}`, accountID)
	return "header." + base64.RawURLEncoding.EncodeToString([]byte(payload)) + ".signature"
}

func (p *IssuerState) count() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.calls
}

func (p *ProviderState) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p.mu.Lock()
	p.calls = append(p.calls, r.Header.Get("Authorization"))
	p.mu.Unlock()
	if r.URL.Path != "/api/v1/key" {
		http.NotFound(w, r)
		return
	}
	if strings.Contains(r.Header.Get("Authorization"), oldKey) {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(`{"data":{"label":"controlled-test-provider"}}`))
}

func (p *ProviderState) snapshot() []string {
	p.mu.Lock()
	defer p.mu.Unlock()
	return append([]string(nil), p.calls...)
}
