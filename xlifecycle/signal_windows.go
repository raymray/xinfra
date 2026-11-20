//go:build windows

package xlifecycle

import (
	"os"
	"syscall"
)

var exitSignals = []os.Signal{os.Interrupt, syscall.SIGTERM}
