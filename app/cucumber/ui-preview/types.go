package uipreview

import (
	"bufio"
	"os/exec"
	"testing"
)

type Suite struct {
	t           *testing.T
	server      *exec.Cmd
	browser     *exec.Cmd
	stdin       *bufio.Writer
	stdout      *bufio.Scanner
	port        int
	binary      string
	baseURL     string
	page        Page
	full        string
	part        string
	body        string
	status      int
	copyDir     string
	changedName string
	responses   []Response
	portErrors  []string
}

type Page struct {
	URL     string  `json:"url"`
	Title   string  `json:"title"`
	Heading string  `json:"heading"`
	Text    string  `json:"text"`
	Links   []Link  `json:"links"`
	Help    string  `json:"help"`
	Assets  []Asset `json:"assets"`
}

type Asset struct {
	Path   string `json:"path"`
	Status int    `json:"status"`
}

type Link struct {
	Label string `json:"label"`
	Href  string `json:"href"`
}

type Reply struct {
	OK    bool   `json:"ok"`
	Page  Page   `json:"page"`
	Error string `json:"error"`
}

type Response struct {
	Body   string
	Status int
	Header string
}
