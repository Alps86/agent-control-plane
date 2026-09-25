package geheimnisreferenzen

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

func (i *Issuer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/api/accounts/deviceauth/usercode" {
		_ = json.NewEncoder(w).Encode(map[string]string{"device_auth_id": "story79-device", "user_code": "STORY-79", "interval": "0"})
		return
	}
	if r.URL.Path == "/api/accounts/deviceauth/token" {
		_ = json.NewEncoder(w).Encode(map[string]string{"authorization_code": "story79-code", "code_verifier": "story79-verifier"})
		return
	}
	if r.URL.Path == "/oauth/token" {
		i.token(w, r)
		return
	}
	http.NotFound(w, r)
}

func (i *Issuer) token(w http.ResponseWriter, r *http.Request) {
	if i.revoked.Load() {
		http.Error(w, `{"error":"invalid_grant"}`, http.StatusBadRequest)
		return
	}
	var body map[string]string
	_ = json.NewDecoder(r.Body).Decode(&body)
	expiry := time.Now().Add(time.Hour)
	if i.short.Load() {
		expiry = time.Now().Add(time.Minute)
	}
	if body["grant_type"] == "refresh_token" {
		expiry = time.Now().Add(time.Hour)
	}
	_ = json.NewEncoder(w).Encode(map[string]string{"access_token": i.jwt(expiry), "refresh_token": testRefresh, "id_token": i.jwt(time.Now().Add(time.Hour))})
}

func (i *Issuer) jwt(expiry time.Time) string {
	payload := fmt.Sprintf(`{"exp":%d,"https://api.openai.com/auth":{"chatgpt_account_id":"story79-account"}}`, expiry.Unix())
	return "header." + base64.RawURLEncoding.EncodeToString([]byte(payload)) + ".signature"
}

func (s *Suite) providerResponse(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/api/v1/chat/completions" {
		s.providerCalls.Add(1)
		if r.Header.Get("Authorization") != "Bearer "+testKey {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"ok"}}]}`))
		return
	}
	if r.URL.Path == "/api/v1/key" && strings.HasSuffix(r.Header.Get("Authorization"), testKey) {
		_, _ = w.Write([]byte(`{"data":{"label":"synthetic"}}`))
		return
	}
	http.NotFound(w, r)
}
