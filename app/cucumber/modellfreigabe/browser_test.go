package modellfreigabe

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

type browserPage struct {
	URL     string `json:"url"`
	Heading string `json:"heading"`
	Text    string `json:"text"`
	HTML    string `json:"html"`
	Fields  []struct {
		Name  string `json:"name"`
		Value string `json:"value"`
		Type  string `json:"type"`
	} `json:"fields"`
	Links []struct {
		Text string `json:"text"`
		Href string `json:"href"`
	} `json:"links"`
}

type browserResponse struct {
	URL    string `json:"url"`
	Status int    `json:"status"`
	Body   string `json:"body"`
}

type browserReply struct {
	OK        bool              `json:"ok"`
	Error     string            `json:"error"`
	Page      browserPage       `json:"page"`
	Responses []browserResponse `json:"responses"`
}

type chromeBrowser struct {
	process *exec.Cmd
	input   *bufio.Writer
	output  *bufio.Scanner
	pipe    *os.File
	page    browserPage
	replies []browserResponse
}

func (b *chromeBrowser) start() error {
	script, err := filepath.Abs("browser.mjs")
	if err != nil {
		return err
	}
	b.process = exec.Command("node", script)
	b.process.Stderr = os.Stderr
	b.process.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	in, err := b.process.StdinPipe()
	if err != nil {
		return err
	}
	out, err := b.process.StdoutPipe()
	if err != nil {
		return err
	}
	b.input = bufio.NewWriter(in)
	b.output = bufio.NewScanner(out)
	b.output.Buffer(make([]byte, 64*1024), 8*1024*1024)
	return b.process.Start()
}

func (b *chromeBrowser) command(input map[string]any) error {
	encoded, err := json.Marshal(input)
	if err != nil {
		return err
	}
	if _, err := b.input.Write(append(encoded, '\n')); err != nil {
		return err
	}
	if err := b.input.Flush(); err != nil {
		return err
	}
	if !b.output.Scan() {
		return fmt.Errorf("Chrome endete: %v", b.output.Err())
	}
	var reply browserReply
	if err := json.Unmarshal(b.output.Bytes(), &reply); err != nil {
		return err
	}
	if !reply.OK {
		return fmt.Errorf("Chrome: %s", reply.Error)
	}
	b.page = reply.Page
	b.replies = append(b.replies, reply.Responses...)
	return nil
}

func (b *chromeBrowser) stop() {
	if b.process == nil {
		return
	}
	_ = syscall.Kill(-b.process.Process.Pid, syscall.SIGKILL)
	_ = b.process.Wait()
	b.process = nil
}
