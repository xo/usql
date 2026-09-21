//go:build linux

package capture

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

// OpenPTY opens a pseudo-terminal. It returns the leader side, which this
// process drives, and the follower side, which the child uses as its terminal.
//
// The leader fd is put in non-blocking mode before os.NewFile wraps it. A
// blocking terminal fd does not reach the Go poller, and then SetReadDeadline
// accepts a deadline but never interrupts a read.
func OpenPTY() (*os.File, *os.File, error) {
	fd, err := syscall.Open("/dev/ptmx", syscall.O_RDWR|syscall.O_NOCTTY, 0)
	if err != nil {
		return nil, nil, fmt.Errorf("opening /dev/ptmx: %w", err)
	}
	var unlock int32
	if _, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd),
		syscall.TIOCSPTLCK, uintptr(unsafe.Pointer(&unlock))); errno != 0 {
		_ = syscall.Close(fd)
		return nil, nil, fmt.Errorf("unlocking pseudo-terminal: %w", errno)
	}
	var num uint32
	if _, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd),
		syscall.TIOCGPTN, uintptr(unsafe.Pointer(&num))); errno != 0 {
		_ = syscall.Close(fd)
		return nil, nil, fmt.Errorf("reading pseudo-terminal number: %w", errno)
	}
	if err := syscall.SetNonblock(fd, true); err != nil {
		_ = syscall.Close(fd)
		return nil, nil, fmt.Errorf("setting non-blocking mode: %w", err)
	}
	leader := os.NewFile(uintptr(fd), "/dev/ptmx")
	name := fmt.Sprintf("/dev/pts/%d", num)
	follower, err := os.OpenFile(name, os.O_RDWR|syscall.O_NOCTTY, 0)
	if err != nil {
		_ = leader.Close()
		return nil, nil, fmt.Errorf("opening %s: %w", name, err)
	}
	return leader, follower, nil
}
