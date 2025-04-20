package utils

import (
	"fmt"
	"slices"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// StyleColors provides a set of commonly used colors
var StyleColors = map[string]lipgloss.Color{
	"red":          lipgloss.Color("9"),   // standard red
	"green":        lipgloss.Color("2"),   // muted green
	"yellow":       lipgloss.Color("11"),  // yellow
	"blue":         lipgloss.Color("12"),  // blue
	"purple":       lipgloss.Color("13"),  // purple
	"cyan":         lipgloss.Color("14"),  // cyan
	"white":        lipgloss.Color("255"), // white
	"black":        lipgloss.Color("0"),   // black
	"grey":         lipgloss.Color("240"), // grey
	"brightRed":    lipgloss.Color("196"), // bright red
	"brightGreen":  lipgloss.Color("46"),  // bright green
	"brightYellow": lipgloss.Color("226"), // bright yellow
	"brightBlue":   lipgloss.Color("33"),  // bright blue
	"brightPurple": lipgloss.Color("164"), // bright purple
	"brightCyan":   lipgloss.Color("51"),  // bright cyan
	"brightWhite":  lipgloss.Color("255"), // bright white
	"darkRed":      lipgloss.Color("124"), // dark red
	"darkGreen":    lipgloss.Color("28"),  // dark green
	"darkYellow":   lipgloss.Color("142"), // dark yellow
	"darkBlue":     lipgloss.Color("19"),  // dark blue
	"darkPurple":   lipgloss.Color("91"),  // dark purple
	"darkCyan":     lipgloss.Color("31"),  // dark cyan
	"darkWhite":    lipgloss.Color("245"), // dark white (light grey)
	"lightGrey":    lipgloss.Color("250"), // light grey
	"mediumGrey":   lipgloss.Color("244"), // medium grey
	"orange":       lipgloss.Color("208"), // orange
	"pink":         lipgloss.Color("200"), // pink
	"teal":         lipgloss.Color("37"),  // teal
	"success":      lipgloss.Color("2"),   // success color (muted green)
	"error":        lipgloss.Color("9"),   // error color (red)
	"warning":      lipgloss.Color("11"),  // warning color (yellow)
	"info":         lipgloss.Color("12"),  // info color (blue)
	"detail":       lipgloss.Color("13"),  // detail color (purple)
	"pass":         lipgloss.Color("2"),   // pass color (muted green)
	"fail":         lipgloss.Color("9"),   // fail color (red)
}

// StyleSymbols provides common symbols for statuses
var StyleSymbols = map[string]string{
	"pass":    "✓",
	"fail":    "✗",
	"warning": "!",
	"pending": "○",
	"info":    "ℹ",
	"arrow":   "→",
	"bullet":  "•",
	"dot":     "·",
}

// SuccessMessage formats a message with success styling
func SuccessMessage(msg string) string {
	return lipgloss.NewStyle().
		Foreground(StyleColors["success"]).
		Render(msg)
}

// ErrorMessage formats a message with error styling
func ErrorMessage(msg string) string {
	return lipgloss.NewStyle().
		Foreground(StyleColors["error"]).
		Render(msg)
}

// WarningMessage formats a message with warning styling
func WarningMessage(msg string) string {
	return lipgloss.NewStyle().
		Foreground(StyleColors["warning"]).
		Render(msg)
}

// InfoMessage formats a message with info styling
func InfoMessage(msg string) string {
	return lipgloss.NewStyle().
		Foreground(StyleColors["info"]).
		Render(msg)
}

// DetailMessage formats a message with detail styling
func DetailMessage(msg string) string {
	return lipgloss.NewStyle().
		Foreground(StyleColors["detail"]).
		Render(msg)
}

// DebugMessage formats a message with debug styling (grey)
func DebugMessage(msg string) string {
	return lipgloss.NewStyle().
		Foreground(StyleColors["grey"]).
		Render(msg)
}

// Colorize formats a message with the specified color
func Colorize(msg string, color string) string {
	colorValue, exists := StyleColors[color]
	if !exists {
		colorValue = StyleColors["white"]
	}
	return lipgloss.NewStyle().
		Foreground(colorValue).
		Render(msg)
}

// FormatTitle creates a formatted title with optional width and color
func FormatTitle(title string, width int, color string) string {
	if width <= 0 {
		width = 80
	}

	colorValue, exists := StyleColors[color]
	if !exists {
		colorValue = StyleColors["blue"]
	}

	style := lipgloss.NewStyle().
		Foreground(colorValue).
		Bold(true)

	padding := max(width-len(title)-4, 2)
	leftPad := padding / 2
	rightPad := padding - leftPad

	result := style.Render(fmt.Sprintf("[ %s%s%s ]",
		strings.Repeat("=", leftPad),
		title,
		strings.Repeat("=", rightPad)))

	return result
}

// PrintProgress creates a progress bar with the given percentage
func PrintProgress(current, total int, width int, color string) string {
	if width <= 0 {
		width = 30
	}

	colorValue, exists := StyleColors[color]
	if !exists {
		colorValue = StyleColors["blue"]
	}

	style := lipgloss.NewStyle().
		Foreground(colorValue)

	percent := float64(current) / float64(total)
	filled := min(int(percent*float64(width)), width)

	bar := "["
	bar += strings.Repeat("=", filled)
	if filled < width {
		bar += ">"
		bar += strings.Repeat(" ", width-filled-1)
	}
	bar += "]"

	return style.Render(fmt.Sprintf("%s %.1f%%", bar, percent*100))
}

// SliceSame checks if two slices contain the same elements (order-insensitive)
func SliceSame(slice1, slice2 []any) bool {
	if len(slice1) != len(slice2) {
		return false
	}
	for _, elem := range slice1 {
		if !slices.Contains(slice2, elem) {
			return false
		}
	}
	for _, elem := range slice2 {
		if !slices.Contains(slice1, elem) {
			return false
		}
	}
	return true
}

// FormatStatusSymbol returns a formatted status symbol with appropriate color
func FormatStatusSymbol(status string) string {
	switch status {
	case "success", "pass":
		return lipgloss.NewStyle().
			Foreground(StyleColors["success"]).
			Render(StyleSymbols["pass"])
	case "error", "fail":
		return lipgloss.NewStyle().
			Foreground(StyleColors["error"]).
			Render(StyleSymbols["fail"])
	case "warning":
		return lipgloss.NewStyle().
			Foreground(StyleColors["warning"]).
			Render(StyleSymbols["warning"])
	case "pending":
		return lipgloss.NewStyle().
			Foreground(StyleColors["info"]).
			Render(StyleSymbols["pending"])
	default:
		return lipgloss.NewStyle().
			Foreground(StyleColors["info"]).
			Render(StyleSymbols["bullet"])
	}
}

// PaddedString returns a string with the specified padding
func PaddedString(s string, padAmount int) string {
	return strings.Repeat(" ", padAmount) + s
}

// JoinStringsWithPadding joins strings with a specified padding
func JoinStringsWithPadding(str []string, padAmount int) string {
	if len(str) == 0 {
		return ""
	}

	result := str[0]
	for i := 1; i < len(str); i++ {
		result += strings.Repeat(" ", padAmount) + str[i]
	}

	return result
}
