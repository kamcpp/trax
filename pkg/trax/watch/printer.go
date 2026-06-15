package watch

import (
	"fmt"
	"testing"
)

// Printer is an interface for outputting saga watch status.
// This allows the watch functionality to be used both in CLI (fmt.Printf)
// and in tests (t.Logf).
type Printer interface {
	// Print outputs a formatted line (without trailing newline)
	Print(format string, args ...interface{})
	// PrintLine outputs a formatted line (with trailing newline)
	PrintLine(format string, args ...interface{})
	// PrintEmpty outputs an empty line
	PrintEmpty()
}

// FmtPrinter implements Printer using fmt.Printf for CLI usage
type FmtPrinter struct{}

// NewFmtPrinter creates a new FmtPrinter
func NewFmtPrinter() *FmtPrinter {
	return &FmtPrinter{}
}

func (p *FmtPrinter) Print(format string, args ...interface{}) {
	fmt.Printf(format, args...)
}

func (p *FmtPrinter) PrintLine(format string, args ...interface{}) {
	fmt.Printf(format+"\n", args...)
}

func (p *FmtPrinter) PrintEmpty() {
	fmt.Println()
}

// TestPrinter implements Printer using t.Logf for test usage
type TestPrinter struct {
	t *testing.T
}

// NewTestPrinter creates a new TestPrinter
func NewTestPrinter(t *testing.T) *TestPrinter {
	return &TestPrinter{t: t}
}

func (p *TestPrinter) Print(format string, args ...interface{}) {
	p.t.Helper()
	p.t.Logf(format, args...)
}

func (p *TestPrinter) PrintLine(format string, args ...interface{}) {
	p.t.Helper()
	p.t.Logf(format, args...)
}

func (p *TestPrinter) PrintEmpty() {
	p.t.Helper()
	p.t.Log("")
}
