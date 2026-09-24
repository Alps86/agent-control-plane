package modellanbieter

import (
	"errors"

	"agentcontrolplane/app/internal/app/modellverbindung"
)

var ErrUntrustedBind = errors.New("Codex settings require a loopback listener bind address")

type Handler struct {
	flow        modellverbindung.DeviceFlow
	trustedBind bool
}
