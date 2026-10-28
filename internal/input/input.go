package input

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"
)

const dateLayout = "02/01/2006"

type Reader struct {
	scanner *bufio.Scanner
	out     io.Writer
	exit    func(code int)
}

func NewReader(in io.Reader, out io.Writer) *Reader {
	return &Reader{scanner: bufio.NewScanner(in), out: out, exit: os.Exit}
}

func (r *Reader) line(prompt string) string {
	fmt.Fprint(r.out, prompt)
	if !r.scanner.Scan() {
		fmt.Fprintln(r.out, "\nEntrada encerrada.")
		r.exit(0)
	}
	return strings.TrimSpace(r.scanner.Text())
}

func (r *Reader) invalid(msg string) {
	fmt.Fprintln(r.out, "  ! "+msg)
}

func (r *Reader) Text(label string) string {
	for {
		if v := r.line(label + ": "); v != "" {
			return v
		}
		r.invalid("Campo obrigatório.")
	}
}

func (r *Reader) TextOr(label, current string) string {
	if v := r.line(fmt.Sprintf("%s [%s]: ", label, current)); v != "" {
		return v
	}
	return current
}

func (r *Reader) Int(label string) int {
	for {
		if n, err := strconv.Atoi(r.line(label + ": ")); err == nil {
			return n
		}
		r.invalid("Informe um número inteiro.")
	}
}

func (r *Reader) IntOr(label string, current int) int {
	for {
		v := r.line(fmt.Sprintf("%s [%d]: ", label, current))
		if v == "" {
			return current
		}
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
		r.invalid("Informe um número inteiro.")
	}
}

func (r *Reader) Option(label string, min, max int) int {
	for {
		n, err := strconv.Atoi(r.line(label + ": "))
		if err == nil && n >= min && n <= max {
			return n
		}
		r.invalid(fmt.Sprintf("Escolha uma opção entre %d e %d.", min, max))
	}
}

func (r *Reader) DateOr(label string, current time.Time) time.Time {
	for {
		v := r.line(fmt.Sprintf("%s (dd/mm/aaaa) [%s]: ", label, current.Format(dateLayout)))
		if v == "" {
			return current
		}
		if t, err := time.Parse(dateLayout, v); err == nil {
			return t
		}
		r.invalid("Data inválida.")
	}
}

// OptionalDate keeps the current value on Enter and clears it on "-".
func (r *Reader) OptionalDate(label string, current *time.Time) *time.Time {
	shown := "vazio"
	if current != nil {
		shown = current.Format(dateLayout)
	}
	for {
		v := r.line(fmt.Sprintf("%s (dd/mm/aaaa, '-' para limpar) [%s]: ", label, shown))
		switch v {
		case "":
			return current
		case "-":
			return nil
		}
		if t, err := time.Parse(dateLayout, v); err == nil {
			return &t
		}
		r.invalid("Data inválida.")
	}
}

func (r *Reader) Confirm(question string) bool {
	for {
		switch strings.ToUpper(r.line(question + " (S/N): ")) {
		case "S", "SIM":
			return true
		case "N", "NAO", "NÃO":
			return false
		}
		r.invalid("Responda S ou N.")
	}
}

func (r *Reader) Pause() {
	r.line("\nPressione Enter para continuar...")
}
