package fskit

import (
	"path/filepath"
	"testing"
)

func TestCSVRoundTrip(t *testing.T) {
	p := filepath.Join(t.TempDir(), "sub", "a.csv")
	data := []map[string]any{{"name": "a", "n": 1}, {"name": "b", "n": 2}}
	if err := WriteCSV(data, p, "name", "n"); err != nil {
		t.Fatalf("WriteCSV: %v", err)
	}
	headers, rows, err := ReadCSVAsMaps(p)
	if err != nil {
		t.Fatalf("ReadCSVAsMaps: %v", err)
	}
	if len(headers) != 2 || headers[0] != "name" || len(rows) != 2 || rows[1]["name"] != "b" {
		t.Fatalf("headers=%v rows=%v", headers, rows)
	}
}

func TestExcelRoundTrip(t *testing.T) {
	p := filepath.Join(t.TempDir(), "a.xlsx")
	data := []map[string]any{{"name": "a", "n": 1}, {"name": "b", "n": 2}}
	if err := WriteExcel(data, p, "S1", "name", "n"); err != nil {
		t.Fatalf("WriteExcel: %v", err)
	}
	rows, err := ReadExcelColumns(p, "S1", true, "name")
	if err != nil {
		t.Fatalf("ReadExcelColumns: %v", err)
	}
	if len(rows) != 2 || rows[0][0] != "a" || rows[1][0] != "b" {
		t.Fatalf("rows=%v", rows)
	}
}

func TestJSONRoundTrip(t *testing.T) {
	type item struct{ Name string }
	p := filepath.Join(t.TempDir(), "a.json")
	if err := WriteJSON([]item{{"x"}}, p, true); err != nil {
		t.Fatalf("WriteJSON: %v", err)
	}
	got, err := ReadJSON[[]item](p)
	if err != nil || len(got) != 1 || got[0].Name != "x" {
		t.Fatalf("got %v %v", got, err)
	}
}

func TestReadCSVMissingFileReturnsError(t *testing.T) {
	if _, err := ReadCSV(filepath.Join(t.TempDir(), "nope.csv")); err == nil {
		t.Fatal("want error")
	}
}
