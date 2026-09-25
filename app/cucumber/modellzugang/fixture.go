package modellzugang

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"

	"agentcontrolplane/app/internal/adapter/model/codexabo"
	"agentcontrolplane/app/internal/app/modellpruefung"
)

func (f *statusFixture) Status(_ context.Context, projekt string) (modellpruefung.Status, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.accesses++
	status := f.status
	status.Projekt = projekt
	return status, nil
}

func (f *statusFixture) Accesses() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.accesses
}

func (c *credential) Access(context.Context) (string, string, error) {
	return c.token, c.account, nil
}

func (l *localTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if l.suite.mode == "stream_cancel" {
		return l.suite.openCancelStream(req), nil
	}

	response := httptest.NewRecorder()
	l.suite.serve(response, req)
	return response.Result(), nil
}

func (b *cancelBody) Read(dst []byte) (int, error) {
	b.mu.Lock()
	if b.offset < len(b.data) {
		read := copy(dst, b.data[b.offset:])
		b.offset += read
		b.mu.Unlock()
		return read, nil
	}

	b.mu.Unlock()
	<-b.done
	return 0, io.EOF
}

func (b *cancelBody) Close() error {
	b.once.Do(func() { close(b.done) })
	b.mu.Lock()
	b.closed = true
	b.mu.Unlock()
	return nil
}

func (b *cancelBody) Closed() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.closed
}

func (o *observationCollector) RecordModelResponse(entry codexabo.Observation) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.entries = append(o.entries, entry)
}

func (o *observationCollector) Entries() []codexabo.Observation {
	o.mu.Lock()
	defer o.mu.Unlock()
	return append([]codexabo.Observation(nil), o.entries...)
}
