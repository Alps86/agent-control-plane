package uibootstrap

import (
	"bufio"
	"os/exec"
	"testing"
)

type Suite struct {
	t       *testing.T
	server  *exec.Cmd
	browser *exec.Cmd
	stdin   *bufio.Writer
	input   *bufio.Scanner
	baseURL string
	fixture string
	view    string
	narrow  bool
	page    Page
}

type Page struct {
	URL        string `json:"url"`
	Title      string `json:"title"`
	Heading    string `json:"heading"`
	Text       string `json:"text"`
	Alert      string `json:"alert"`
	Links      []Link `json:"links"`
	Active     string `json:"active"`
	Horizontal bool   `json:"horizontal"`
	Help       string `json:"help"`
}

type Link struct {
	Label   string `json:"label"`
	Href    string `json:"href"`
	Visible bool   `json:"visible"`
}

type Reply struct {
	OK    bool   `json:"ok"`
	Page  Page   `json:"page"`
	Error string `json:"error"`
}
