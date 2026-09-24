package codexabo

import (
	"context"
	"errors"
	"io"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

func (a *Adapter) Generate(ctx context.Context, in []*schema.Message, opts ...model.Option) (*schema.Message, error) {
	sr, err := a.Stream(ctx, in, opts...)
	if err != nil {
		return nil, err
	}

	defer sr.Close()
	result := &schema.Message{Role: schema.Assistant}
	for {
		chunk, e := sr.Recv()
		if errors.Is(e, io.EOF) {
			return result, nil
		}

		if e != nil {
			return nil, e
		}

		a.mergeChunk(result, chunk)
	}
}

func (a *Adapter) mergeChunk(result, chunk *schema.Message) {
	if chunk == nil {
		return
	}

	result.Content += chunk.Content
	result.ToolCalls = append(result.ToolCalls, chunk.ToolCalls...)
	if chunk.ResponseMeta != nil {
		result.ResponseMeta = chunk.ResponseMeta
	}

	if chunk.Extra != nil {
		result.Extra = chunk.Extra
	}
}

func (a *Adapter) Stream(ctx context.Context, in []*schema.Message, opts ...model.Option) (*schema.StreamReader[*schema.Message], error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	req, err := a.makeRequest(ctx, in, opts...)
	if err != nil {
		return nil, err
	}

	resp, err := a.send(req)
	if err != nil {
		return nil, err
	}

	return a.openStream(ctx, resp), nil
}
