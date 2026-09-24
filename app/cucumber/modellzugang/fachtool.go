package modellzugang

import (
	"context"
	"encoding/json"
	"fmt"

	"agentcontrolplane/app/internal/app/modellpruefung"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

func (f *fachTool) Info(context.Context) (*schema.ToolInfo, error) {
	parameter := map[string]*schema.ParameterInfo{"projekt": {Type: schema.String, Desc: "Projektname", Required: true}}
	return &schema.ToolInfo{Name: "projektstatus_lesen", Desc: "Liest den freigegebenen Projektstatus.",
		ParamsOneOf: schema.NewParamsOneOfByParams(parameter)}, nil
}

func (f *fachTool) InvokableRun(ctx context.Context, raw string, _ ...tool.Option) (string, error) {
	if f.hold {
		f.entered <- struct{}{}
		<-ctx.Done()
		return "", ctx.Err()
	}

	return f.runStatus(ctx, raw)
}

func (f *fachTool) runStatus(ctx context.Context, raw string) (string, error) {
	var argument struct {
		Projekt string `json:"projekt"`
	}
	if err := json.Unmarshal([]byte(raw), &argument); err != nil {
		return "", err
	}

	status, err := f.pruefung.Rufe(ctx, f.agent, modellpruefung.StatusAktion, argument.Projekt)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("Projekt %s: %s; Prüfkennung %s", status.Projekt, status.Wert, status.Pruefkennung), nil
}
