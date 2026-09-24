//go:build browser

package browser

import (
	"sync/atomic"
	"testing"

	"agentcontrolplane/ui/bridge"
)

// Suite prüft die öffentlichen Settings-Ansichten über die UI-Bridge.
type Suite struct {
	t              *testing.T
	bridge         *bridge.Bridge
	state          string
	html           string
	issuerStarts   int
	connectionLive bool
	browserRace    browserRaceResult
}

type browserRaceResult struct {
	State       string `json:"state"`
	Text        string `json:"text"`
	CodeVisible bool   `json:"codeVisible"`
	Code        string `json:"code"`
	Error       string `json:"error"`
}

type raceServer struct {
	bridge       *bridge.Bridge
	mode         string
	origin       string
	starts       atomic.Int32
	statuses     atomic.Int32
	cancels      atomic.Int32
	startOrigin  atomic.Value
	cancelOrigin atomic.Value
}
