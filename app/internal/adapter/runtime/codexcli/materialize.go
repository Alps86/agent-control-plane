package codexcli

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

func (c Config) materialize(scratch string) error {
	for _, binary := range [][3]string{{c.CLIPath, c.CLISHA256, "codex"},
		{c.BwrapPath, c.BwrapSHA256, "bwrap"},
		{c.ActionServerPath, c.ActionServerSHA256, "action-server"}} {
		if err := c.copyPinned(binary[0], binary[1], filepath.Join(scratch, binary[2])); err != nil {
			return err
		}
	}

	return nil
}

func (c Config) copyPinned(path, expected, destination string) error {
	fd, err := syscall.Open(path, syscall.O_RDONLY|syscall.O_NOFOLLOW|syscall.O_CLOEXEC, 0)
	if err != nil {
		return err
	}

	source := os.NewFile(uintptr(fd), "pinned-binary")
	defer source.Close()
	if info, err := source.Stat(); err != nil || !info.Mode().IsRegular() {
		return errors.New("pinned binary is not regular")
	}

	return c.copyAndVerify(source, expected, destination)
}

func (c Config) copyAndVerify(source *os.File, expected, destination string) error {
	target, err := os.OpenFile(destination, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0700)
	if err != nil {
		return err
	}

	defer target.Close()
	hash := sha256.New()
	if _, err := io.Copy(io.MultiWriter(target, hash), source); err != nil {
		return err
	}

	return c.finishCopy(target, hash.Sum(nil), expected)
}

func (c Config) finishCopy(target *os.File, digest []byte, expected string) error {
	if !strings.EqualFold(hex.EncodeToString(digest), expected) {
		return errors.New("pinned binary digest mismatch")
	}

	if err := target.Sync(); err != nil {
		return err
	}

	return target.Chmod(0500)
}
