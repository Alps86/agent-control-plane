package modellzugang

import (
	"fmt"
	"io"
	"net/http"
)

func (p *e2eResponses) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	p.mu.Lock()
	p.requests = append(p.requests, e2eRequest{bearer: r.Header.Get("Authorization"), account: r.Header.Get("ChatGPT-Account-ID"), body: string(body)})
	count, mode := len(p.requests), p.mode
	requestID, model := fmt.Sprintf("local-e2e-request-%d", count), "local-e2e-model"
	if mode == "echo" {
		requestID, model = p.echoID, p.echoModel
	}
	p.metadata = append(p.metadata, e2eMetadata{requestID: requestID, model: model})
	p.mu.Unlock()
	w.Header().Set("x-request-id", requestID)
	w.Header().Set("OpenAI-Model", model)
	if mode == "limit" {
		w.WriteHeader(http.StatusTooManyRequests)
		return
	}
	if mode == "cancel" {
		p.openStream(w, r)
		return
	}
	p.toolRound(w, count, model)
}

func (p *e2eResponses) toolRound(w http.ResponseWriter, count int, model string) {
	w.Header().Set("Content-Type", "text/event-stream")
	if count == 1 {
		fmt.Fprint(w, "event: response.output_item.done\n")
		fmt.Fprint(w, "data: {\"type\":\"response.output_item.done\",\"item\":{\"type\":\"function_call\",\"call_id\":\"connection-1\",\"name\":\"verbindungsstatus_lesen\",\"arguments\":\"{}\"}}\n\n")
	}
	if count > 1 {
		fmt.Fprint(w, "event: response.output_text.delta\n")
		fmt.Fprint(w, "data: {\"type\":\"response.output_text.delta\",\"delta\":\"connected\"}\n\n")
	}
	fmt.Fprint(w, "event: response.completed\n")
	fmt.Fprintf(w, "data: {\"type\":\"response.completed\",\"response\":{\"status\":\"completed\",\"model\":%q}}\n\n", model)
}

func (p *e2eResponses) openStream(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	fmt.Fprint(w, "event: response.output_text.delta\n")
	fmt.Fprint(w, "data: {\"type\":\"response.output_text.delta\",\"delta\":\"Zwischenstand\"}\n\n")
	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}
	p.once.Do(func() { close(p.firstDelta) })
	<-r.Context().Done()
	close(p.cancelled)
}

func (p *e2eResponses) Requests() []e2eRequest {
	p.mu.Lock()
	defer p.mu.Unlock()
	return append([]e2eRequest(nil), p.requests...)
}

func (p *e2eResponses) Metadata() []e2eMetadata {
	p.mu.Lock()
	defer p.mu.Unlock()
	return append([]e2eMetadata(nil), p.metadata...)
}

func (r *e2eReroute) RoundTrip(request *http.Request) (*http.Response, error) {
	if request.URL.Scheme != "https" || request.URL.Host != "chatgpt.com" ||
		request.URL.Path != "/backend-api/codex/responses" {
		return nil, fmt.Errorf("unerwartete Providerroute")
	}
	clone := request.Clone(request.Context())
	redirected := *request.URL
	redirected.Scheme, redirected.Host = r.target.Scheme, r.target.Host
	clone.URL, clone.Host = &redirected, r.target.Host
	return r.base.RoundTrip(clone)
}
