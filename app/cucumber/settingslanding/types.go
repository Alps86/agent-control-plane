package settingslanding

import (
	"bufio"
	"io"
	"os/exec"
	"testing"
)

type Suite struct {
	t              *testing.T
	binary         string
	process        *exec.Cmd
	exited         chan error
	baseURL        string
	status         int
	contentType    string
	body           string
	browser        browserResult
	browserProcess *exec.Cmd
	browserInput   io.WriteCloser
	browserOutput  *bufio.Scanner
	unsafeFailed   bool
	unsafeOutput   string
}

type browserResult struct {
	Path          string `json:"path"`
	ProviderTitle string `json:"providerTitle"`
	Links         int    `json:"links"`
	Error         string `json:"error"`
}
