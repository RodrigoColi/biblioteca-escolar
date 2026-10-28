package input

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

type exitCalled struct{ code int }

func newReader(lines ...string) (*Reader, *bytes.Buffer) {
	out := &bytes.Buffer{}
	r := NewReader(strings.NewReader(strings.Join(lines, "\n")+"\n"), out)
	r.exit = func(code int) { panic(exitCalled{code}) }
	return r, out
}

func date(s string) time.Time {
	t, err := time.Parse(dateLayout, s)
	if err != nil {
		panic(err)
	}
	return t
}

func TestTextRepromptsUntilNotEmpty(t *testing.T) {
	r, out := newReader("", "   ", "  Ana  ")
	if got := r.Text("Nome"); got != "Ana" {
		t.Errorf("got %q", got)
	}
	if n := strings.Count(out.String(), "Campo obrigatório"); n != 2 {
		t.Errorf("expected 2 warnings, got %d", n)
	}
}

func TestTextOr(t *testing.T) {
	r, out := newReader("", "Novo")
	if got := r.TextOr("Nome", "Atual"); got != "Atual" {
		t.Errorf("Enter should keep current, got %q", got)
	}
	if got := r.TextOr("Nome", "Atual"); got != "Novo" {
		t.Errorf("got %q", got)
	}
	if !strings.Contains(out.String(), "Nome [Atual]: ") {
		t.Errorf("prompt should show current value: %q", out.String())
	}
}

func TestIntRepromptsOnInvalid(t *testing.T) {
	r, out := newReader("abc", "1.5", "", "42")
	if got := r.Int("Matrícula"); got != 42 {
		t.Errorf("got %d", got)
	}
	if n := strings.Count(out.String(), "Informe um número inteiro"); n != 3 {
		t.Errorf("expected 3 warnings, got %d", n)
	}
}

func TestIntOr(t *testing.T) {
	r, _ := newReader("", "x", "7")
	if got := r.IntOr("Ano", 2000); got != 2000 {
		t.Errorf("Enter should keep current, got %d", got)
	}
	if got := r.IntOr("Ano", 2000); got != 7 {
		t.Errorf("got %d", got)
	}
}

func TestOptionRange(t *testing.T) {
	r, out := newReader("9", "-1", "x", "0")
	if got := r.Option("Opção", 0, 3); got != 0 {
		t.Errorf("got %d", got)
	}
	if n := strings.Count(out.String(), "entre 0 e 3"); n != 3 {
		t.Errorf("expected 3 warnings, got %d", n)
	}
}

func TestDateOr(t *testing.T) {
	current := date("10/10/2026")
	r, out := newReader("", "31/02/2026", "2026-10-01", "01/10/2026")
	if got := r.DateOr("Data", current); !got.Equal(current) {
		t.Errorf("Enter should keep current, got %v", got)
	}
	if got := r.DateOr("Data", current); !got.Equal(date("01/10/2026")) {
		t.Errorf("got %v", got)
	}
	if n := strings.Count(out.String(), "Data inválida"); n != 2 {
		t.Errorf("expected 2 warnings, got %d", n)
	}
}

func TestOptionalDate(t *testing.T) {
	current := date("10/10/2026")

	r, out := newReader("", "", "-", "invalida", "05/11/2026")
	if got := r.OptionalDate("Devolução", nil); got != nil {
		t.Errorf("Enter on empty should stay nil, got %v", got)
	}
	if got := r.OptionalDate("Devolução", &current); got == nil || !got.Equal(current) {
		t.Errorf("Enter should keep current, got %v", got)
	}
	if got := r.OptionalDate("Devolução", &current); got != nil {
		t.Errorf("'-' should clear, got %v", got)
	}
	if got := r.OptionalDate("Devolução", nil); got == nil || !got.Equal(date("05/11/2026")) {
		t.Errorf("got %v", got)
	}
	if !strings.Contains(out.String(), "[vazio]") || !strings.Contains(out.String(), "[10/10/2026]") {
		t.Errorf("prompts should show current value: %q", out.String())
	}
}

func TestConfirm(t *testing.T) {
	tests := []struct {
		in   string
		want bool
	}{
		{"S", true}, {"s", true}, {"sim", true}, {"SIM", true},
		{"N", false}, {"n", false}, {"nao", false}, {"não", false}, {"NÃO", false},
	}
	for _, tt := range tests {
		r, _ := newReader(tt.in)
		if got := r.Confirm("Continuar?"); got != tt.want {
			t.Errorf("Confirm(%q) = %v, want %v", tt.in, got, tt.want)
		}
	}

	r, out := newReader("talvez", "", "y", "s")
	if !r.Confirm("Continuar?") {
		t.Error("expected true after reprompts")
	}
	if n := strings.Count(out.String(), "Responda S ou N"); n != 3 {
		t.Errorf("expected 3 warnings, got %d", n)
	}
}

func TestEOFCallsExit(t *testing.T) {
	r, out := newReader()
	r.line("x") // consumes the single empty line
	defer func() {
		e, ok := recover().(exitCalled)
		if !ok || e.code != 0 {
			t.Fatalf("expected exit(0), got %v", e)
		}
		if !strings.Contains(out.String(), "Entrada encerrada") {
			t.Errorf("missing EOF message: %q", out.String())
		}
	}()
	r.Text("Nome")
}
