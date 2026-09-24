package modellverbindung

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"agentcontrolplane/app/internal/adapter/model/codexauth"
)

const mockAccountID = "account-test"
const mockForeignAccountID = "account-foreign"
const mockRefreshToken = "synthetic-refresh-private"
const mockRotatedRefreshToken = "synthetic-refresh-rotated-private"

func (f *FakeIssuer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/api/accounts/deviceauth/usercode" {
		f.userCode(w, r)
		return
	}

	if r.URL.Path == "/api/accounts/deviceauth/token" {
		f.poll(w, r)
		return
	}

	if r.URL.Path == "/oauth/token" {
		f.token(w, r)
		return
	}

	http.NotFound(w, r)
}

func (f *FakeIssuer) userCode(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var request map[string]string
	if json.NewDecoder(r.Body).Decode(&request) != nil || request["client_id"] != codexauth.CodexCLIClientID {
		f.problem(w, "invalid_request")
		return
	}

	if f.startDisabled {
		http.Error(w, `{"error":"device_code_disabled"}`, http.StatusNotFound)
		return
	}

	f.starts++
	f.json(w, map[string]string{"device_auth_id": "device-auth-test", "user_code": "ABCD-EFGH", "interval": "0"})
}

func (f *FakeIssuer) poll(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var request map[string]string
	if json.NewDecoder(r.Body).Decode(&request) != nil || request["device_auth_id"] != "device-auth-test" || request["user_code"] != "ABCD-EFGH" {
		f.problem(w, "invalid_request")
		return
	}

	f.polls++
	f.pollResult(w)
}

func (f *FakeIssuer) pollResult(w http.ResponseWriter) {
	if f.pollOutcome == "expired" {
		f.problem(w, "expired_token")
		return
	}

	if f.pollOutcome == "denied" {
		f.problem(w, "access_denied")
		return
	}

	if f.pollOutcome != "connected" {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	f.json(w, map[string]string{"authorization_code": "synthetic-auth-code", "code_challenge": "synthetic-challenge", "code_verifier": "synthetic-verifier"})
}

func (f *FakeIssuer) token(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if strings.HasPrefix(r.Header.Get("Content-Type"), "application/x-www-form-urlencoded") {
		f.authCode(w, r)
		return
	}

	f.refresh(w, r)
}

func (f *FakeIssuer) authCode(w http.ResponseWriter, r *http.Request) {
	if r.ParseForm() != nil || r.FormValue("grant_type") != "authorization_code" || r.FormValue("client_id") != codexauth.CodexCLIClientID {
		f.problem(w, "invalid_request")
		return
	}

	if r.FormValue("code") != "synthetic-auth-code" || r.FormValue("code_verifier") != "synthetic-verifier" || !strings.HasSuffix(r.FormValue("redirect_uri"), "/deviceauth/callback") {
		f.problem(w, "invalid_request")
		return
	}

	f.issuedToken = f.accessToken(f.issueLifetime, mockAccountID)
	f.json(w, map[string]string{"id_token": f.idToken(), "access_token": f.issuedToken, "refresh_token": mockRefreshToken})
}

func (f *FakeIssuer) refresh(w http.ResponseWriter, r *http.Request) {
	var request map[string]any
	if json.NewDecoder(r.Body).Decode(&request) != nil || request["grant_type"] != "refresh_token" || request["client_id"] != codexauth.CodexCLIClientID || request["refresh_token"] != mockRefreshToken {
		f.problem(w, "invalid_request")
		return
	}

	f.refreshes++
	if f.refreshOutcome == "invalid_grant" {
		f.problem(w, "invalid_grant")
		return
	}

	f.refreshResult(w)
}

func (f *FakeIssuer) refreshResult(w http.ResponseWriter) {
	if f.refreshOutcome == "missing_access" {
		f.json(w, map[string]string{"id_token": f.idToken(), "refresh_token": mockRotatedRefreshToken})
		return
	}

	if f.refreshOutcome == "foreign_account" {
		f.rotatedToken = f.accessToken(3600, mockForeignAccountID)
		f.json(w, map[string]string{"access_token": f.rotatedToken, "refresh_token": mockRotatedRefreshToken})
		return
	}

	f.rotatedToken = f.accessToken(3600, mockAccountID)
	if f.refreshOutcome == "matching_no_id" {
		f.json(w, map[string]string{"access_token": f.rotatedToken, "refresh_token": mockRotatedRefreshToken})
		return
	}

	f.json(w, map[string]string{"id_token": f.idToken(), "access_token": f.rotatedToken, "refresh_token": mockRotatedRefreshToken})
}

func (f *FakeIssuer) problem(w http.ResponseWriter, reason string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	f.json(w, map[string]string{"error": reason})
}

func (f *FakeIssuer) json(w http.ResponseWriter, value any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(value)
}

func (f *FakeIssuer) idToken() string {
	payload := `{"https://api.openai.com/auth":{"chatgpt_account_id":"` + mockAccountID + `"}}`
	return "header." + base64.RawURLEncoding.EncodeToString([]byte(payload)) + ".signature"
}

func (f *FakeIssuer) accessToken(seconds int64, accountID string) string {
	expires := time.Now().Unix() + seconds
	payload := fmt.Sprintf(`{"exp":%d,"https://api.openai.com/auth":{"chatgpt_account_id":%q}}`, expires, accountID)
	return "header." + base64.RawURLEncoding.EncodeToString([]byte(payload)) + ".signature"
}
