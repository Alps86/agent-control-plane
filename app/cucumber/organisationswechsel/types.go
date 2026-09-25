package organisationswechsel

import (
	"agentcontrolplane/app/internal/domain/rechte"
	"bufio"
	"io"
	"net/http"
	"os"
	"os/exec"
	"testing"
)

type Suite struct {
	t                         *testing.T
	binary, database, address string
	process                   *exec.Cmd
	exited                    chan error
	logFile                   *os.File
	client                    *http.Client
	organizations             map[string]Organization
	response                  Response
	previous                  Response
	baseline                  map[string]any
	initialRules              map[string]map[string]any
	revokeRevision            any
	revokedSnapshot           map[string]any
	lastPreview               map[string]any
	lastArea                  string
	lastOrganization          string
	foreignServer             string
	foreignClose              func()
	foreignReplies            []Response
	unknownReplies            []Response
	browser                   *exec.Cmd
	stdin                     *bufio.Writer
	browserInput              io.WriteCloser
	stdout                    *bufio.Scanner
	page                      BrowserPage
}

type FixedIdentity struct{ Actor rechte.Actor }

type BrowserPage struct {
	URL           string   `json:"url"`
	Text          string   `json:"text"`
	Heading       string   `json:"heading"`
	Status        int      `json:"status"`
	SelectedNames []string `json:"selectedNames"`
}

type Organization struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type Response struct {
	Status int
	Body   []byte
	Header http.Header
}
