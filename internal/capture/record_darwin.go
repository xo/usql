//go:build darwin

package capture

import (
	"bytes"
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

// OpenPTY opens a pseudo-terminal. It returns the leader side, which this
// process drives, and the follower side, which the child uses as its terminal.
//
// macOS opens the same /dev/ptmx that Linux does, but the three requests that
// follow are its own. Linux unlocks the pair and asks for its number, then
// builds the name as /dev/pts/N. macOS grants the follower to this user,
// unlocks it, and asks for the whole name, which looks like /dev/ttys003.
// There is no number to format, and no /dev/pts directory.
//
// The leader fd is put in non-blocking mode before os.NewFile wraps it. A
// blocking terminal fd does not reach the Go poller, and then SetReadDeadline
// accepts a deadline but never interrupts a read.
func OpenPTY() (*os.File, *os.File, error) {
	fd, err := syscall.Open("/dev/ptmx", syscall.O_RDWR|syscall.O_NOCTTY, 0)
	if err != nil {
		return nil, nil, fmt.Errorf("opening /dev/ptmx: %w", err)
	}
	for _, req := range []struct {
		name string
		code uintptr
	}{
		{"granting the pseudo-terminal", syscall.TIOCPTYGRANT},
		{"unlocking the pseudo-terminal", syscall.TIOCPTYUNLK},
	} {
		if _, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), req.code, 0); errno != 0 {
			_ = syscall.Close(fd)
			return nil, nil, fmt.Errorf("%s: %w", req.name, errno)
		}
	}
	// TIOCPTYGNAME writes the name of the follower into a buffer of 128
	// bytes, which is the size the request is defined for.
	var buf [128]byte
	if _, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd),
		syscall.TIOCPTYGNAME, uintptr(unsafe.Pointer(&buf[0]))); errno != 0 {
		_ = syscall.Close(fd)
		return nil, nil, fmt.Errorf("reading the pseudo-terminal name: %w", errno)
	}
	end := bytes.IndexByte(buf[:], 0)
	if end <= 0 {
		_ = syscall.Close(fd)
		return nil, nil, fmt.Errorf("the pseudo-terminal name is not a string: %q", buf[:])
	}
	name := string(buf[:end])
	if err := syscall.SetNonblock(fd, true); err != nil {
		_ = syscall.Close(fd)
		return nil, nil, fmt.Errorf("setting non-blocking mode: %w", err)
	}
	leader := os.NewFile(uintptr(fd), "/dev/ptmx")
	follower, err := os.OpenFile(name, os.O_RDWR|syscall.O_NOCTTY, 0)
	if err != nil {
		_ = leader.Close()
		return nil, nil, fmt.Errorf("opening %s: %w", name, err)
	}
	return leader, follower, nil
}
