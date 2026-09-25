package codexprofil

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

// NewLocator bindet den Arbeitsbereich ausschließlich an den Datenbankpfad.
func NewLocator(databasePath string) *Locator {
	return &Locator{databasePath: databasePath}
}

// Ensure erstellt den privaten Workspace ohne Symlink-Verfolgung.
func (l *Locator) Ensure(ctx context.Context, organizationID, agentID string) error {
	if !l.validID(organizationID) || !l.validID(agentID) {
		return ErrIntegrity
	}

	parent, err := l.parent(ctx)
	if err != nil {
		return err
	}

	defer syscall.Close(parent)
	return l.ensureBelow(ctx, parent, organizationID, agentID)
}

// Verify öffnet den bereits angelegten Bereich erneut ohne Symlink-Verfolgung.
func (l *Locator) Verify(ctx context.Context, organizationID, agentID string) error {
	fd, err := l.openWorkspace(ctx, organizationID, agentID)
	if err != nil {
		return err
	}

	defer syscall.Close(fd)
	return l.bounded(fd)
}

// WriteProof schreibt nur das fest benannte, private Markdown-Artefakt.
func (l *Locator) WriteProof(ctx context.Context, organizationID, agentID, markdown string) error {
	if len(markdown) == 0 || len(markdown) > 65536 {
		return ErrIntegrity
	}

	fd, err := l.openWorkspace(ctx, organizationID, agentID)
	if err != nil {
		return err
	}

	defer syscall.Close(fd)
	if err := l.bounded(fd); err != nil {
		return err
	}

	return l.writeProof(ctx, fd, markdown)
}

func (l *Locator) validID(id string) bool {
	return id != "" && id != "." && id != ".." && len(id) <= 255 &&
		!strings.ContainsAny(id, "/\\\x00") && filepath.Base(id) == id
}

func (l *Locator) parent(ctx context.Context) (int, error) {
	if l == nil || l.databasePath == "" {
		return -1, ErrIntegrity
	}

	absolute, err := filepath.Abs(l.databasePath)
	if err != nil {
		return -1, fmt.Errorf("Datenbankpfad: %w", err)
	}

	return l.openAbsolute(ctx, filepath.Dir(absolute))
}

func (l *Locator) openAbsolute(ctx context.Context, path string) (int, error) {
	fd, err := syscall.Open("/", syscall.O_RDONLY|syscall.O_DIRECTORY|syscall.O_NOFOLLOW|syscall.O_CLOEXEC, 0)
	if err != nil {
		return -1, fmt.Errorf("Arbeitsbereichwurzel: %w", err)
	}

	for _, segment := range strings.Split(strings.TrimPrefix(filepath.Clean(path), "/"), "/") {
		if segment == "" {
			continue
		}

		next, stepErr := l.step(ctx, fd, segment, false)
		syscall.Close(fd)
		fd, err = next, stepErr
		if err != nil {
			return -1, err
		}
	}

	return fd, nil
}

func (l *Locator) ensureBelow(ctx context.Context, parent int, organizationID, agentID string) error {
	fd := parent
	for _, segment := range []string{"codex-workspaces", organizationID, agentID} {
		next, err := l.step(ctx, fd, segment, true)
		if fd != parent {
			syscall.Close(fd)
		}

		if err != nil {
			return err
		}

		fd = next
	}

	defer syscall.Close(fd)
	return l.bounded(fd)
}

func (l *Locator) openWorkspace(ctx context.Context, organizationID, agentID string) (int, error) {
	if !l.validID(organizationID) || !l.validID(agentID) {
		return -1, ErrIntegrity
	}

	parent, err := l.parent(ctx)
	if err != nil {
		return -1, err
	}

	defer syscall.Close(parent)
	return l.openBelow(ctx, parent, organizationID, agentID)
}

func (l *Locator) openBelow(ctx context.Context, parent int, organizationID, agentID string) (int, error) {
	fd := parent
	for _, segment := range []string{"codex-workspaces", organizationID, agentID} {
		next, err := l.step(ctx, fd, segment, false)
		if fd != parent {
			syscall.Close(fd)
		}

		if err != nil {
			return -1, err
		}

		fd = next
		if !l.private(fd) {
			syscall.Close(fd)
			return -1, ErrIntegrity
		}
	}

	return fd, nil
}

func (l *Locator) step(ctx context.Context, parent int, segment string, create bool) (int, error) {
	if err := ctx.Err(); err != nil {
		return -1, err
	}

	if create {
		err := syscall.Mkdirat(parent, segment, 0700)
		if err != nil && !errors.Is(err, syscall.EEXIST) {
			return -1, fmt.Errorf("Arbeitsbereich erstellen: %w", ErrIntegrity)
		}
	}

	fd, err := syscall.Openat(parent, segment, syscall.O_RDONLY|syscall.O_DIRECTORY|syscall.O_NOFOLLOW|syscall.O_CLOEXEC, 0)
	if err != nil {
		return -1, fmt.Errorf("Arbeitsbereich öffnen: %w", ErrIntegrity)
	}

	if create && !l.private(fd) {
		syscall.Close(fd)
		return -1, ErrIntegrity
	}

	return fd, nil
}

func (l *Locator) private(fd int) bool {
	var stat syscall.Stat_t
	if syscall.Fstat(fd, &stat) != nil {
		return false
	}

	return stat.Uid == uint32(os.Geteuid()) && stat.Mode&0777 == 0700
}

func (l *Locator) bounded(fd int) error {
	copyFD, err := syscall.Dup(fd)
	if err != nil {
		return ErrIntegrity
	}

	file := os.NewFile(uintptr(copyFD), "workspace")
	defer file.Close()
	names, err := file.Readdirnames(-1)
	if err != nil && !errors.Is(err, io.EOF) {
		return ErrIntegrity
	}

	for _, name := range names {
		if name != "runtime-proof.md" || !l.proofFile(fd) {
			return ErrIntegrity
		}
	}

	return nil
}

func (l *Locator) proofFile(parent int) bool {
	fd, err := syscall.Openat(parent, "runtime-proof.md", syscall.O_RDONLY|syscall.O_NOFOLLOW|syscall.O_CLOEXEC, 0)
	if err != nil {
		return false
	}

	defer syscall.Close(fd)
	var stat syscall.Stat_t
	if syscall.Fstat(fd, &stat) != nil {
		return false
	}

	return stat.Mode&syscall.S_IFMT == syscall.S_IFREG && stat.Mode&0777 == 0600 && stat.Uid == uint32(os.Geteuid())
}

func (l *Locator) writeProof(ctx context.Context, parent int, markdown string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	fd, err := syscall.Openat(parent, "runtime-proof.md", syscall.O_WRONLY|syscall.O_CREAT|syscall.O_EXCL|syscall.O_NOFOLLOW|syscall.O_CLOEXEC, 0600)
	if errors.Is(err, syscall.EEXIST) {
		return l.sameProof(parent, markdown)
	}

	if err != nil {
		return ErrIntegrity
	}

	file := os.NewFile(uintptr(fd), "runtime-proof.md")
	_, err = file.WriteString(markdown)
	if err == nil {
		err = file.Sync()
	}

	closeErr := file.Close()
	if err != nil || closeErr != nil {
		syscall.Unlinkat(parent, "runtime-proof.md")
		return ErrIntegrity
	}

	return nil
}

func (l *Locator) sameProof(parent int, markdown string) error {
	fd, err := syscall.Openat(parent, "runtime-proof.md", syscall.O_RDONLY|syscall.O_NOFOLLOW|syscall.O_CLOEXEC, 0)
	if err != nil {
		return ErrIntegrity
	}

	file := os.NewFile(uintptr(fd), "runtime-proof.md")
	defer file.Close()
	var stat syscall.Stat_t
	if syscall.Fstat(fd, &stat) != nil || stat.Mode&syscall.S_IFMT != syscall.S_IFREG ||
		stat.Mode&0777 != 0600 || stat.Uid != uint32(os.Geteuid()) {
		return ErrIntegrity
	}

	content, err := io.ReadAll(io.LimitReader(file, 65537))
	if err != nil || !bytes.Equal(content, []byte(markdown)) {
		return ErrIntegrity
	}

	return nil
}
