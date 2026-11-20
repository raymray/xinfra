//go:build !windows

package xlifecycle

import (
	"os"
	"syscall"
)

var exitSignals = []os.Signal{syscall.SIGINT, syscall.SIGTERM}
