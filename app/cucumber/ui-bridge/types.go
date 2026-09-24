package uibridge

import (
	"bufio"
	"net/http/httptest"
	"os/exec"
	"testing"

	"agentcontrolplane/ui/bridge"
)

type Suite struct {
	t         *testing.T
	bridge    *bridge.Bridge
	server    *httptest.Server
	browser   *exec.Cmd
	stdin     *bufio.Writer
	stdout    *bufio.Scanner
	data      map[string]any
	page      string
	part      string
	fixture   string
	view      string
	asset     string
	denied    []string
	status    int
	built     bool
	check     string
	checkPart string
}

type BrowserReply struct {
	OK    bool        `json:"ok"`
	Page  BrowserPage `json:"page"`
	Error string      `json:"error"`
}

type BrowserPage struct {
	Title   string `json:"title"`
	Heading string `json:"heading"`
	Text    string `json:"text"`
	Help    string `json:"help"`
}
