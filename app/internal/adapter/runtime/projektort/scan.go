package projektort

import (
	"context"
	"io"
	"os"
	"syscall"
)

func (l *Locator) scan(ctx context.Context, parent, depth int, count *int) error {
	if depth > 32 || *count > 10000 || ctx.Err() != nil {
		return ErrIntegrity
	}

	copyFD, err := syscall.Dup(parent)
	if err != nil {
		return ErrIntegrity
	}

	file := os.NewFile(uintptr(copyFD), "project-workspace")
	defer file.Close()
	names, err := file.Readdirnames(-1)
	if err != nil && err != io.EOF {
		return ErrIntegrity
	}

	for _, name := range names {
		*count = *count + 1
		if *count > 10000 || l.scanEntry(ctx, parent, name, depth, count) != nil {
			return ErrIntegrity
		}
	}

	return nil
}

func (l *Locator) scanEntry(ctx context.Context, parent int, name string, depth int, count *int) error {
	fd, err := syscall.Openat(parent, name, syscall.O_RDONLY|syscall.O_NOFOLLOW|syscall.O_CLOEXEC|syscall.O_NONBLOCK, 0)
	if err != nil {
		return ErrIntegrity
	}

	defer syscall.Close(fd)
	var stat syscall.Stat_t
	if syscall.Fstat(fd, &stat) != nil || stat.Uid != uint32(os.Geteuid()) {
		return ErrIntegrity
	}

	if stat.Mode&syscall.S_IFMT == syscall.S_IFDIR && stat.Mode&0777 == 0700 {
		return l.scan(ctx, fd, depth+1, count)
	}

	if stat.Mode&syscall.S_IFMT != syscall.S_IFREG || stat.Mode&0077 != 0 || stat.Nlink != 1 {
		return ErrIntegrity
	}

	return nil
}
