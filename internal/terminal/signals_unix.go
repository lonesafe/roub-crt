// +build !windows

package terminal

import (
	"os"
	"syscall"
)

// resizeSignals lists the OS signals that indicate a terminal resize event on
// Unix-like systems. SIGWINCH is delivered by the kernel when the controlling
// terminal is resized.
var resizeSignals = []os.Signal{syscall.SIGWINCH}
