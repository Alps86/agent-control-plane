package verbindungsstatus

import "context"

func (p codexPort) Read(ctx context.Context) (State, error) {
	status, err := p.service.Status(ctx)
	if err != nil {
		return State{}, err
	}

	state := State{Status: p.codexState(status.State)}
	state.Ready = state.Status == "connected"
	return state, nil
}

func (p codexPort) codexState(status string) string {
	if status == "connected" || status == "pending" || status == "reauthentication_required" || status == "unavailable" {
		return status
	}

	return "disconnected"
}

func (p openRouterPort) Read(ctx context.Context) (State, error) {
	status, err := p.service.Status(ctx)
	if err != nil {
		return State{}, err
	}

	if !status.Connected {
		return State{Status: "disconnected"}, nil
	}

	if status.Status == "einsatzbereit" {
		return State{Status: "connected", Ready: true}, nil
	}

	return State{Status: "not_ready"}, nil
}
