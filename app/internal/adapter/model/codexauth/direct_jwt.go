package codexauth

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"time"

	"agentcontrolplane/app/internal/app/modellverbindung"
)

func (d *Direct) bundle(tokens tokenResponse) (modellverbindung.TokenBundle, error) {
	idClaims, err := d.claims(tokens.IDToken)
	if err != nil || idClaims.Auth.AccountID == "" || strings.TrimSpace(idClaims.Auth.AccountID) != idClaims.Auth.AccountID {
		return modellverbindung.TokenBundle{}, errProtocol
	}

	if len(d.allowed) > 0 && !d.allowed[idClaims.Auth.AccountID] {
		return modellverbindung.TokenBundle{}, modellverbindung.ErrReauthenticationRequired
	}

	accessClaims, err := d.claims(tokens.AccessToken)
	if err != nil || accessClaims.ExpiresAt <= time.Now().Unix() || accessClaims.Auth.AccountID != idClaims.Auth.AccountID || tokens.RefreshToken == "" {
		return modellverbindung.TokenBundle{}, errProtocol
	}

	return modellverbindung.TokenBundle{IDToken: tokens.IDToken, AccessToken: tokens.AccessToken, RefreshToken: tokens.RefreshToken, AccountID: idClaims.Auth.AccountID, ExpiresAt: time.Unix(accessClaims.ExpiresAt, 0)}, nil
}

func (d *Direct) claims(token string) (jwtClaims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
		return jwtClaims{}, errProtocol
	}

	data, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return jwtClaims{}, errProtocol
	}

	var claims jwtClaims
	if json.Unmarshal(data, &claims) != nil {
		return jwtClaims{}, errProtocol
	}

	return claims, nil
}
