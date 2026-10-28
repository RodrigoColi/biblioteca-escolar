package sqlutil

import (
	"database/sql"
	"testing"
	"time"
)

func TestText(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"Ana", "'Ana'"},
		{"", "''"},
		{"O'Neil", "'O''Neil'"},
		{"''", "''''''"},
		{"Robert'); DROP TABLE alunos;--", "'Robert''); DROP TABLE alunos;--'"},
		{"Ação ç ã", "'Ação ç ã'"},
	}
	for _, tt := range tests {
		if got := Text(tt.in); got != tt.want {
			t.Errorf("Text(%q) = %s, want %s", tt.in, got, tt.want)
		}
	}
}

func TestNum(t *testing.T) {
	for n, want := range map[int]string{0: "0", 42: "42", -7: "-7"} {
		if got := Num(n); got != want {
			t.Errorf("Num(%d) = %s, want %s", n, got, want)
		}
	}
}

func TestDate(t *testing.T) {
	d := time.Date(2026, 2, 5, 23, 59, 0, 0, time.UTC)
	if got := Date(d); got != "'2026-02-05'" {
		t.Errorf("Date = %s", got)
	}
	if got := NullableDate(&d); got != "'2026-02-05'" {
		t.Errorf("NullableDate(&d) = %s", got)
	}
	if got := NullableDate(nil); got != "NULL" {
		t.Errorf("NullableDate(nil) = %s", got)
	}
}

func TestTimePtr(t *testing.T) {
	if TimePtr(sql.NullTime{}) != nil {
		t.Error("TimePtr of invalid NullTime should be nil")
	}
	d := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	got := TimePtr(sql.NullTime{Time: d, Valid: true})
	if got == nil || !got.Equal(d) {
		t.Errorf("TimePtr = %v, want %v", got, d)
	}
}
