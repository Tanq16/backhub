package utils

import (
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

// Table represents a data table wrapper over lipgloss table
type Table struct {
	Headers []string
	Rows    [][]string
	table   *table.Table
}

// NewTable creates a new table with the given headers
func NewTable(headers []string) *Table {
	t := &Table{
		Headers: headers,
		Rows:    [][]string{},
	}

	// Initialize the lipgloss table with headers
	t.table = table.New().Headers(headers...)

	return t
}

// AddRow adds a row to the table
func (t *Table) AddRow(row []string) {
	t.Rows = append(t.Rows, row)
	t.table.Row(row...)
}

// FormatLipglossTable formats the table using lipgloss styling
func (t *Table) FormatLipglossTable(innerDividers bool) string {
	// Use default lipgloss table styling
	return t.table.String()
}

// PrintTable prints the table to stdout
func (t *Table) PrintTable(innerDividers bool) {
	os.Stdout.WriteString(t.table.String())
}

// PrintMarkdownTable prints the markdown table to stdout
func (t *Table) PrintMarkdownTable() {
	mdTable := t.table.Border(lipgloss.MarkdownBorder())
	os.Stdout.WriteString(mdTable.String())
}

// FormatMarkdownTable formats the table as a markdown table
func (t *Table) FormatMarkdownTable() string {
	return t.table.Border(lipgloss.MarkdownBorder()).String()
}

// WriteMarkdownTableToFile writes the markdown table to a file
func (t *Table) WriteMarkdownTableToFile(outputPath string) error {
	formatted := t.FormatMarkdownTable()
	return os.WriteFile(outputPath, []byte(formatted), 0644)
}
