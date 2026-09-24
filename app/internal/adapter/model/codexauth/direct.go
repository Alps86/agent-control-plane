package codexauth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"agentcontrolplane/app/internal/app/modellverbindung"
)

func NewDirect(config Config) (*Direct, error) {
	issuer, err := url.Parse(config.Issuer)
	if err != nil || !config.validIssuer(issuer) || config.ClientID != CodexCLIClientID {
		return nil, errProtocol
	}

	client := config.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 15 * time.Second}
	}

	copyClient := *client
	copyClient.CheckRedirect = config.noRedirect
	return &Direct{issuer: strings.TrimRight(config.Issuer, "/"), clientID: config.ClientID, http: &copyClient, allowed: config.workspaceSet()}, nil
}

func (config Config) workspaceSet() map[string]bool {
	allowed := make(map[string]bool, len(config.AllowedWorkspaceIDs))
	for _, id := range config.AllowedWorkspaceIDs {
		allowed[id] = true
	}

	return allowed
}

func (config Config) validIssuer(issuer *url.URL) bool {
	if issuer == nil || issuer.Host == "" || issuer.User != nil || issuer.RawQuery != "" || issuer.Fragment != "" || (issuer.Path != "" && issuer.Path != "/") {
		return false
	}

	if issuer.Scheme == "https" {
		return true
	}

	return issuer.Scheme == "http" && (issuer.Hostname() == "127.0.0.1" || issuer.Hostname() == "localhost")
}

func (config Config) noRedirect(_ *http.Request, _ []*http.Request) error {
	return http.ErrUseLastResponse
}

func (d *Direct) Start(ctx context.Context) (modellverbindung.Challenge, error) {
	resp, err := d.postJSON(ctx, "/api/accounts/deviceauth/usercode", map[string]string{"client_id": d.clientID})
	if err != nil {
		return modellverbindung.Challenge{}, err
	}

	if resp.status == http.StatusNotFound {
		return modellverbindung.Challenge{}, modellverbindung.ErrDeviceCodeDisabled
	}

	if resp.status != http.StatusOK {
		return modellverbindung.Challenge{}, errProtocol
	}

	return d.challenge(resp.body)
}

func (d *Direct) challenge(body []byte) (modellverbindung.Challenge, error) {
	var result userCodeResponse
	if json.Unmarshal(body, &result) != nil || result.DeviceAuthID == "" {
		return modellverbindung.Challenge{}, errProtocol
	}

	code := result.UserCode
	if code == "" {
		code = result.UserCodeAlt
	}

	interval, err := d.interval(result.Interval)
	if err != nil || interval < 0 || code == "" {
		return modellverbindung.Challenge{}, errProtocol
	}

	return modellverbindung.Challenge{VerificationURL: d.issuer + "/codex/device", UserCode: code, DeviceAuthID: result.DeviceAuthID, IntervalSeconds: interval, ExpiresAt: time.Now().Add(15 * time.Minute)}, nil
}

func (d *Direct) interval(raw string) (int, error) {
	if strings.TrimSpace(raw) == "" {
		return 0, nil
	}

	seconds, err := strconv.ParseUint(strings.TrimSpace(raw), 10, 64)
	if err != nil {
		return 0, err
	}

	if seconds > 900 {
		return 900, nil
	}

	return int(seconds), nil
}

func (d *Direct) Poll(ctx context.Context, challenge modellverbindung.Challenge) (modellverbindung.PollResult, error) {
	request := map[string]string{"device_auth_id": challenge.DeviceAuthID, "user_code": challenge.UserCode}
	resp, err := d.postJSON(ctx, "/api/accounts/deviceauth/token", request)
	if err != nil {
		return modellverbindung.PollResult{}, err
	}

	return d.pollResponse(ctx, resp)
}

func (d *Direct) pollResponse(ctx context.Context, resp response) (modellverbindung.PollResult, error) {
	if resp.status == http.StatusForbidden || resp.status == http.StatusNotFound {
		return modellverbindung.PollResult{State: "pending"}, nil
	}

	if resp.status == http.StatusBadRequest {
		return d.pollFailure(resp.body)
	}

	if resp.status != http.StatusOK {
		return modellverbindung.PollResult{}, errProtocol
	}

	return d.completePoll(ctx, resp.body)
}

func (d *Direct) pollFailure(body []byte) (modellverbindung.PollResult, error) {
	var failure oauthError
	if json.Unmarshal(body, &failure) != nil {
		return modellverbindung.PollResult{}, errProtocol
	}

	if failure.Error == "expired_token" {
		return modellverbindung.PollResult{State: "expired"}, nil
	}

	if failure.Error == "access_denied" {
		return modellverbindung.PollResult{State: "denied"}, nil
	}

	return modellverbindung.PollResult{}, errProtocol
}

func (d *Direct) completePoll(ctx context.Context, body []byte) (modellverbindung.PollResult, error) {
	var code pollResponse
	if json.Unmarshal(body, &code) != nil || code.AuthorizationCode == "" || code.CodeVerifier == "" {
		return modellverbindung.PollResult{}, errProtocol
	}

	bundle, err := d.exchange(ctx, code)
	if err != nil {
		return modellverbindung.PollResult{}, modellverbindung.ErrExchangeFailed
	}

	return modellverbindung.PollResult{State: "connected", Tokens: bundle}, nil
}

func (d *Direct) exchange(ctx context.Context, code pollResponse) (modellverbindung.TokenBundle, error) {
	form := url.Values{"grant_type": {"authorization_code"}, "client_id": {d.clientID}, "code": {code.AuthorizationCode}, "redirect_uri": {d.issuer + "/deviceauth/callback"}, "code_verifier": {code.CodeVerifier}}
	resp, err := d.postForm(ctx, "/oauth/token", form)
	if err != nil || resp.status != http.StatusOK {
		return modellverbindung.TokenBundle{}, errProtocol
	}

	var tokens tokenResponse
	if json.Unmarshal(resp.body, &tokens) != nil {
		return modellverbindung.TokenBundle{}, errProtocol
	}

	return d.bundle(tokens)
}

func (d *Direct) Refresh(ctx context.Context, prior modellverbindung.TokenBundle) (modellverbindung.TokenBundle, error) {
	request := map[string]string{"grant_type": "refresh_token", "client_id": d.clientID, "refresh_token": prior.RefreshToken}
	resp, err := d.postJSON(ctx, "/oauth/token", request)
	if err != nil {
		return modellverbindung.TokenBundle{}, err
	}

	if d.reauthentication(resp) {
		return modellverbindung.TokenBundle{}, modellverbindung.ErrReauthenticationRequired
	}

	if resp.status != http.StatusOK {
		return modellverbindung.TokenBundle{}, errProtocol
	}

	return d.refreshedBundle(resp.body, prior)
}

func (d *Direct) reauthentication(resp response) bool {
	if resp.status == http.StatusUnauthorized {
		return true
	}

	if resp.status != http.StatusBadRequest {
		return false
	}

	var failure oauthError
	if json.Unmarshal(resp.body, &failure) != nil {
		return false
	}

	return failure.Error == "invalid_grant" || strings.HasPrefix(failure.Error, "refresh_token_")
}

func (d *Direct) refreshedBundle(body []byte, prior modellverbindung.TokenBundle) (modellverbindung.TokenBundle, error) {
	var tokens tokenResponse
	if json.Unmarshal(body, &tokens) != nil || tokens.AccessToken == "" {
		return modellverbindung.TokenBundle{}, modellverbindung.ErrReauthenticationRequired
	}

	d.mergeTokens(&tokens, prior)
	next, err := d.bundle(tokens)
	if err != nil || next.AccountID != prior.AccountID {
		return modellverbindung.TokenBundle{}, modellverbindung.ErrReauthenticationRequired
	}

	return next, nil
}

func (d *Direct) mergeTokens(tokens *tokenResponse, prior modellverbindung.TokenBundle) {
	if tokens.IDToken == "" {
		tokens.IDToken = prior.IDToken
	}

	if tokens.RefreshToken == "" {
		tokens.RefreshToken = prior.RefreshToken
	}

}

var _ modellverbindung.Client = (*Direct)(nil)
