package utils

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// Define style constants
var (
	// Theme colors
	successColor = lipgloss.Color("2")   // muted green
	errorColor   = lipgloss.Color("9")   // red
	warningColor = lipgloss.Color("11")  // yellow
	pendingColor = lipgloss.Color("12")  // blue
	infoColor    = lipgloss.Color("14")  // cyan
	streamColor  = lipgloss.Color("245") // gray
	detailColor  = lipgloss.Color("13")  // magenta
	timeColor    = lipgloss.Color("246") // light gray
	headerColor  = lipgloss.Color("69")  // blue/purple for function headers
	tableBorder  = lipgloss.Color("245") // table border color
	tableHeader  = lipgloss.Color("255") // white for table headers

	// Basic styles
	successStyle = lipgloss.NewStyle().Foreground(successColor)
	errorStyle   = lipgloss.NewStyle().Foreground(errorColor)
	warningStyle = lipgloss.NewStyle().Foreground(warningColor)
	pendingStyle = lipgloss.NewStyle().Foreground(pendingColor)
	infoStyle    = lipgloss.NewStyle().Foreground(infoColor)
	streamStyle  = lipgloss.NewStyle().Foreground(streamColor)
	detailStyle  = lipgloss.NewStyle().Foreground(detailColor)
	timeStyle    = lipgloss.NewStyle().Foreground(StyleColors["lightGrey"])

	// Status indicators
	successMark = successStyle.Render("✓")
	errorMark   = errorStyle.Render("✗")
	warningMark = warningStyle.Render("!")
	pendingMark = pendingStyle.Render("○")

	// Header style
	headerStyle = lipgloss.NewStyle().Bold(true).Foreground(headerColor)

	// Initial padding for all lines
	basePadding = 2
)

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
	displayTick     time.Duration // Interval between display updates
	functionCount   int           // Counter for function index
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

		// Determine color based on progress
		// colorName := "blue"
		// if percentage < 30 {
		// 	colorName = "yellow"
		// } else if percentage >= 80 {
		// 	colorName = "green"
		// }

		// Generate the progress bar using the existing utility function
		progressBar := PrintProgress(int(percentage), 100, 30)

		// Add text if provided
		display := progressBar
		display += Colorize(text, "lightGrey")

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
	return FormatStatusSymbol(status)
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
			timeStyle.Render(elapsedStr),
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
			timeStyle.Render(timeStr),
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
	go func() {
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
				m.updateDisplay()
				m.displayTables()
				if m.unlimitedOutput {
					m.displayErrors()
				}
				return
			}
		}
	}()
}

// StopDisplay stops the automatic display updates
func (m *Manager) StopDisplay() {
	close(m.doneCh)
}

// SetUpdateInterval sets the interval between display updates
func (m *Manager) SetUpdateInterval(interval time.Duration) {
	m.displayTick = interval
}

// displayTables displays all tables at the end
func (m *Manager) displayTables() {
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
			timeStyle.Render(fmt.Sprintf("[%s]", err.Time.Format("15:04:05"))),
			errorStyle.Render(fmt.Sprintf("Function: %s", err.FunctionName)))

		fmt.Printf("%s%s\n",
			strings.Repeat(" ", basePadding+4),
			errorStyle.Render(fmt.Sprintf("Error: %v", err.Error)))
	}
}

// ShowSummary displays a final summary of all functions
func (m *Manager) ShowSummary() {
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
	summaryStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(infoColor).
		Padding(0, basePadding)

	// Format statistics with colors
	totalOps := fmt.Sprintf("Total Operations: %d", len(m.outputs))
	succeeded := fmt.Sprintf("Succeeded: %s", successStyle.Render(fmt.Sprintf("%d", success)))
	failed := fmt.Sprintf("Failed: %s", errorStyle.Render(fmt.Sprintf("%d", failures)))

	// Print the summary
	fmt.Println(summaryStyle.Render(fmt.Sprintf("%s, %s, %s", totalOps, succeeded, failed)))
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
