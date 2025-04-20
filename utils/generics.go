package utils

import (
	"fmt"
	"slices"
	"strings"
)

var Colors = map[string]string{
	"red":            "\033[31m",
	"green":          "\033[32m",
	"yellow":         "\033[33m",
	"blue":           "\033[34m",
	"purple":         "\033[35m",
	"cyan":           "\033[36m",
	"white":          "\033[37m",
	"black":          "\033[30m",
	"grey":           "\033[90m",
	"brightRed":      "\033[91m",
	"brightGreen":    "\033[92m",
	"brightYellow":   "\033[93m",
	"brightBlue":     "\033[94m",
	"brightPurple":   "\033[95m",
	"brightCyan":     "\033[96m",
	"brightWhite":    "\033[97m",
	"darkRed":        "\033[38;5;124m",
	"darkGreen":      "\033[38;5;28m",
	"darkYellow":     "\033[38;5;214m",
	"darkBlue":       "\033[38;5;32m",
	"darkPurple":     "\033[38;5;93m",
	"darkCyan":       "\033[38;5;39m",
	"darkWhite":      "\033[38;5;255m",
	"bgRed":          "\033[41m",
	"bgGreen":        "\033[42m",
	"bgYellow":       "\033[43m",
	"bgBlue":         "\033[44m",
	"bgPurple":       "\033[45m",
	"bgCyan":         "\033[46m",
	"bgWhite":        "\033[47m",
	"bgBlack":        "\033[40m",
	"bgGrey":         "\033[100m",
	"bgBrightRed":    "\033[101m",
	"bgBrightGreen":  "\033[102m",
	"bgBrightYellow": "\033[103m",
	"bgBrightBlue":   "\033[104m",
	"bgBrightPurple": "\033[105m",
	"bgBrightCyan":   "\033[106m",
	"bgBrightWhite":  "\033[107m",
	"bold":           "\033[1m",
	"dim":            "\033[2m",
	"italic":         "\033[3m",
	"underline":      "\033[4m",
	"blink":          "\033[5m",
	"reset":          "\033[0m",
	"teal":           "\033[38;5;37m",
	"orange":         "\033[38;5;208m",
	"pink":           "\033[38;5;200m",
	"brown":          "\033[38;5;130m",
	"lightBlue":      "\033[38;5;153m",
	"lightGreen":     "\033[38;5;82m",
	"lightYellow":    "\033[38;5;226m",
	"lightPurple":    "\033[38;5;129m",
	"lightCyan":      "\033[38;5;87m",
	"lightWhite":     "\033[38;5;255m",
	"lightGrey":      "\033[38;5;245m",
	"lightBlack":     "\033[38;5;235m",
	"lightRed":       "\033[38;5;196m",
	"lightMagenta":   "\033[38;5;201m",
	"lightOrange":    "\033[38;5;214m",
	"lightPink":      "\033[38;5;213m",
	"lightBrown":     "\033[38;5;130m",
	"lightTeal":      "\033[38;5;37m",
	"lightTurquoise": "\033[38;5;51m",
	"lightViolet":    "\033[38;5;93m",
	"lightCoral":     "\033[38;5;196m",
	"lightSalmon":    "\033[38;5;210m",
	"lightLavender":  "\033[38;5;189m",
	"lightMint":      "\033[38;5;121m",
	"lightPeach":     "\033[38;5;214m",
	"lightLime":      "\033[38;5;190m",
	"lightOlive":     "\033[38;5;178m",
	"lightPlum":      "\033[38;5;129m",
	"lightSlate":     "\033[38;5;240m",
	"lightSteel":     "\033[38;5;246m",
	"lightSea":       "\033[38;5;44m",
	"lightSky":       "\033[38;5;153m",
	"pass":           "✓",
	"fail":           "✗",
}

// Formatting messages
func SuccessMessage(msg string) string {
	return fmt.Sprintf("%s%s%s", Colors["green"], msg, Colors["reset"])
}
func ErrorMessage(msg string) string {
	return fmt.Sprintf("%s%s%s", Colors["red"], msg, Colors["reset"])
}
func WarningMessage(msg string) string {
	return fmt.Sprintf("%s%s%s", Colors["yellow"], msg, Colors["reset"])
}
func InfoMessage(msg string) string {
	return fmt.Sprintf("%s%s%s", Colors["blue"], msg, Colors["reset"])
}
func DetailMessage(msg string) string {
	return fmt.Sprintf("%s%s%s", Colors["purple"], msg, Colors["reset"])
}
func DebugMessage(msg string) string {
	return fmt.Sprintf("%s%s%s", Colors["grey"], msg, Colors["reset"])
}
func Colorize(msg string, color string) string {
	colorCode, exists := Colors[color]
	if !exists {
		colorCode = Colors["reset"]
	}
	return fmt.Sprintf("%s%s%s", colorCode, msg, Colors["reset"])
}

// Creates a title with formatting
func FormatTitle(title string, width int, color string) string {
	if width <= 0 {
		width = 80
	}
	colorCode, exists := Colors[color]
	if !exists {
		colorCode = Colors["blue"]
	}
	padding := max(width-len(title)-4, 2)
	leftPad := padding / 2
	rightPad := padding - leftPad
	result := fmt.Sprintf("%s[ %s%s%s ]%s",
		colorCode,
		strings.Repeat("=", leftPad),
		title,
		strings.Repeat("=", rightPad),
		Colors["reset"])
	return result
}

// Prints a simple progress bar
func PrintProgress(current, total int, width int, color string) string {
	if width <= 0 {
		width = 30
	}
	colorCode, exists := Colors[color]
	if !exists {
		colorCode = Colors["blue"]
	}
	percent := float64(current) / float64(total)
	filled := min(int(percent*float64(width)), width)

	bar := "["
	bar += strings.Repeat("=", filled)
	if filled < width {
		bar += ">"
		bar += strings.Repeat(" ", width-filled-1)
	}
	bar += "]"
	return fmt.Sprintf("%s%s%s %.1f%%", colorCode, bar, Colors["reset"], percent*100)
}

// SliceSame checks if two slices are same, but is not order-sensitive
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
