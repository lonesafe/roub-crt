// +build windows

package terminal

import "os"

// resizeSignals is empty on Windows because there is no equivalent of
// SIGWINCH on this platform. Terminal resizing must be handled differently
// (e.g. via the console API or fyne's window resize events).
var resizeSignals = []os.Signal{}
