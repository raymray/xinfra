package xkit

import (
	"context"
	"encoding/json"
	"os/exec"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestBiMapRoundTrip(t *testing.T) {
	m := NewBiMapKV(map[string]int{"a": 1})
	m.AddKV("b", 2)
	if v, ok := m.K2V("b"); !ok || v != 2 {
		t.Fatalf("K2V(b) = %v %v", v, ok)
	}
	if k, ok := m.V2K(1); !ok || k != "a" {
		t.Fatalf("V2K(1) = %v %v", k, ok)
	}
	if !m.KExist("a") || !m.VExist(2) || len(m.Ks()) != 2 || len(m.Vs()) != 2 {
		t.Fatal("exist/keys/values wrong")
	}
}

func TestBiMapMissingKey(t *testing.T) {
	m := NewBiMap[string, int]()
	if _, ok := m.K2V("x"); ok {
		t.Fatal("missing key reported present")
	}
}

func TestMChannelWriteAfterEndReturnsFalse(t *testing.T) {
	ch := NewMChannel[int](1)
	ch.End()
	if ch.Write(1) || ch.WriteCtx(context.Background(), 1) || ch.WriteAll([]int{1, 2}) != 0 {
		t.Fatal("write after End should fail")
	}
}

func TestMChannelReadDrainsThenDone(t *testing.T) {
	ch := NewMChannel[int](2)
	ch.Write(1)
	ch.Write(2)
	ch.End()
	for _, want := range []int{1, 2} {
		if v, done := ch.Read(); done || v != want {
			t.Fatalf("Read = %d %v, want %d false", v, done, want)
		}
	}
	if _, done := ch.Read(); !done {
		t.Fatal("Read after drain should report done")
	}
}

func TestMChannelCtxCancel(t *testing.T) {
	ch := NewMChannel[int](0)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if ch.WriteCtx(ctx, 1) {
		t.Fatal("WriteCtx on cancelled ctx should fail")
	}
	if _, done := ch.ReadCtx(ctx); done {
		t.Fatal("ReadCtx on cancelled ctx should not report done")
	}
}

func TestMChannelConcurrentEndIsSafe(t *testing.T) {
	ch := NewMChannel[int](10)
	var wg sync.WaitGroup
	for range 20 {
		wg.Add(2)
		go func() { defer wg.Done(); ch.End() }()
		go func() { defer wg.Done(); ch.Write(1) }()
	}
	wg.Wait()
	if !ch.IsEnded() {
		t.Fatal("not ended")
	}
}

func TestMChannelMultiReadCtxBatches(t *testing.T) {
	ch := NewMChannel[int](10)
	ch.WriteAll([]int{1, 2, 3})
	rows, done := ch.MultiReadCtx(context.Background(), 2, time.Second)
	if len(rows) != 2 || done {
		t.Fatalf("rows=%v done=%v", rows, done)
	}
	rows, done = ch.MultiReadCtx(context.Background(), 5, 50*time.Millisecond)
	if len(rows) != 1 || done {
		t.Fatalf("timeout batch rows=%v done=%v", rows, done)
	}
}

func TestToTimeFormats(t *testing.T) {
	cases := map[string]any{
		"date":     "2026-11-03",
		"datetime": "2026-11-03 10:20:30",
		"seconds":  int64(1793700000),
		"millis":   int64(1793700000123),
	}
	for name, in := range cases {
		t.Run(name, func(t *testing.T) {
			got, err := ToTime(in)
			if err != nil || got.Year() < 2026 {
				t.Fatalf("ToTime(%v) = %v, %v", in, got, err)
			}
		})
	}
}

func TestToTimeMillisPrecision(t *testing.T) {
	got, _ := ToTime(int64(1793700000123))
	if got.UnixMilli() != 1793700000123 {
		t.Fatalf("UnixMilli = %d", got.UnixMilli())
	}
}

func TestToTimeInvalidReturnsError(t *testing.T) {
	if _, err := ToTime("not a time"); err == nil {
		t.Fatal("want error")
	}
	if _, err := ToTime(struct{}{}); err == nil {
		t.Fatal("want error for unsupported type")
	}
}

func TestRootPackageHasNoThirdPartyDeps(t *testing.T) {
	out, err := exec.Command("go", "list", "-deps", "-f", "{{if not .Standard}}{{.ImportPath}}{{end}}", ".").Output()
	if err != nil {
		t.Fatalf("go list: %v", err)
	}
	for _, p := range strings.Fields(string(out)) {
		if p != "github.com/raymray/xinfra/xkit" {
			t.Errorf("root package depends on %s", p)
		}
	}
}

func TestToTimeNumericTypes(t *testing.T) {
	for _, in := range []any{uint32(1793700000), int32(1793700000), float64(1793700000), json.Number("1793700000")} {
		got, err := ToTime(in)
		if err != nil || got.Unix() != 1793700000 {
			t.Fatalf("ToTime(%T %v) = %v, %v", in, in, got, err)
		}
	}
}

func TestBiMapMustPanicsOnMissing(t *testing.T) {
	m := NewBiMap[string, int]()
	for name, fn := range map[string]func(){"MustK2V": func() { m.MustK2V("x") }, "MustV2K": func() { m.MustV2K(1) }} {
		t.Run(name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatal("want panic on missing entry")
				}
			}()
			fn()
		})
	}
}

func TestBiMapRemapKeepsBijection(t *testing.T) {
	m := NewBiMap[string, int]()
	m.AddKV("a", 1)
	m.AddKV("a", 2) // a 改映射到 2，旧值 1 不应再反查到 a
	if m.VExist(1) {
		t.Fatal("stale reverse entry 1→a remains")
	}
	m.AddKV("b", 2) // 值 2 改由 b 持有，a 不应再存在
	if m.KExist("a") || len(m.Ks()) != len(m.Vs()) {
		t.Fatalf("bijection broken: keys=%v values=%v", m.Ks(), m.Vs())
	}
}

func TestMChannelBlockedWriteReturnsFalseAfterEnd(t *testing.T) {
	ch := NewMChannel[int](1)
	ch.Write(1) // 填满
	res := make(chan bool, 1)
	go func() { res <- ch.Write(2) }()
	time.Sleep(50 * time.Millisecond)
	ch.End()
	select {
	case ok := <-res:
		if ok {
			t.Fatal("blocked write should report false after End")
		}
	case <-time.After(time.Second):
		t.Fatal("blocked write did not return after End")
	}
	if v, done := ch.Read(); done || v != 1 {
		t.Fatalf("buffered item lost: %d %v", v, done)
	}
}
