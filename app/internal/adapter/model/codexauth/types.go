package codexauth

import (
	"errors"
	"net/http"
)

var errProtocol = errors.New("codex direct auth protocol failure")

// CodexCLIClientID is published by openai/codex login/src/auth/manager.rs.
const CodexCLIClientID = "app_EMoamEEZ73f0CkXaXp7hrann"

type Config struct {
	Issuer              string
	ClientID            string
	HTTPClient          *http.Client
	AllowedWorkspaceIDs []string
}

type Direct struct {
	issuer   string
	clientID string
	http     *http.Client
	allowed  map[string]bool
}

type userCodeResponse struct {
	DeviceAuthID string `json:"device_auth_id"`
	UserCode     string `json:"user_code"`
	UserCodeAlt  string `json:"usercode"`
	Interval     string `json:"interval"`
}

type pollResponse struct {
	AuthorizationCode string `json:"authorization_code"`
	CodeChallenge     string `json:"code_challenge"`
	CodeVerifier      string `json:"code_verifier"`
}

type tokenResponse struct {
	IDToken      string `json:"id_token"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type oauthError struct {
	Error string `json:"error"`
}

type jwtClaims struct {
	ExpiresAt int64 `json:"exp"`
	Auth      struct {
		AccountID string `json:"chatgpt_account_id"`
	} `json:"https://api.openai.com/auth"`
}

type response struct {
	status int
	body   []byte
}
