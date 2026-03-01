package utilx

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
)

// Table is returned when NewTable() is called.
type Table struct {
	writer *tabwriter.Writer
}

// NewTable returns a new *tabwriter.Writer with default config
func NewTable() *Table {
	return &Table{
		writer: tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0),
	}
}

// AddLine will write a new table line
func (t *Table) AddLine(args ...interface{}) {
	formatString := t.buildFormatString(args)
	_, _ = fmt.Fprintf(t.writer, formatString, args...)
}

// AddHeader will write a new table line followed by a seperator
func (t *Table) AddHeader(args ...interface{}) {
	t.AddLine(args...)
	t.addSeparator(args)
}

// Print will write the table to the terminal
func (t *Table) Print() {
	_ = t.writer.Flush()
}

// addSeparator will write a new dash seperator line based on the args length
func (t *Table) addSeparator(args []interface{}) {
	var b bytes.Buffer
	for idx, arg := range args {
		length := len(fmt.Sprintf("%v", arg))
		b.WriteString(strings.Repeat("-", length))
		if idx+1 != len(args) {
			// Add a tab as long as its not the last column
			b.WriteString("\t")
		}
	}
	b.WriteString("\n")
	_, _ = b.WriteTo(t.writer)
}

// buildFormatString will build up the formatting string used by the *tabwriter.Writer
func (t *Table) buildFormatString(args []interface{}) string {
	var b bytes.Buffer
	for idx := range args {
		b.WriteString("%v")
		if idx+1 != len(args) {
			// Add a tab as long as its not the last column
			b.WriteString("\t")
		}
	}
	b.WriteString("\n")
	return b.String()
}
