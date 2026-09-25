package codexprofil

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"os"
	"path/filepath"
	"time"

	appprofil "agentcontrolplane/app/internal/app/codexprofil"
	portprofil "agentcontrolplane/app/internal/port/codexprofil"
)

// NewBrokerFactory bindet die Host-Seite des isolierten MCP-Sockets.
func NewBrokerFactory(service *appprofil.Service) *Factory {
	return &Factory{service: service}
}

// Start erstellt eine neue, an Organisation und Agent gebundene Session.
func (f *Factory) Start(ctx context.Context, organizationID, agentID string) (portprofil.Session, error) {
	if f == nil || f.service == nil {
		return nil, ErrIntegrity
	}

	b := &Broker{service: f.service, organizationID: organizationID, agentID: agentID}
	if err := b.start(ctx); err != nil {
		return nil, err
	}

	return b, nil
}

func (b *Broker) start(ctx context.Context) error {
	if b == nil || b.service == nil || b.listener != nil {
		return ErrIntegrity
	}

	if err := b.service.Authorize(ctx, b.organizationID, b.agentID, appprofil.ActionRuntimeProbe); err != nil {
		return err
	}

	directory, err := os.MkdirTemp("", "acp-codex-action-")
	if err != nil {
		return err
	}

	return b.listen(ctx, directory)
}

func (b *Broker) listen(ctx context.Context, directory string) error {
	path := filepath.Join(directory, "action.sock")
	listener, err := net.Listen("unix", path)
	if err != nil {
		os.RemoveAll(directory)
		return err
	}

	if err := os.Chmod(path, 0600); err != nil {
		listener.Close()
		os.RemoveAll(directory)
		return err
	}

	b.listener, b.directory, b.socketPath = listener, directory, path
	go b.serve(ctx, listener)
	return nil
}

// SocketPath wird nur der vertrauenswürdigen bwrap-Mountkonfiguration übergeben.
func (b *Broker) SocketPath() string {
	if b == nil {
		return ""
	}

	return b.socketPath
}

// Close beendet die Scope-Grenze und entfernt nur ihren privaten Socket.
func (b *Broker) Close() error {
	if b == nil || b.listener == nil {
		return nil
	}

	err := b.listener.Close()
	if errors.Is(err, net.ErrClosed) {
		err = nil
	}
	b.listener = nil
	if removeErr := os.RemoveAll(b.directory); err == nil {
		err = removeErr
	}

	return err
}

func (b *Broker) serve(ctx context.Context, listener net.Listener) {
	stop := context.AfterFunc(ctx, func() { listener.Close() })
	defer stop()
	for {
		connection, err := listener.Accept()
		if err != nil {
			return
		}

		b.handle(ctx, connection)
	}
}

func (b *Broker) handle(ctx context.Context, connection net.Conn) {
	defer connection.Close()
	connection.SetDeadline(time.Now().Add(10 * time.Second))
	request, err := b.read(connection)
	if err != nil || request.ActionID != appprofil.ActionMarkdownSave {
		json.NewEncoder(connection).Encode(portprofil.ActionResponse{Code: "invalid_action"})
		return
	}

	err = b.service.SaveProof(ctx, b.organizationID, b.agentID, request.Markdown)
	if err != nil {
		json.NewEncoder(connection).Encode(portprofil.ActionResponse{Code: b.denialCode(err)})
		return
	}

	json.NewEncoder(connection).Encode(portprofil.ActionResponse{Allowed: true, Code: "artifact_saved"})
}

func (b *Broker) denialCode(err error) string {
	if errors.Is(err, appprofil.ErrWriteDisabled) {
		return "write_denied"
	}

	if errors.Is(err, ErrIntegrity) {
		return "workspace_integrity_denied"
	}

	return "policy_denied"
}

func (b *Broker) read(connection net.Conn) (portprofil.ActionRequest, error) {
	var request portprofil.ActionRequest
	decoder := json.NewDecoder(io.LimitReader(connection, 1<<17))
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&request)
	return request, err
}
