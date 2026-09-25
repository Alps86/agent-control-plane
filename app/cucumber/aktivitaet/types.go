package aktivitaet

import (
	"bufio"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os/exec"
	"testing"

	"agentcontrolplane/app/internal/adapter/sqlite"
	domainaktivitaet "agentcontrolplane/app/internal/domain/aktivitaet"
	portaktivitaet "agentcontrolplane/app/internal/port/aktivitaet"
)

var ErrInjectedWrite = errors.New("kontrollierter Fehler beim zweiten Aktivitätsereignis")

type Suite struct {
	t             *testing.T
	binary        string
	database      string
	address       string
	process       *exec.Cmd
	client        *http.Client
	status        int
	body          []byte
	organizations map[string]string
	projects      map[string]string
	agents        map[string]string
	taskID        string
	before        []Event
	browser       *exec.Cmd
	input         *bufio.Writer
	output        *bufio.Scanner
	page          Page
	fixtureServer *httptest.Server
	fixtureDB     *sqlite.Database
	failingStore  *FailSecondStore
}

type Story66 struct {
	*Suite
	query       url.Values
	pageEvents  []Event
	allEvents   []Event
	csvRows     [][]string
	csvBody     []byte
	contentType string
	disposition string
	pageBefore  []Event
	rowsBefore  [][]string
	selected    Event
}

// FailSecondStore injiziert nur im HTTP-Abnahmefixture den zweiten Schreibfehler.
type FailSecondStore struct {
	Underlying portaktivitaet.Store
	Calls      int
	TaskID     string
}

func (s *FailSecondStore) Record(ctx context.Context, event domainaktivitaet.Event) error {
	s.Calls++
	s.TaskID = event.TaskID
	if s.Calls == 2 {
		return ErrInjectedWrite
	}

	return s.Underlying.Record(ctx, event)
}

func (s *FailSecondStore) ListEvents(ctx context.Context, org string) ([]domainaktivitaet.Event, error) {
	return s.Underlying.ListEvents(ctx, org)
}

type Event struct {
	ID             string `json:"id"`
	OrganizationID string `json:"organization_id"`
	ProjectID      string `json:"project_id"`
	TaskID         string `json:"task_id"`
	Kind           string `json:"kind"`
	Actor          string `json:"actor"`
	Source         string `json:"source"`
	OccurredAt     string `json:"occurred_at"`
	ObjectTitle    string `json:"object_title"`
	DeepLink       string `json:"deep_link"`
	AssigneeID     string `json:"assignee_id"`
	AssigneeName   string `json:"assignee_name"`
}

type EventList struct {
	Events []Event `json:"events"`
}

type Page struct {
	URL       string   `json:"url"`
	Heading   string   `json:"heading"`
	Text      string   `json:"text"`
	Links     []string `json:"links"`
	LinkTexts []string `json:"linkTexts"`
	Times     []string `json:"times"`
}

type BrowserReply struct {
	OK       bool   `json:"ok"`
	Error    string `json:"error"`
	Page     Page   `json:"page"`
	Download string `json:"download"`
}
