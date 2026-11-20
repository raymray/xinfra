package xrotate

import (
	"testing"
	"time"
)

func TestNewLumberjackMaxAgeInDays(t *testing.T) {
	lj := NewLumberjack("app.log", WithMaxAge(7*24*time.Hour))
	if lj.MaxAge != 7 {
		t.Fatalf("MaxAge = %d days, want 7", lj.MaxAge)
	}
}

func TestNewLumberjackKeepsMaxAgeWithRotationCount(t *testing.T) {
	lj := NewLumberjack("app.log", WithMaxAge(7*24*time.Hour), WithRotationCount(10))
	if lj.MaxAge != 7 || lj.MaxBackups != 10 {
		t.Fatalf("MaxAge=%d MaxBackups=%d, want 7 and 10", lj.MaxAge, lj.MaxBackups)
	}
}

func TestNewLumberjackSizeInMB(t *testing.T) {
	lj := NewLumberjack("app.log", WithRotationSize(200*1024*1024))
	if lj.MaxSize != 200 {
		t.Fatalf("MaxSize = %d MB, want 200", lj.MaxSize)
	}
}
