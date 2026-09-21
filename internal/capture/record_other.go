//go:build !linux && !darwin

package capture

import (
	"context"
	"fmt"
)

// Record reports that this platform cannot record a session yet.
//
// Linux and macOS both record. Windows has no pseudo-terminal device file, so
// it needs a pseudo console, which it creates with CreatePseudoConsole.
func Record(_ context.Context, _ string, _ Session) (*Transcript, error) {
	return nil, fmt.Errorf("recording a session on this platform: %w", ErrUnsupported)
}
