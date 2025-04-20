package utils

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

// Table represents a data table
type Table struct {
	Headers []string
	Rows    [][]string
}

// NewTable creates a new table with the given headers
func NewTable(headers []string) *Table {
	return &Table{
		Headers: headers,
		Rows:    [][]string{},
	}
}

// AddRow adds a row to the table
func (t *Table) AddRow(row []string) {
	t.Rows = append(t.Rows, row)
}

// getTerminalWidth returns the width of the terminal in columns
func getTerminalWidth() int {
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err == nil && width > 0 {
		return width
	}
	return 80 // Default fallback width
}

// FormatLipglossTable formats the table using lipgloss styling
func (t *Table) FormatLipglossTable(innerDividers bool) string {
	if len(t.Headers) == 0 {
		return ""
	}

	// Get terminal width
	termWidth := getTerminalWidth()

	// Define styles
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("255")). // White
		Align(lipgloss.Center)

	cellStyle := lipgloss.NewStyle().
		Padding(0, 1).
		Foreground(lipgloss.Color("252")) // Light gray

	oddRowStyle := cellStyle.Copy().
		Foreground(lipgloss.Color("252"))

	evenRowStyle := cellStyle.Copy().
		Foreground(lipgloss.Color("245"))

	borderStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")) // Dark gray for borders

	// Calculate column widths
	colWidths := make([]int, len(t.Headers))
	for i, header := range t.Headers {
		colWidths[i] = len(header) + 2 // Add padding
	}

	// Check all rows for content length
	for _, row := range t.Rows {
		for i, cell := range row {
			if i < len(colWidths) {
				contentWidth := len(cell) + 2
				if contentWidth > colWidths[i] {
					colWidths[i] = contentWidth
				}
			}
		}
	}

	// Adjust for terminal width if needed
	totalWidth := 1 // Start with 1 for the left border
	for _, width := range colWidths {
		totalWidth += width + 1 // Add column width plus border
	}

	// If table would be too wide, reduce column widths proportionally
	if totalWidth > termWidth && termWidth > len(colWidths)*3 {
		extraWidth := totalWidth - termWidth
		// Calculate how much to reduce each column
		reduction := extraWidth / len(colWidths)
		remainingReduction := extraWidth % len(colWidths)

		for i := range colWidths {
			// Ensure minimum width of 3 (1 char + 2 padding)
			if colWidths[i] > 3 {
				colWidths[i] -= reduction
				if i < remainingReduction {
					colWidths[i]--
				}
				if colWidths[i] < 3 {
					colWidths[i] = 3
				}
			}
		}
	}

	// Create a renderer
	var sb strings.Builder

	// Create border characters with lipgloss styles
	border := lipgloss.Border{
		Top:          "─",
		Bottom:       "─",
		Left:         "│",
		Right:        "│",
		TopLeft:      "┌",
		TopRight:     "┐",
		BottomLeft:   "└",
		BottomRight:  "┘",
		MiddleLeft:   "├",
		MiddleRight:  "┤",
		MiddleTop:    "┬",
		MiddleBottom: "┴",
		Middle:       "┼",
	}

	// Build the table
	// Top border
	sb.WriteString(borderStyle.Render(border.TopLeft))
	for i, width := range colWidths {
		sb.WriteString(borderStyle.Render(strings.Repeat(border.Top, width)))
		if i < len(colWidths)-1 {
			sb.WriteString(borderStyle.Render(border.MiddleTop))
		}
	}
	sb.WriteString(borderStyle.Render(border.TopRight))
	sb.WriteString("\n")

	// Headers
	sb.WriteString(borderStyle.Render(border.Left))
	for i, header := range t.Headers {
		headerCell := headerStyle.Copy().Width(colWidths[i]).Render(header)
		sb.WriteString(headerCell)
		if i < len(t.Headers)-1 {
			sb.WriteString(borderStyle.Render(border.Left))
		}
	}
	sb.WriteString(borderStyle.Render(border.Right))
	sb.WriteString("\n")

	// Header-Data separator
	sb.WriteString(borderStyle.Render(border.MiddleLeft))
	for i, width := range colWidths {
		sb.WriteString(borderStyle.Render(strings.Repeat(border.Bottom, width)))
		if i < len(colWidths)-1 {
			sb.WriteString(borderStyle.Render(border.Middle))
		}
	}
	sb.WriteString(borderStyle.Render(border.MiddleRight))
	sb.WriteString("\n")

	// Data rows
	for rowIdx, row := range t.Rows {
		rowStyle := oddRowStyle
		if rowIdx%2 == 0 {
			rowStyle = evenRowStyle
		}

		sb.WriteString(borderStyle.Render(border.Left))
		for colIdx, width := range colWidths {
			var cell string
			if colIdx < len(row) {
				cell = row[colIdx]
			} else {
				cell = ""
			}
			cellContent := rowStyle.Copy().Width(width).Render(cell)
			sb.WriteString(cellContent)
			if colIdx < len(colWidths)-1 {
				sb.WriteString(borderStyle.Render(border.Left))
			}
		}
		sb.WriteString(borderStyle.Render(border.Right))
		sb.WriteString("\n")

		// Inner row divider
		if innerDividers && rowIdx < len(t.Rows)-1 {
			sb.WriteString(borderStyle.Render(border.MiddleLeft))
			for i, width := range colWidths {
				sb.WriteString(borderStyle.Render(strings.Repeat(border.Bottom, width)))
				if i < len(colWidths)-1 {
					sb.WriteString(borderStyle.Render(border.Middle))
				}
			}
			sb.WriteString(borderStyle.Render(border.MiddleRight))
			sb.WriteString("\n")
		}
	}

	// Bottom border
	sb.WriteString(borderStyle.Render(border.BottomLeft))
	for i, width := range colWidths {
		sb.WriteString(borderStyle.Render(strings.Repeat(border.Bottom, width)))
		if i < len(colWidths)-1 {
			sb.WriteString(borderStyle.Render(border.MiddleBottom))
		}
	}
	sb.WriteString(borderStyle.Render(border.BottomRight))
	sb.WriteString("\n")

	return sb.String()
}

// FormatMarkdownTable formats the table as a markdown table
func (t *Table) FormatMarkdownTable() string {
	var sb strings.Builder

	// Calculate column widths
	colWidths := make([]int, len(t.Headers))
	for i, header := range t.Headers {
		colWidths[i] = len(header)
	}

	for _, row := range t.Rows {
		for i, cell := range row {
			if i < len(colWidths) && len(cell) > colWidths[i] {
				colWidths[i] = len(cell)
			}
		}
	}

	// Write header row
	sb.WriteString("|")
	for i, header := range t.Headers {
		format := fmt.Sprintf(" %%-%ds |", colWidths[i])
		sb.WriteString(fmt.Sprintf(format, header))
	}
	sb.WriteString("\n")

	// Write separator row
	sb.WriteString("|")
	for i := range t.Headers {
		sb.WriteString(fmt.Sprintf(" %s |", strings.Repeat("-", colWidths[i])))
	}
	sb.WriteString("\n")

	// Write data rows
	for _, row := range t.Rows {
		sb.WriteString("|")
		for i, width := range colWidths {
			var cell string
			if i < len(row) {
				cell = row[i]
			} else {
				cell = ""
			}
			format := fmt.Sprintf(" %%-%ds |", width)
			sb.WriteString(fmt.Sprintf(format, cell))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// PrintTable prints the table to stdout
func (t *Table) PrintTable(innerDividers bool) {
	fmt.Print(t.FormatLipglossTable(innerDividers))
}

// PrintMarkdownTable prints the markdown table to stdout
func (t *Table) PrintMarkdownTable() {
	fmt.Print(t.FormatMarkdownTable())
}

// WriteMarkdownTableToFile writes the markdown table to a file
func (t *Table) WriteMarkdownTableToFile(outputPath string) error {
	formatted := t.FormatMarkdownTable()
	var outFile *os.File
	_, err := os.Stat(outputPath)
	if err != nil {
		if os.IsNotExist(err) {
			outFile, err = os.Create(outputPath)
			if err != nil {
				return fmt.Errorf("error creating file: %w", err)
			}
			defer outFile.Close()
		} else {
			return fmt.Errorf("error checking file: %w", err)
		}
	} else {
		outFile, err = os.OpenFile(outputPath, os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			return fmt.Errorf("error opening file: %w", err)
		}
		defer outFile.Close()
	}
	_, err = outFile.WriteString(formatted)
	if err != nil {
		return fmt.Errorf("error writing to file: %w", err)
	}
	return nil
}
