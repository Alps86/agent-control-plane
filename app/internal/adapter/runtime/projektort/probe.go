package projektort

import (
	"context"
	"syscall"
)

// NewStartProbe bindet die enge Adapterprobe an denselben privaten Locator.
func NewStartProbe(locator *Locator) *StartProbe {
	return &StartProbe{locator: locator}
}

// Probe konsumiert den Pfad nur, wenn er weiterhin dem geprüften Projektort entspricht.
func (p *StartProbe) Probe(ctx context.Context, organizationID, projectID, verifiedPath string) error {
	if p == nil || p.locator == nil || verifiedPath == "" {
		return ErrIntegrity
	}

	fd, path, err := p.locator.openVerified(ctx, organizationID, projectID)
	if err != nil || path != verifiedPath {
		if fd >= 0 {
			syscall.Close(fd)
		}

		return ErrIntegrity
	}

	defer syscall.Close(fd)
	if !p.locator.privateDirectory(fd) {
		return ErrIntegrity
	}

	return nil
}
