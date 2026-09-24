package modellzugang

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"time"

	"agentcontrolplane/app/internal/adapter/model/codexauth"
)

func (s *httpE2ESuite) prepareIssuer() {
	s.issuer = &e2eIssuer{accountID: "local-e2e-account", refreshToken: "local-e2e-refresh-private"}
	s.issuer.accessToken = s.issuer.jwt(time.Now().Add(time.Hour).Unix())
}

func (i *e2eIssuer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/api/accounts/deviceauth/usercode" {
		i.start(w, r)
		return
	}
	if r.URL.Path == "/api/accounts/deviceauth/token" {
		i.poll(w, r)
		return
	}
	if r.URL.Path == "/oauth/token" {
		i.exchange(w, r)
		return
	}
	http.NotFound(w, r)
}

func (i *e2eIssuer) start(w http.ResponseWriter, r *http.Request) {
	var body map[string]string
	if json.NewDecoder(r.Body).Decode(&body) != nil || body["client_id"] != codexauth.CodexCLIClientID {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	i.mu.Lock()
	i.starts++
	i.mu.Unlock()
	i.json(w, map[string]string{"device_auth_id": "local-device-id", "user_code": "E2E-LOCAL", "interval": "0"})
}

func (i *e2eIssuer) poll(w http.ResponseWriter, r *http.Request) {
	var body map[string]string
	if json.NewDecoder(r.Body).Decode(&body) != nil || body["device_auth_id"] != "local-device-id" || body["user_code"] != "E2E-LOCAL" {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	i.mu.Lock()
	i.polls++
	i.mu.Unlock()
	i.json(w, map[string]string{"authorization_code": "local-auth-code", "code_challenge": "local-challenge", "code_verifier": "local-verifier"})
}

func (i *e2eIssuer) exchange(w http.ResponseWriter, r *http.Request) {
	if r.ParseForm() != nil || r.FormValue("grant_type") != "authorization_code" ||
		r.FormValue("client_id") != codexauth.CodexCLIClientID || r.FormValue("code") != "local-auth-code" ||
		r.FormValue("code_verifier") != "local-verifier" {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	i.mu.Lock()
	i.exchanges++
	i.mu.Unlock()
	i.json(w, map[string]string{"id_token": i.jwt(0), "access_token": i.accessToken, "refresh_token": i.refreshToken})
}

func (i *e2eIssuer) jwt(exp int64) string {
	claims := map[string]any{"exp": exp, "https://api.openai.com/auth": map[string]string{"chatgpt_account_id": i.accountID}}
	data, _ := json.Marshal(claims)
	return "header." + base64.RawURLEncoding.EncodeToString(data) + ".signature"
}

func (i *e2eIssuer) json(w http.ResponseWriter, value any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(value)
}

func (i *e2eIssuer) Counts() (int, int, int) {
	i.mu.Lock()
	defer i.mu.Unlock()
	return i.starts, i.polls, i.exchanges
}
