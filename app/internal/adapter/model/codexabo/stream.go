package codexabo

import (
	"bufio"
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/cloudwego/eino/schema"
)

func (a *Adapter) openStream(ctx context.Context, resp *http.Response) *schema.StreamReader[*schema.Message] {
	streamCtx, cancel := context.WithCancel(ctx)
	go func() { <-streamCtx.Done(); resp.Body.Close() }()
	reader, writer := schema.Pipe[*schema.Message](8)
	events := make(chan streamEvent, 8)
	id := a.safeID(resp.Header.Get("x-request-id"))
	model := a.safeID(resp.Header.Get("OpenAI-Model"))
	parser := &eventParser{ctx: streamCtx, out: events, observer: a.Observer, requestID: id, headerModel: model}
	go func() { defer close(events); parser.parse(resp.Body) }()
	pump := &streamPump{ctx: streamCtx, cancel: cancel, body: resp.Body, writer: writer, events: events}
	go pump.run()
	return schema.StreamReaderWithConvert(reader, func(m *schema.Message) (*schema.Message, error) {
		if m == nil {
			return nil, schema.ErrNoValue
		}

		return m, nil
	})
}

func (p *streamPump) run() {
	defer p.cancel()
	defer p.body.Close()
	defer p.writer.Close()
	p.tick = time.NewTicker(100 * time.Millisecond)
	defer p.tick.Stop()
	for {
		if !p.next() {
			return
		}
	}
}

func (p *streamPump) next() bool {
	select {
	case <-p.ctx.Done():
		p.writer.Send(nil, p.ctx.Err())
		return false
	case e, ok := <-p.events:
		if !ok {
			return false
		}

		return !p.writer.Send(e.msg, e.err) && e.err == nil
	case <-p.tick.C:
		return !p.writer.Send(nil, nil)
	}
}

func (p *eventParser) parse(r io.Reader) {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 4096), 4<<20)
	for scanner.Scan() {
		if p.ctx.Err() != nil {
			p.aborted()
			return
		}

		if !p.acceptLine(scanner.Text()) {
			return
		}
	}

	if p.ctx.Err() != nil {
		p.aborted()
		return
	}

	p.finish(scanner.Err())
}

func (p *eventParser) finish(readErr error) {
	p.flush()
	if p.completed || p.failed {
		return
	}

	if readErr != nil {
		p.observe(Observation{RequestID: p.requestID, Model: p.headerModel, Status: "stream_error"})
		p.send(nil, errors.New("subscription stream read failed"))
		return
	}

	p.observe(Observation{RequestID: p.requestID, Model: p.headerModel, Status: "stream_truncated"})
	p.send(nil, errors.New("subscription stream ended before completion"))
}

func (p *eventParser) acceptLine(line string) bool {
	if line == "" {
		p.flush()
		p.event = ""
		p.data = nil
		return !p.failed
	}

	if strings.HasPrefix(line, "event:") {
		p.event = strings.TrimSpace(strings.TrimPrefix(line, "event:"))
	}

	if strings.HasPrefix(line, "data:") {
		p.data = append(p.data, strings.TrimSpace(strings.TrimPrefix(line, "data:")))
	}

	return true
}
