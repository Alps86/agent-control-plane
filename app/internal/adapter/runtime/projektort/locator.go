package projektort

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

// NewLocator bindet ausschließlich den vertrauten Datenbankpfad.
func NewLocator(databasePath string) *Locator {
	return &Locator{databasePath: databasePath}
}

// Ensure erstellt den privaten Projektort ohne Symlink-Verfolgung.
func (l *Locator) Ensure(ctx context.Context, organizationID, projectID string) (string, error) {
	if !l.validID(organizationID) || !l.validID(projectID) {
		return "", ErrIntegrity
	}

	parent, _, err := l.parent(ctx)
	if err != nil {
		return "", err
	}

	defer syscall.Close(parent)
	if err := l.ensureBelow(ctx, parent, organizationID, projectID); err != nil {
		return "", err
	}

	return l.Resolve(ctx, organizationID, projectID)
}

// Resolve verifiziert Ort und gesamten Inhalt vor jeder Pfadübergabe.
func (l *Locator) Resolve(ctx context.Context, organizationID, projectID string) (string, error) {
	fd, path, err := l.openVerified(ctx, organizationID, projectID)
	if err != nil {
		return "", err
	}

	defer syscall.Close(fd)
	return path, nil
}

func (l *Locator) openVerified(ctx context.Context, organizationID, projectID string) (int, string, error) {
	if !l.validID(organizationID) || !l.validID(projectID) {
		return -1, "", ErrIntegrity
	}

	parent, path, err := l.parent(ctx)
	if err != nil {
		return -1, "", err
	}

	defer syscall.Close(parent)
	fd, err := l.openBelow(ctx, parent, organizationID, projectID)
	if err != nil {
		return -1, "", err
	}

	if err := l.scan(ctx, fd, 0, new(int)); err != nil {
		syscall.Close(fd)
		return -1, "", err
	}

	return fd, filepath.Join(path, "project-workspaces", organizationID, projectID), nil
}

func (l *Locator) validID(id string) bool {
	return id != "" && id != "." && id != ".." && len(id) <= 255 &&
		!strings.ContainsAny(id, "/\\\x00") && filepath.Base(id) == id
}

func (l *Locator) parent(ctx context.Context) (int, string, error) {
	if l == nil || l.databasePath == "" {
		return -1, "", ErrIntegrity
	}

	absolute, err := filepath.Abs(l.databasePath)
	if err != nil {
		return -1, "", fmt.Errorf("Datenbankpfad: %w", err)
	}

	path := filepath.Dir(absolute)
	fd, err := l.openAbsolute(ctx, path)
	return fd, path, err
}

func (l *Locator) openAbsolute(ctx context.Context, path string) (int, error) {
	fd, err := syscall.Open("/", syscall.O_RDONLY|syscall.O_DIRECTORY|syscall.O_NOFOLLOW|syscall.O_CLOEXEC, 0)
	if err != nil {
		return -1, ErrIntegrity
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

func (l *Locator) ensureBelow(ctx context.Context, parent int, organizationID, projectID string) error {
	fd := parent
	for _, segment := range []string{"project-workspaces", organizationID, projectID} {
		next, err := l.step(ctx, fd, segment, true)
		if fd != parent {
			syscall.Close(fd)
		}

		if err != nil {
			return err
		}

		fd = next
		if !l.privateDirectory(fd) {
			syscall.Close(fd)
			return ErrIntegrity
		}
	}

	return syscall.Close(fd)
}

func (l *Locator) openBelow(ctx context.Context, parent int, organizationID, projectID string) (int, error) {
	fd := parent
	for _, segment := range []string{"project-workspaces", organizationID, projectID} {
		next, err := l.step(ctx, fd, segment, false)
		if fd != parent {
			syscall.Close(fd)
		}

		if err != nil {
			return -1, err
		}

		fd = next
		if !l.privateDirectory(fd) {
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
			return -1, ErrIntegrity
		}
	}

	fd, err := syscall.Openat(parent, segment, syscall.O_RDONLY|syscall.O_DIRECTORY|syscall.O_NOFOLLOW|syscall.O_CLOEXEC, 0)
	if err != nil {
		return -1, ErrIntegrity
	}

	return fd, nil
}

func (l *Locator) privateDirectory(fd int) bool {
	var stat syscall.Stat_t
	if syscall.Fstat(fd, &stat) != nil {
		return false
	}

	return stat.Uid == uint32(os.Geteuid()) && stat.Mode&syscall.S_IFMT == syscall.S_IFDIR && stat.Mode&0777 == 0700
}
