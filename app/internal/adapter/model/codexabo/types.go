package codexabo

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

const endpoint = "https://chatgpt.com/backend-api/codex/responses"

// CredentialPort provides an existing ChatGPT subscription token only in memory.
type CredentialPort interface {
	Access(ctx context.Context) (token, accountID string, err error)
}

// ObservationSink receives one redacted outcome for each provider request.
type ObservationSink interface {
	RecordModelResponse(Observation)
}

// Usage preserves the distinction between an absent count and a measured zero.
type Usage struct {
	InputTokens  *int
	OutputTokens *int
	TotalTokens  *int
}

// Observation contains only provider metadata safe for a technical report.
type Observation struct {
	RequestID string
	Model     string
	Status    string
	Usage     *Usage
}

// Adapter implements Eino's tool-calling chat model interface.
type Adapter struct {
	Model    string
	Auth     CredentialPort
	Client   *http.Client
	Observer ObservationSink
	tools    []*schema.ToolInfo
}

var _ model.ToolCallingChatModel = (*Adapter)(nil)

type ProviderError struct {
	Kind      string
	Status    int
	RequestID string
	Reason    string
}

type streamEvent struct {
	msg *schema.Message
	err error
}
type streamPump struct {
	ctx    context.Context
	cancel context.CancelFunc
	body   io.ReadCloser
	writer *schema.StreamWriter[*schema.Message]
	events <-chan streamEvent
	tick   *time.Ticker
}
type wireError struct {
	Code string `json:"code"`
	Type string `json:"type"`
}
type wirePart struct {
	Type string `json:"type"`
	Text string `json:"text"`
}
type wireItem struct {
	Type      string     `json:"type"`
	CallID    string     `json:"call_id"`
	Name      string     `json:"name"`
	Arguments string     `json:"arguments"`
	Content   []wirePart `json:"content"`
}
type wireUsage struct {
	InputTokens  *int `json:"input_tokens"`
	OutputTokens *int `json:"output_tokens"`
	TotalTokens  *int `json:"total_tokens"`
}
type wireResponse struct {
	Model             string    `json:"model"`
	Status            string    `json:"status"`
	Error             wireError `json:"error"`
	IncompleteDetails struct {
		Reason string `json:"reason"`
	} `json:"incomplete_details"`
	Usage  *wireUsage `json:"usage"`
	Output []wireItem `json:"output"`
}
type wireEvent struct {
	Type     string       `json:"type"`
	Delta    string       `json:"delta"`
	Item     wireItem     `json:"item"`
	Error    wireError    `json:"error"`
	Response wireResponse `json:"response"`
}
type eventParser struct {
	ctx         context.Context
	out         chan<- streamEvent
	observer    ObservationSink
	requestID   string
	headerModel string
	event       string
	data        []string
	pending     []schema.ToolCall
	delivered   bool
	completed   bool
	failed      bool
	observed    bool
}
type wireBuilder struct{}

var safeLabel = regexp.MustCompile(`^[A-Za-z0-9._/-]{1,100}$`)

func New(modelID string, auth CredentialPort, client *http.Client, observers ...ObservationSink) (*Adapter, error) {
	if modelID == "" || auth == nil {
		return nil, fmt.Errorf("subscription model or credential port absent")
	}

	a := &Adapter{Model: modelID, Auth: auth, Client: client}
	if len(observers) > 0 {
		a.Observer = observers[0]
	}

	return a, nil
}

func (a *Adapter) WithTools(tools []*schema.ToolInfo) (model.ToolCallingChatModel, error) {
	for _, t := range tools {
		if t == nil || t.Name == "" {
			return nil, fmt.Errorf("invalid tool definition")
		}
	}

	clone := *a
	clone.tools = append([]*schema.ToolInfo(nil), tools...)
	return &clone, nil
}

func (e *ProviderError) Error() string {
	return fmt.Sprintf("codex subscription %s (HTTP %d, reason %s, request %s)", e.Kind, e.Status, e.Reason, e.RequestID)
}

func (a *Adapter) statusKind(status int) string {
	if status == 401 || status == 403 {
		return "authentication"
	}

	if status == 429 {
		return "limit"
	}

	if status >= 500 {
		return "unavailable"
	}

	return "rejected"
}

func (a *Adapter) safeID(value string) string {
	if safeLabel.MatchString(value) {
		return value
	}

	return ""
}

func (p *eventParser) safeID(value string) string {
	if safeLabel.MatchString(value) {
		return value
	}

	return ""
}

func (p *eventParser) providerReason(code, reason string) (string, string) {
	value := strings.ToLower(code + " " + reason)
	if strings.Contains(value, "rate_limit") || strings.Contains(value, "usage_limit") || strings.Contains(value, "quota") || strings.Contains(value, "too_many_requests") || strings.Contains(value, "limit_reached") {
		return "limit", "provider_limit"
	}

	if strings.Contains(value, "expired") || strings.Contains(value, "unauthorized") || strings.Contains(value, "authentication") {
		return "authentication", "invalid_session"
	}

	if strings.Contains(value, "content_filter") || strings.Contains(value, "safety") {
		return "rejected", "safety"
	}

	return "rejected", "provider_error"
}

func (p *eventParser) count(value *int) *int {
	if value == nil || *value < 0 || *value > 1000000000 {
		return nil
	}

	copy := *value
	return &copy
}

func (p *eventParser) usage(raw *wireUsage) *Usage {
	if raw == nil {
		return nil
	}

	return &Usage{InputTokens: p.count(raw.InputTokens), OutputTokens: p.count(raw.OutputTokens), TotalTokens: p.count(raw.TotalTokens)}
}

func (p *eventParser) einoUsage(usage *Usage) *schema.TokenUsage {
	if usage == nil || usage.InputTokens == nil || usage.OutputTokens == nil || usage.TotalTokens == nil {
		return nil
	}

	return &schema.TokenUsage{PromptTokens: *usage.InputTokens, CompletionTokens: *usage.OutputTokens, TotalTokens: *usage.TotalTokens}
}

func (p *eventParser) observe(o Observation) {
	p.observed = true
	if p.observer != nil {
		p.observer.RecordModelResponse(o)
	}
}

func (p *eventParser) aborted() {
	if p.observed {
		return
	}

	p.observe(Observation{RequestID: p.requestID, Model: p.headerModel, Status: "canceled"})
}
