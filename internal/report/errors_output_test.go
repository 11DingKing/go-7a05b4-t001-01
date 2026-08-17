package report

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"goldbar/internal/model"
)

// threeErrorRows returns three row errors in a deliberately unsorted order so
// that both the completeness and the line-number ordering of the error report
// are exercised.
func threeErrorRows() []model.LineError {
	return []model.LineError{
		{LineNumber: 7, OrderID: "A006", Code: "DISCOUNT_OVER_SALE", Message: "旧金折现额 7523.77 超过新品售价 780.50"},
		{LineNumber: 3, OrderID: "A002", Code: "KARAT_TOO_LOW", Message: "成色 750 低于 900"},
		{LineNumber: 5, OrderID: "A004", Code: "NEGATIVE_WEIGHT", Message: "克重为负: old_weight=-1 new_weight=5"},
	}
}

func wantErrorCSVRows() [][]string {
	return [][]string{
		{"line_number", "order_id", "code", "message"},
		{"3", "A002", "KARAT_TOO_LOW", "成色 750 低于 900"},
		{"5", "A004", "NEGATIVE_WEIGHT", "克重为负: old_weight=-1 new_weight=5"},
		{"7", "A006", "DISCOUNT_OVER_SALE", "旧金折现额 7523.77 超过新品售价 780.50"},
	}
}

func assertErrorCSV(t *testing.T, raw []byte) {
	t.Helper()
	rows, err := csv.NewReader(strings.NewReader(string(raw))).ReadAll()
	if err != nil {
		t.Fatalf("parse %s: %v", ErrorFile, err)
	}
	want := wantErrorCSVRows()
	if len(rows) != len(want) {
		t.Fatalf("%s has %d row(s) %q, want %d row(s) %q", ErrorFile, len(rows), rows, len(want), want)
	}
	for i := range want {
		if len(rows[i]) != len(want[i]) {
			t.Fatalf("%s row %d = %q, want %q", ErrorFile, i, rows[i], want[i])
		}
		for j := range want[i] {
			if rows[i][j] != want[i][j] {
				t.Errorf("%s row %d col %d = %q, want %q", ErrorFile, i, j, rows[i][j], want[i][j])
			}
		}
	}
}

// TestBuildErrorReportContainsEveryErrorRow asserts the full content of
// errors.csv: one data row per row error, ordered by source line number, with
// line number, order id, code and message preserved.
func TestBuildErrorReportContainsEveryErrorRow(t *testing.T) {
	r := sampleResult()
	r.Settlements = []model.Settlement{validSettlement()}
	r.Errors = threeErrorRows()

	files := Build(r)
	raw, ok := files[ErrorFile]
	if !ok {
		t.Fatalf("%s missing from build output", ErrorFile)
	}
	assertErrorCSV(t, raw)
}

// TestWriteAllErrorReportContainsEveryErrorRow asserts the same content for the
// file actually committed to the output directory.
func TestWriteAllErrorReportContainsEveryErrorRow(t *testing.T) {
	r := sampleResult()
	r.Settlements = []model.Settlement{validSettlement()}
	r.Errors = threeErrorRows()
	r.Summary = model.Summary{
		StoreID: "S", StoreName: "N", TradeDate: "2026-08-17", GoldPrice: 768.5,
		Currency: "CNY", TotalOrders: 4, ValidOrders: 1, ErrorOrders: 3,
		TotalPayable: 4739.25, NetStoreReceivable: 4739.25,
	}

	dir := t.TempDir()
	if err := WriteAll(dir, r); err != nil {
		t.Fatalf("WriteAll: %v", err)
	}
	raw, err := os.ReadFile(filepath.Join(dir, ErrorFile))
	if err != nil {
		t.Fatalf("read %s: %v", ErrorFile, err)
	}
	assertErrorCSV(t, raw)
}

// TestBuildErrorReportSingleRow covers the smallest possible error report.
func TestBuildErrorReportSingleRow(t *testing.T) {
	r := sampleResult()
	r.Settlements = []model.Settlement{validSettlement()}
	r.Errors = []model.LineError{
		{LineNumber: 4, OrderID: "A003", Code: "KARAT_TOO_LOW", Message: "成色 750 低于 900"},
	}

	raw, ok := Build(r)[ErrorFile]
	if !ok {
		t.Fatalf("%s missing from build output", ErrorFile)
	}
	rows, err := csv.NewReader(strings.NewReader(string(raw))).ReadAll()
	if err != nil {
		t.Fatalf("parse %s: %v", ErrorFile, err)
	}
	want := [][]string{
		{"line_number", "order_id", "code", "message"},
		{"4", "A003", "KARAT_TOO_LOW", "成色 750 低于 900"},
	}
	if len(rows) != len(want) {
		t.Fatalf("%s has %d row(s) %q, want %d row(s) %q", ErrorFile, len(rows), rows, len(want), want)
	}
	for i := range want {
		for j := range want[i] {
			if rows[i][j] != want[i][j] {
				t.Errorf("%s row %d col %d = %q, want %q", ErrorFile, i, j, rows[i][j], want[i][j])
			}
		}
	}
}

// TestBuildErrorReportLeavesCallerSliceUntouched asserts the caller's error
// slice is not reordered or rewritten by rendering the report.
func TestBuildErrorReportLeavesCallerSliceUntouched(t *testing.T) {
	r := sampleResult()
	r.Settlements = []model.Settlement{validSettlement()}
	r.Errors = threeErrorRows()
	before := threeErrorRows()

	_ = Build(r)

	if len(r.Errors) != len(before) {
		t.Fatalf("caller error slice length = %d, want %d", len(r.Errors), len(before))
	}
	for i := range before {
		if r.Errors[i] != before[i] {
			t.Errorf("caller error slice[%d] = %+v, want %+v", i, r.Errors[i], before[i])
		}
	}
}
