package preview

import "agentcontrolplane/ui/bridge"

// Preview serves the embedded UI with synthetic fixture data on loopback.
type Preview struct {
	bridge *bridge.Bridge
}
