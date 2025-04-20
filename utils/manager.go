package utils

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

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

var (
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))             // muted green
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))             // red
	warningStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))            // yellow
	pendingStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("12"))            // blue
	infoStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("14"))            // cyan
	debugStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("250"))           // light grey
	detailStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("13"))            // purple
	streamStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))           // grey
	headerStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("69")) // purple
	// More styles
	// whiteStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("255")) // white
	// darkRedStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("124")) // dark red
	// darkGreenStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("28"))  // dark green
	// darkYellowStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("142")) // dark yellow
	// darkBlueStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("19"))  // dark blue
	// darkCyanStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("31"))  // dark cyan
	// orangeStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("208")) // orange
	// pinkStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("200")) // pink
	// tealStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("37"))  // teal
	// More variables
	basePadding = 2
)

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

func PrintProgress(current, total int, width int) string {
	if width <= 0 {
		width = 30
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
	return debugStyle.Render(fmt.Sprintf("%s %.1f%% - ", bar, percent*100))
}

// FunctionOutput represents the output of a specific function
type FunctionOutput struct {
	Name        string
	Status      string
	Message     string
	StreamLines []string
	Complete    bool
	StartTime   time.Time
	LastUpdated time.Time
	Error       error             // Store the error for detailed output in debug mode
	Tables      map[string]*Table // Tables associated with this function
	Index       int               // Order in which this function was registered
}

// ErrorReport represents an error from a function for summary reporting
type ErrorReport struct {
	FunctionName string
	Error        error
	Time         time.Time
}

// OutputManager handles all terminal output management
type Manager struct {
	outputs         map[string]*FunctionOutput
	mutex           sync.RWMutex
	numLines        int
	maxStreams      int               // Max stream lines per function
	unlimitedOutput bool              // When true, unlimited output per function
	tables          map[string]*Table // Global tables that can be displayed
	functionTables  []Table           // Tables for functions to be displayed at the end
	errors          []ErrorReport     // Collection of errors for debug output
	doneCh          chan struct{}     // Channel to signal stopping the display
	pauseCh         chan bool         // Channel to pause/resume display updates
	isPaused        bool
	displayTick     time.Duration  // Interval between display updates
	functionCount   int            // Counter for function index
	displayWg       sync.WaitGroup // WaitGroup to coordinate display goroutine shutdown
	tablesDisplayed bool           // Flag to track if tables have been displayed
}

// NewManager creates a new output manager
func NewManager(maxStreams int) *Manager {
	if maxStreams <= 0 {
		maxStreams = 15 // Default value - updated to 15 as requested
	}
	return &Manager{
		outputs:         make(map[string]*FunctionOutput),
		tables:          make(map[string]*Table),
		functionTables:  []Table{},
		errors:          []ErrorReport{},
		maxStreams:      maxStreams,
		unlimitedOutput: false,
		doneCh:          make(chan struct{}),
		pauseCh:         make(chan bool),
		isPaused:        false,
		displayTick:     200 * time.Millisecond, // 200ms default update interval
		functionCount:   0,
		tablesDisplayed: false,
	}
}

// SetUnlimitedOutput toggles whether output should be unlimited
func (m *Manager) SetUnlimitedOutput(unlimited bool) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.unlimitedOutput = unlimited
}

// Pause pauses the display updates
func (m *Manager) Pause() {
	if !m.isPaused {
		m.pauseCh <- true
		m.isPaused = true
	}
}

// Resume resumes the display updates
func (m *Manager) Resume() {
	if m.isPaused {
		m.pauseCh <- false
		m.isPaused = false
	}
}

// Register adds a new function to be tracked
func (m *Manager) Register(name string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.functionCount++
	m.outputs[name] = &FunctionOutput{
		Name:        name,
		Status:      "pending",
		StreamLines: []string{},
		StartTime:   time.Now(),
		LastUpdated: time.Now(),
		Tables:      make(map[string]*Table),
		Index:       m.functionCount,
	}
}

// SetMessage sets the primary message for a function
func (m *Manager) SetMessage(name, message string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if info, exists := m.outputs[name]; exists {
		info.Message = message
		info.LastUpdated = time.Now()
	}
}

// SetStatus updates the status of a function
func (m *Manager) SetStatus(name, status string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if info, exists := m.outputs[name]; exists {
		info.Status = status
		info.LastUpdated = time.Now()
	}
}

// GetStatus retrieves the status of a function
func (m *Manager) GetStatus(name string) string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	if info, exists := m.outputs[name]; exists {
		return info.Status
	}
	return "unknown"
}

// Complete marks a function as complete
func (m *Manager) Complete(name string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if info, exists := m.outputs[name]; exists {
		if !m.unlimitedOutput {
			info.StreamLines = []string{}
		}
		info.Complete = true
		info.Status = "success"
		info.LastUpdated = time.Now()
	}
}

// ReportError sets a function's status to error
func (m *Manager) ReportError(name string, err error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if info, exists := m.outputs[name]; exists {
		info.Complete = true
		info.Status = "error"
		info.Message = fmt.Sprintf("Error: %v", err)
		info.Error = err
		info.LastUpdated = time.Now()

		// Add to errors collection for debug output
		m.errors = append(m.errors, ErrorReport{
			FunctionName: name,
			Error:        err,
			Time:         time.Now(),
		})
	}
}

// UpdateStreamOutput adds lines to a function's stream output
func (m *Manager) UpdateStreamOutput(name string, output []string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if info, exists := m.outputs[name]; exists {
		if m.unlimitedOutput { // just append
			info.StreamLines = append(info.StreamLines, output...)
		} else { // enforce size limit
			currentLen := len(info.StreamLines)
			if currentLen+len(output) > m.maxStreams {
				startIndex := currentLen + len(output) - m.maxStreams
				if startIndex > currentLen {
					startIndex = 0
				}
				newLines := append(info.StreamLines[startIndex:], output...)
				if len(newLines) > m.maxStreams {
					newLines = newLines[len(newLines)-m.maxStreams:]
				}
				info.StreamLines = newLines
			} else {
				info.StreamLines = append(info.StreamLines, output...)
			}
		}
		info.LastUpdated = time.Now()
	}
}

// AddProgressBar sets a progress bar as the stream content
// percentage: value between 0 and 100
// text: additional text to show after the progress bar
func (m *Manager) AddProgressBar(name string, percentage float64, text string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if info, exists := m.outputs[name]; exists {
		// Clamp percentage between 0 and 100
		if percentage < 0 {
			percentage = 0
		} else if percentage > 100 {
			percentage = 100
		}

		// Generate the progress bar using the existing utility function
		progressBar := PrintProgress(int(percentage), 100, 30)

		// Add text if provided
		display := progressBar
		display += debugStyle.Render(text)

		// Set this as the only stream line, replacing any existing lines
		info.StreamLines = []string{display}
		info.LastUpdated = time.Now()
	}
}

// AddStreamLine adds a single line to a function's stream output
func (m *Manager) AddStreamLine(name, line string) {
	m.UpdateStreamOutput(name, []string{line})
}

// ClearOutput clears the entire screen
func (m *Manager) ClearOutput() {
	fmt.Print("\033[H\033[2J")
	m.numLines = 0
}

// ClearLines clears 'n' previous lines
func (m *Manager) ClearLines(n int) {
	if n <= 0 {
		return
	}
	fmt.Printf("\033[%dA\033[J", n)
	m.numLines = max(m.numLines-n, 0)
}

// ClearFunction clears the output of a specific function
func (m *Manager) ClearFunction(name string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if info, exists := m.outputs[name]; exists {
		info.StreamLines = []string{}
		info.Message = ""
		info.LastUpdated = time.Now()
	}
}

// ClearAll clears all function outputs
func (m *Manager) ClearAll() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	for name := range m.outputs {
		m.outputs[name].StreamLines = []string{}
		m.outputs[name].Message = ""
		m.outputs[name].LastUpdated = time.Now()
	}
}

// GetStatusDisplay returns a styled status indicator based on status
func (m *Manager) GetStatusDisplay(status string) string {
	switch status {
	case "success", "pass":
		return successStyle.Render(StyleSymbols["pass"])
	case "error", "fail":
		return errorStyle.Render(StyleSymbols["fail"])
	case "warning":
		return warningStyle.Render(StyleSymbols["warning"])
	case "pending":
		return pendingStyle.Render(StyleSymbols["pending"])
	default:
		return infoStyle.Render(StyleSymbols["bullet"])
	}
}

// RegisterTable adds a table to the manager
func (m *Manager) RegisterTable(name string, headers []string) *Table {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	table := NewTable(headers)
	m.tables[name] = table
	return table
}

// RegisterFunctionTable adds a table to a specific function
func (m *Manager) RegisterFunctionTable(funcName string, name string, headers []string) *Table {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if info, exists := m.outputs[funcName]; exists {
		table := NewTable(headers)
		info.Tables[name] = table
		return table
	}
	return nil
}

// GetTable retrieves a table by name
func (m *Manager) GetTable(name string) *Table {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.tables[name]
}

// GetFunctionTable retrieves a function's table by name
func (m *Manager) GetFunctionTable(funcName string, tableName string) *Table {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	if info, exists := m.outputs[funcName]; exists {
		if table, tableExists := info.Tables[tableName]; tableExists {
			return table
		}
	}
	return nil
}

// DisplayTable displays a specific table
func (m *Manager) DisplayTable(name string, innerDividers bool) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	if table, exists := m.tables[name]; exists {
		if m.numLines > 0 {
			fmt.Printf("\033[%dA\033[J", m.numLines)
		}
		tableStr := table.FormatLipglossTable(innerDividers)
		fmt.Print(tableStr)
		m.numLines = strings.Count(tableStr, "\n")
	}
}

// RemoveTable removes a table from the manager
func (m *Manager) RemoveTable(name string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	delete(m.tables, name)
}

// updateDisplay updates the console display with all function outputs
func (m *Manager) updateDisplay() {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	if m.numLines > 0 && !m.unlimitedOutput {
		fmt.Printf("\033[%dA\033[J", m.numLines)
	}
	lineCount := 0

	// Sort functions by their registration order
	type sortableFunc struct {
		name  string
		info  *FunctionOutput
		index int
	}

	var funcs []sortableFunc
	for name, info := range m.outputs {
		funcs = append(funcs, sortableFunc{name, info, info.Index})
	}

	// Sort by index (registration order)
	sort.Slice(funcs, func(i, j int) bool {
		return funcs[i].index < funcs[j].index
	})

	// Group functions: active, pending, completed
	var activeFuncs []sortableFunc
	var pendingFuncs []sortableFunc
	var completedFuncs []sortableFunc

	for _, f := range funcs {
		if f.info.Complete {
			completedFuncs = append(completedFuncs, f)
		} else if f.info.Status == "pending" && f.info.Message == "" {
			pendingFuncs = append(pendingFuncs, f)
		} else {
			activeFuncs = append(activeFuncs, f)
		}
	}

	// Print active functions
	for idx, f := range activeFuncs {
		info := f.info
		statusDisplay := m.GetStatusDisplay(info.Status)

		// Calculate elapsed time
		elapsed := time.Since(info.StartTime).Round(time.Millisecond)
		elapsedStr := fmt.Sprintf("[%s]", elapsed)

		// Style the message based on status
		styledMessage := info.Message
		prefixStyle := pendingStyle

		switch info.Status {
		case "success":
			styledMessage = successStyle.Render(info.Message)
			prefixStyle = successStyle
		case "error":
			styledMessage = errorStyle.Render(info.Message)
			prefixStyle = errorStyle
		case "warning":
			styledMessage = warningStyle.Render(info.Message)
			prefixStyle = warningStyle
		case "pending":
			styledMessage = pendingStyle.Render(info.Message)
			prefixStyle = pendingStyle
		}

		// Format with proper padding and numbering
		functionPrefix := strings.Repeat(" ", basePadding) +
			prefixStyle.Render(fmt.Sprintf("%d. ", idx+1))

		fmt.Printf("%s%s %s %s\n",
			functionPrefix,
			statusDisplay,
			debugStyle.Render(elapsedStr),
			styledMessage)
		lineCount++

		// Print stream lines with proper indentation
		if len(info.StreamLines) > 0 {
			indent := strings.Repeat(" ", basePadding+4) // Additional indentation for stream output
			for _, line := range info.StreamLines {
				fmt.Printf("%s%s\n", indent, streamStyle.Render(line))
				lineCount++
			}
		}
	}

	// Print pending functions
	for idx, f := range pendingFuncs {
		info := f.info
		statusDisplay := m.GetStatusDisplay(info.Status)

		// Format with proper padding and numbering
		functionPrefix := strings.Repeat(" ", basePadding) +
			pendingStyle.Render(fmt.Sprintf("%d. ", len(activeFuncs)+idx+1))

		fmt.Printf("%s%s %s\n",
			functionPrefix,
			statusDisplay,
			pendingStyle.Render("Waiting..."))
		lineCount++

		if len(info.StreamLines) > 0 {
			indent := strings.Repeat(" ", basePadding+4)
			for _, line := range info.StreamLines {
				fmt.Printf("%s%s\n", indent, streamStyle.Render(line))
				lineCount++
			}
		}
	}

	// Print completed functions
	for idx, f := range completedFuncs {
		info := f.info
		statusDisplay := m.GetStatusDisplay(info.Status)

		// Calculate total time
		totalTime := info.LastUpdated.Sub(info.StartTime).Round(time.Millisecond)
		timeStr := fmt.Sprintf("[%s]", totalTime)

		// Style based on status
		prefixStyle := successStyle
		if info.Status == "error" {
			prefixStyle = errorStyle
		}

		// Style message based on status
		styledMessage := info.Message
		if info.Status == "success" {
			styledMessage = successStyle.Render(info.Message)
		} else if info.Status == "error" {
			styledMessage = errorStyle.Render(info.Message)
		}

		// Format with proper padding and numbering
		functionPrefix := strings.Repeat(" ", basePadding) +
			prefixStyle.Render(fmt.Sprintf("%d. ", len(activeFuncs)+len(pendingFuncs)+idx+1))

		fmt.Printf("%s%s %s %s\n",
			functionPrefix,
			statusDisplay,
			debugStyle.Render(timeStr),
			styledMessage)
		lineCount++

		if m.unlimitedOutput && len(info.StreamLines) > 0 {
			indent := strings.Repeat(" ", basePadding+4)
			for _, line := range info.StreamLines {
				fmt.Printf("%s%s\n", indent, streamStyle.Render(line))
				lineCount++
			}
		}
	}

	m.numLines = lineCount
}

// StartDisplay starts the automatic display update goroutine
func (m *Manager) StartDisplay() {
	m.displayWg.Add(1)
	go func() {
		defer m.displayWg.Done()
		ticker := time.NewTicker(m.displayTick)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if !m.isPaused {
					m.updateDisplay()
				}
			case pauseState := <-m.pauseCh:
				m.isPaused = pauseState
			case <-m.doneCh:
				if !m.unlimitedOutput {
					m.ClearAll()
				}
				// Display final output
				m.displayTables()
				m.ShowSummary()
				return
			}
		}
	}()
}

// StopDisplay stops the automatic display updates and waits for completion
func (m *Manager) StopDisplay() {
	close(m.doneCh)    // Signal the goroutine to stop
	m.displayWg.Wait() // Wait for goroutine to finish
}

// SetUpdateInterval sets the interval between display updates
func (m *Manager) SetUpdateInterval(interval time.Duration) {
	m.displayTick = interval
}

// displayTables displays all tables at the end
func (m *Manager) displayTables() {
	m.mutex.Lock()
	m.tablesDisplayed = true
	m.mutex.Unlock()

	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// First display global tables
	if len(m.tables) > 0 {
		fmt.Println(strings.Repeat(" ", basePadding) + headerStyle.Render("Global Tables:"))
		for name, table := range m.tables {
			fmt.Println(strings.Repeat(" ", basePadding+2) + headerStyle.Render(name))
			tableStr := table.FormatLipglossTable(true)
			fmt.Println(tableStr)
		}
	}

	// Then display function tables
	hasFunctionTables := false
	for _, info := range m.outputs {
		if len(info.Tables) > 0 {
			hasFunctionTables = true
			break
		}
	}

	if hasFunctionTables {
		fmt.Println(strings.Repeat(" ", basePadding) + headerStyle.Render("Function Tables:"))
		for _, info := range m.outputs {
			if len(info.Tables) > 0 {
				fmt.Println(strings.Repeat(" ", basePadding+2) + headerStyle.Render(info.Name))
				for tableName, table := range info.Tables {
					fmt.Println(strings.Repeat(" ", basePadding+4) + infoStyle.Render(tableName))
					tableStr := table.FormatLipglossTable(true)
					fmt.Println(tableStr)
				}
			}
		}
	}
}

// displayErrors displays all errors in debug mode
func (m *Manager) displayErrors() {
	if len(m.errors) == 0 {
		return
	}

	fmt.Println()
	fmt.Println(strings.Repeat(" ", basePadding) + errorStyle.Bold(true).Render("Errors:"))

	for i, err := range m.errors {
		fmt.Printf("%s%s %s %s\n",
			strings.Repeat(" ", basePadding+2),
			errorStyle.Render(fmt.Sprintf("%d.", i+1)),
			debugStyle.Render(fmt.Sprintf("[%s]", err.Time.Format("15:04:05"))),
			errorStyle.Render(fmt.Sprintf("Function: %s", err.FunctionName)))

		fmt.Printf("%s%s\n",
			strings.Repeat(" ", basePadding+4),
			errorStyle.Render(fmt.Sprintf("Error: %v", err.Error)))
	}
}

// ShowSummary displays a final summary of all functions
func (m *Manager) ShowSummary() {
	m.mutex.RLock()
	alreadyDisplayedTables := m.tablesDisplayed
	m.mutex.RUnlock()

	m.mutex.Lock()
	m.tablesDisplayed = true
	m.mutex.Unlock()

	m.mutex.RLock()
	defer m.mutex.RUnlock()

	fmt.Println()

	var success, failures int
	totalTime := time.Duration(0)
	for _, info := range m.outputs {
		if info.Status == "success" {
			success++
		} else if info.Status == "error" {
			failures++
		}
		if info.Complete {
			totalTime += info.LastUpdated.Sub(info.StartTime)
		}
	}

	// Create a style for the summary
	summaryStyle := infoStyle.Padding(0, basePadding)

	// Format statistics with colors
	totalOps := fmt.Sprintf("Total Operations: %d", len(m.outputs))
	succeeded := fmt.Sprintf("Succeeded: %s", successStyle.Render(fmt.Sprintf("%d", success)))
	failed := fmt.Sprintf("Failed: %s", errorStyle.Render(fmt.Sprintf("%d", failures)))

	// Print the summary
	fmt.Println(summaryStyle.Render(fmt.Sprintf("%s, %s, %s", totalOps, succeeded, failed)))

	// Don't display tables if they've already been displayed by the display goroutine
	if !alreadyDisplayedTables {
		m.displayTables()
	}

	if m.unlimitedOutput {
		m.displayErrors()
	}
}

// Remove removes a function from the manager
func (m *Manager) Remove(name string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	delete(m.outputs, name)
}

// RemoveCompleted removes completed functions from the manager
func (m *Manager) RemoveCompleted() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	for name, info := range m.outputs {
		if info.Complete {
			delete(m.outputs, name)
		}
	}
}
