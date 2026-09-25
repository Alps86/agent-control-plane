package modellzugang

import (
	"fmt"
	"io"
	"net/http"
)

func (p *e2eResponses) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	count, mode, metadata := p.recordRequest(r)
	p.setMetadata(w, metadata)
	if mode == "limit" {
		w.WriteHeader(http.StatusTooManyRequests)
		return
	}
	if mode == "cancel" {
		p.openStream(w, r)
		return
	}
	p.toolRound(w, count, metadata)
}

func (p *e2eResponses) recordRequest(r *http.Request) (int, string, e2eMetadata) {
	body, _ := io.ReadAll(r.Body)
	p.mu.Lock()
	defer p.mu.Unlock()
	p.requests = append(p.requests, e2eRequest{bearer: r.Header.Get("Authorization"), account: r.Header.Get("ChatGPT-Account-ID"), body: string(body)})
	count, mode := len(p.requests), p.mode
	metadata := p.responseMetadata(mode, count)
	p.metadata = append(p.metadata, metadata)
	return count, mode, metadata
}

func (p *e2eResponses) setMetadata(w http.ResponseWriter, metadata e2eMetadata) {
	w.Header().Set("x-request-id", metadata.requestID)
	if metadata.modelPresent {
		w.Header().Set("OpenAI-Model", metadata.model)
	}
}

func (p *e2eResponses) responseMetadata(mode string, count int) e2eMetadata {
	entry := e2eMetadata{requestID: fmt.Sprintf("local-e2e-request-%d", count), model: "local-e2e-model", modelPresent: true}
	if mode == "echo" {
		entry.requestID, entry.model = p.echoID, p.echoModel
	}
	if mode == "mismatch" {
		entry.model = "other-local-model"
	}
	if mode == "missing" {
		entry.model = ""
		entry.modelPresent = count > 1
	}
	return entry
}

func (p *e2eResponses) toolRound(w http.ResponseWriter, count int, metadata e2eMetadata) {
	w.Header().Set("Content-Type", "text/event-stream")
	if count == 1 {
		fmt.Fprint(w, "event: response.output_item.done\n")
		fmt.Fprint(w, "data: {\"type\":\"response.output_item.done\",\"item\":{\"type\":\"function_call\",\"call_id\":\"connection-1\",\"name\":\"verbindungsstatus_lesen\",\"arguments\":\"{}\"}}\n\n")
	}
	if count > 1 {
		p.writeAnswer(w)
	}
	p.writeCompletion(w, count, metadata)
}

func (p *e2eResponses) writeAnswer(w http.ResponseWriter) {
	fmt.Fprint(w, "event: response.output_text.delta\n")
	fmt.Fprintf(w, "data: {\"type\":\"response.output_text.delta\",\"delta\":%q}\n\n", p.answerText())
	if p.mode == "timed" {
		p.awaitCompletion(w)
	}
}

func (p *e2eResponses) writeCompletion(w http.ResponseWriter, count int, metadata e2eMetadata) {
	fmt.Fprint(w, "event: response.completed\n")
	if p.mode == "no_status" && count > 1 {
		fmt.Fprintf(w, "data: {\"type\":\"response.completed\",\"response\":{\"model\":%q}}\n\n", metadata.model)
		p.signalCompletion(count)
		return
	}
	status := p.completionStatus(count)
	if metadata.modelPresent {
		fmt.Fprintf(w, "data: {\"type\":\"response.completed\",\"response\":{\"status\":%q,\"model\":%q}}\n\n", status, metadata.model)
		p.signalCompletion(count)
		return
	}
	fmt.Fprint(w, "data: {\"type\":\"response.completed\",\"response\":{\"status\":\"completed\"}}\n\n")
	p.signalCompletion(count)
}

func (p *e2eResponses) completionStatus(count int) string {
	if p.mode == "bad_status" && count > 1 {
		return "incomplete"
	}
	return "completed"
}

func (p *e2eResponses) answerText() string {
	if p.mode == "wrong_answer" {
		return "connected without expected marker"
	}
	return "connected MS01-PROBE-OK"
}

func (p *e2eResponses) awaitCompletion(w http.ResponseWriter) {
	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}
	close(p.deltaFlushed)
	<-p.releaseCompletion
}

func (p *e2eResponses) signalCompletion(count int) {
	if p.mode == "timed" && count > 1 {
		close(p.completionSent)
	}
}

func (p *e2eResponses) ReleaseCompletion() {
	if p.releaseCompletion != nil {
		p.releaseOnce.Do(func() { close(p.releaseCompletion) })
	}
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
