package utils

import (
	"fmt"
	"math"
	"os"
	"strings"

	"golang.org/x/term"
)

var boxChars = map[string]string{
	"topLeft":     "╭",
	"topRight":    "╮",
	"bottomLeft":  "╰",
	"bottomRight": "╯",
	"horizontal":  "─",
	"vertical":    "│",
	"leftT":       "├",
	"rightT":      "┤",
	"topT":        "┬",
	"bottomT":     "┴",
	"cross":       "┼",
}

type Table struct {
	Headers []string
	Rows    [][]string
}

func NewTable(headers []string) *Table {
	return &Table{
		Headers: headers,
		Rows:    [][]string{},
	}
}

func (t *Table) AddRow(row []string) {
	t.Rows = append(t.Rows, row)
}

func getTerminalWidth() int {
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err == nil && width > 0 {
		return width
	}
	return 80 // Default fallback width
}

func wrapText(text string, width int) []string {
	if width <= 0 || len(text) <= width {
		return []string{text}
	}
	var lines []string
	var currentLine string
	words := strings.Fields(text)
	for _, word := range words {
		if len(currentLine)+len(word)+1 <= width {
			if currentLine == "" {
				currentLine = word
			} else {
				currentLine += " " + word
			}
		} else {
			if currentLine != "" {
				lines = append(lines, currentLine)
			}
			if len(word) > width {
				for len(word) > 0 {
					if len(word) <= width {
						currentLine = word
						break
					}
					lines = append(lines, word[:width])
					word = word[width:]
				}
			} else {
				currentLine = word
			}
		}
	}
	if currentLine != "" {
		lines = append(lines, currentLine)
	}
	return lines
}

func (t *Table) FormatTable(innerDividers bool) string {
	var outstring strings.Builder
	termWidth := getTerminalWidth()

	// Determine column widths
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

	// Calculate total width and adjust if needed
	borderChars := 1 + len(colWidths)
	paddingChars := len(colWidths) * 2
	contentWidth := 0
	for _, w := range colWidths {
		contentWidth += w
	}
	totalWidth := borderChars + paddingChars + contentWidth

	// Adjust column widths proportionally if table is too wide
	if totalWidth > termWidth && termWidth > (borderChars+paddingChars+len(colWidths)) {
		availableWidth := termWidth - borderChars - paddingChars
		totalContentWidth := contentWidth
		for i := range colWidths {
			ratio := float64(colWidths[i]) / float64(totalContentWidth)
			colWidths[i] = max(int(math.Floor(ratio*float64(availableWidth))), 3)
		}
	}
	totalWidth = 1 // left border
	for _, width := range colWidths {
		totalWidth += width + 2 // add column width + padding
	}
	totalWidth += len(colWidths) // add right border and dividers

	// Draw top border & headers
	outstring.WriteString(boxChars["topLeft"])
	for i, width := range colWidths {
		outstring.WriteString(strings.Repeat(boxChars["horizontal"], width+2))
		if i < len(colWidths)-1 {
			outstring.WriteString(boxChars["topT"])
		}
	}
	outstring.WriteString(boxChars["topRight"] + "\n")
	wrappedHeaders := make([][]string, len(t.Headers))
	maxHeaderLines := 1

	// Wrap headers to fit column widths
	for i, header := range t.Headers {
		wrappedHeaders[i] = wrapText(header, colWidths[i])
		if len(wrappedHeaders[i]) > maxHeaderLines {
			maxHeaderLines = len(wrappedHeaders[i])
		}
	}

	// Print header rows
	for line := range maxHeaderLines {
		outstring.WriteString(boxChars["vertical"])
		for i, wrappedHeader := range wrappedHeaders {
			headerLine := ""
			if line < len(wrappedHeader) {
				headerLine = wrappedHeader[line]
			}
			format := fmt.Sprintf(" %%-%ds ", colWidths[i])
			paddedHeader := fmt.Sprintf(format, headerLine)
			paddedHeader = Colors["bold"] + paddedHeader + Colors["reset"]
			outstring.WriteString(paddedHeader)
			if i < len(wrappedHeaders)-1 {
				outstring.WriteString(boxChars["vertical"])
			}
		}
		outstring.WriteString(boxChars["vertical"] + "\n")
	}

	// Draw header/data divider
	outstring.WriteString(boxChars["leftT"])
	for i, width := range colWidths {
		outstring.WriteString(strings.Repeat(boxChars["horizontal"], width+2))
		if i < len(colWidths)-1 {
			outstring.WriteString(boxChars["cross"])
		}
	}
	outstring.WriteString(boxChars["rightT"] + "\n")

	// Draw data rows
	for r, row := range t.Rows {
		wrappedCells := make([][]string, len(row))
		maxLines := 1
		// Wrap cells to fit column widths
		for i, cell := range row {
			if i < len(colWidths) {
				wrappedCells[i] = wrapText(cell, colWidths[i])
				if len(wrappedCells[i]) > maxLines {
					maxLines = len(wrappedCells[i])
				}
			}
		}
		// Print wrapped cell content
		for line := range maxLines {
			outstring.WriteString(boxChars["vertical"])
			for i := range colWidths {
				cellLine := ""
				if i < len(wrappedCells) && line < len(wrappedCells[i]) {
					cellLine = wrappedCells[i][line]
				}
				format := fmt.Sprintf(" %%-%ds ", colWidths[i])
				outstring.WriteString(fmt.Sprintf(format, cellLine))
				if i < len(colWidths)-1 {
					outstring.WriteString(boxChars["vertical"])
				}
			}
			outstring.WriteString(boxChars["vertical"] + "\n")
		}

		// Draw inner row dividers
		if r < len(t.Rows)-1 && innerDividers {
			outstring.WriteString(boxChars["leftT"])
			for i, width := range colWidths {
				outstring.WriteString(strings.Repeat(boxChars["horizontal"], width+2))
				if i < len(colWidths)-1 {
					outstring.WriteString(boxChars["cross"])
				}
			}
			outstring.WriteString(boxChars["rightT"] + "\n")
		}
	}

	// Draw bottom border
	outstring.WriteString(boxChars["bottomLeft"])
	for i, width := range colWidths {
		outstring.WriteString(strings.Repeat(boxChars["horizontal"], width+2))
		if i < len(colWidths)-1 {
			outstring.WriteString(boxChars["bottomT"])
		}
	}
	outstring.WriteString(boxChars["bottomRight"] + "\n")
	return outstring.String()
}

func (t *Table) FormatMarkdownTable() string {
	var outstring strings.Builder

	// Determine column width
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
	headerRow := "|"
	dividerRow := "|"
	for i, header := range t.Headers {
		formatter := fmt.Sprintf(" %%-%ds |", colWidths[i])
		headerRow += fmt.Sprintf(formatter, header)
		dividerRow += fmt.Sprintf(" %s |", strings.Repeat("-", colWidths[i]))
	}
	headerRow += "\n"
	dividerRow += "\n"
	outstring.WriteString(headerRow)
	outstring.WriteString(dividerRow)

	// Write data rows
	for _, row := range t.Rows {
		rowText := "|"
		for i, cell := range row {
			if i < len(colWidths) {
				formatter := fmt.Sprintf(" %%-%ds |", colWidths[i])
				rowText += fmt.Sprintf(formatter, cell)
			}
		}
		rowText += "\n"
		outstring.WriteString(rowText)
	}
	return outstring.String()
}

func (t *Table) PrintTable(innerDividers bool) {
	fmt.Print(t.FormatTable(innerDividers))
}

func (t *Table) PrintMarkdownTable() {
	fmt.Print(t.FormatMarkdownTable())
}

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
