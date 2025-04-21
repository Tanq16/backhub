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

var (
	// Core styles
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))             // green
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))             // red
	warningStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))            // yellow
	pendingStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("12"))            // blue
	infoStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("14"))            // cyan
	debugStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("250"))           // light grey
	detailStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("13"))            // purple
	streamStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))           // grey
	headerStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("69")) // purple

	// Additional config
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

// ======================================== =================
// ======================================== Table Definitions
// ======================================== =================

type Table struct {
	Headers []string
	Rows    [][]string
	table   *table.Table
}

func NewTable(headers []string) *Table {
	t := &Table{
		Headers: headers,
		Rows:    [][]string{},
	}
	t.table = table.New().Headers(headers...)
	return t
}

func (t *Table) ReconcileRows() {
	if len(t.Rows) == 0 {
		return
	}
	for _, row := range t.Rows {
		t.table.Row(row...)
	}
}

func (t *Table) FormatTable(useMarkdown bool) string {
	t.ReconcileRows()
	if useMarkdown {
		return t.table.Border(lipgloss.MarkdownBorder()).String()
	}
	return t.table.String()
}

func (t *Table) PrintTable(useMarkdown bool) {
	os.Stdout.WriteString(t.FormatTable(useMarkdown))
}

func (t *Table) WriteMarkdownTableToFile(outputPath string) error {
	return os.WriteFile(outputPath, []byte(t.FormatTable(true)), 0644)
}

// =========================================== ==============
// =========================================== Output Manager
// =========================================== ==============

type FunctionOutput struct {
	Name        string
	Status      string
	Message     string
	StreamLines []string
	Complete    bool
	StartTime   time.Time
	LastUpdated time.Time
	Error       error
	Tables      map[string]*Table // Function tables
	Index       int
}

type ErrorReport struct {
	FunctionName string
	Error        error
	Time         time.Time
}

// Output manager main structure
type Manager struct {
	outputs         map[string]*FunctionOutput
	mutex           sync.RWMutex
	numLines        int
	maxStreams      int               // Max output stream lines per function
	unlimitedOutput bool              // When true, unlimited output per function
	tables          map[string]*Table // Global tables
	errors          []ErrorReport
	doneCh          chan struct{} // Channel to signal stopping the display
	pauseCh         chan bool     // Channel to pause/resume display updates
	isPaused        bool
	displayTick     time.Duration // Interval between display updates
	functionCount   int
	displayWg       sync.WaitGroup // WaitGroup for display goroutine shutdown
}

func NewManager(maxStreams int) *Manager {
	if maxStreams <= 0 {
		maxStreams = 15 // Default
	}
	return &Manager{
		outputs:         make(map[string]*FunctionOutput),
		tables:          make(map[string]*Table),
		errors:          []ErrorReport{},
		maxStreams:      maxStreams,
		unlimitedOutput: false,
		doneCh:          make(chan struct{}),
		pauseCh:         make(chan bool),
		isPaused:        false,
		displayTick:     200 * time.Millisecond, // Default
		functionCount:   0,
	}
}

func (m *Manager) SetUnlimitedOutput(unlimited bool) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.unlimitedOutput = unlimited
}

func (m *Manager) SetUpdateInterval(interval time.Duration) {
	m.displayTick = interval
}

func (m *Manager) Pause() {
	if !m.isPaused {
		m.pauseCh <- true
		m.isPaused = true
	}
}

func (m *Manager) Resume() {
	if m.isPaused {
		m.pauseCh <- false
		m.isPaused = false
	}
}

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

func (m *Manager) SetMessage(name, message string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if info, exists := m.outputs[name]; exists {
		info.Message = message
		info.LastUpdated = time.Now()
	}
}

func (m *Manager) SetStatus(name, status string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if info, exists := m.outputs[name]; exists {
		info.Status = status
		info.LastUpdated = time.Now()
	}
}

func (m *Manager) GetStatus(name string) string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	if info, exists := m.outputs[name]; exists {
		return info.Status
	}
	return "unknown"
}

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

func (m *Manager) ReportError(name string, err error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if info, exists := m.outputs[name]; exists {
		info.Complete = true
		info.Status = "error"
		info.Message = fmt.Sprintf("Error: %v", err)
		info.Error = err
		info.LastUpdated = time.Now()
		// Add to global error list
		m.errors = append(m.errors, ErrorReport{
			FunctionName: name,
			Error:        err,
			Time:         time.Now(),
		})
	}
}

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

func (m *Manager) AddStreamLine(name string, line string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if info, exists := m.outputs[name]; exists {
		if m.unlimitedOutput { // just append
			info.StreamLines = append(info.StreamLines, line)
		} else { // enforce size limit
			currentLen := len(info.StreamLines)
			if currentLen+1 > m.maxStreams {
				info.StreamLines = append(info.StreamLines[1:], line)
			} else {
				info.StreamLines = append(info.StreamLines, line)
			}
		}
		info.LastUpdated = time.Now()
	}
}

func (m *Manager) AddProgressBarToStream(name string, percentage float64, text string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if info, exists := m.outputs[name]; exists {
		percentage = max(0, min(percentage, 100))
		progressBar := PrintProgressBar(int(percentage), 100, 30)
		display := progressBar + debugStyle.Render(text)
		info.StreamLines = []string{display} // Set as only stream so nothing else is displayed
		info.LastUpdated = time.Now()
	}
}

func PrintProgressBar(current, total int, width int) string {
	if width <= 0 {
		width = 30
	}
	percent := float64(current) / float64(total)
	filled := min(int(percent*float64(width)), width)
	bar := "("
	bar += strings.Repeat(StyleSymbols["bullet"], filled)
	if filled < width {
		bar += ">"
		bar += strings.Repeat(" ", width-filled-1)
	}
	bar += ")"
	return debugStyle.Render(fmt.Sprintf("%s %.1f%% %s ", bar, percent*100, StyleSymbols["dot"]))
}

func (m *Manager) ClearLines(n int) {
	if n <= 0 {
		return
	}
	fmt.Printf("\033[%dA\033[J", n)
	m.numLines = max(m.numLines-n, 0)
}

func (m *Manager) ClearFunction(name string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if info, exists := m.outputs[name]; exists {
		info.StreamLines = []string{}
		info.Message = ""
		info.LastUpdated = time.Now()
	}
}

func (m *Manager) ClearAll() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	for name := range m.outputs {
		m.outputs[name].StreamLines = []string{}
		m.outputs[name].Message = ""
		m.outputs[name].LastUpdated = time.Now()
	}
}

func (m *Manager) GetStatusIndicator(status string) string {
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

// Add a global table
func (m *Manager) RegisterTable(name string, headers []string) *Table {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	table := NewTable(headers)
	m.tables[name] = table
	return table
}

// Adds a function-specific table
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

func (m *Manager) sortFunctions() (active, pending, completed []struct {
	name string
	info *FunctionOutput
}) {
	var allFuncs []struct {
		name  string
		info  *FunctionOutput
		index int
	}
	// Collect all functions
	for name, info := range m.outputs {
		allFuncs = append(allFuncs, struct {
			name  string
			info  *FunctionOutput
			index int
		}{name, info, info.Index})
	}
	// Sort by index (registration order)
	sort.Slice(allFuncs, func(i, j int) bool {
		return allFuncs[i].index < allFuncs[j].index
	})
	// Group functions by status
	for _, f := range allFuncs {
		if f.info.Complete {
			completed = append(completed, struct {
				name string
				info *FunctionOutput
			}{f.name, f.info})
		} else if f.info.Status == "pending" && f.info.Message == "" {
			pending = append(pending, struct {
				name string
				info *FunctionOutput
			}{f.name, f.info})
		} else {
			active = append(active, struct {
				name string
				info *FunctionOutput
			}{f.name, f.info})
		}
	}
	return active, pending, completed
}

func (m *Manager) updateDisplay() {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	if m.numLines > 0 && !m.unlimitedOutput {
		fmt.Printf("\033[%dA\033[J", m.numLines)
	}
	lineCount := 0
	activeFuncs, pendingFuncs, completedFuncs := m.sortFunctions()

	// Display active functions
	for idx, f := range activeFuncs {
		info := f.info
		statusDisplay := m.GetStatusIndicator(info.Status)
		elapsed := time.Since(info.StartTime).Round(time.Millisecond)
		elapsedStr := fmt.Sprintf("[%s]", elapsed)

		// Style the message based on status
		var styledMessage string
		var prefixStyle lipgloss.Style
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
		default: // pending or other
			styledMessage = pendingStyle.Render(info.Message)
			prefixStyle = pendingStyle
		}
		functionPrefix := strings.Repeat(" ", basePadding) + prefixStyle.Render(fmt.Sprintf("%d. ", idx+1))
		fmt.Printf("%s%s %s %s\n", functionPrefix, statusDisplay, debugStyle.Render(elapsedStr), styledMessage)
		lineCount++

		// Print stream lines with indentation
		if len(info.StreamLines) > 0 {
			indent := strings.Repeat(" ", basePadding+4) // Additional indentation for stream output
			for _, line := range info.StreamLines {
				fmt.Printf("%s%s\n", indent, streamStyle.Render(line))
				lineCount++
			}
		}
	}

	// Display pending functions
	for idx, f := range pendingFuncs {
		info := f.info
		statusDisplay := m.GetStatusIndicator(info.Status)
		functionPrefix := strings.Repeat(" ", basePadding) + pendingStyle.Render(fmt.Sprintf("%d. ", len(activeFuncs)+idx+1))
		fmt.Printf("%s%s %s\n", functionPrefix, statusDisplay, pendingStyle.Render("Waiting..."))
		lineCount++
		if len(info.StreamLines) > 0 {
			indent := strings.Repeat(" ", basePadding+4)
			for _, line := range info.StreamLines {
				fmt.Printf("%s%s\n", indent, streamStyle.Render(line))
				lineCount++
			}
		}
	}

	// Display completed functions
	for idx, f := range completedFuncs {
		info := f.info
		statusDisplay := m.GetStatusIndicator(info.Status)
		totalTime := info.LastUpdated.Sub(info.StartTime).Round(time.Millisecond)
		timeStr := fmt.Sprintf("[%s]", totalTime)
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
		functionPrefix := strings.Repeat(" ", basePadding) + prefixStyle.Render(fmt.Sprintf("%d. ", len(activeFuncs)+len(pendingFuncs)+idx+1))
		fmt.Printf("%s%s %s %s\n", functionPrefix, statusDisplay, debugStyle.Render(timeStr), styledMessage)
		lineCount++

		// Print stream lines with indentation if unlimited mode is enabled
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
				m.updateDisplay()
				m.ShowSummary()
				m.displayTables()
				return
			}
		}
	}()
}

func (m *Manager) StopDisplay() {
	close(m.doneCh)
	m.displayWg.Wait() // Wait for goroutine to finish
}

func (m *Manager) displayTables() {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	if len(m.tables) > 0 {
		fmt.Println(strings.Repeat(" ", basePadding) + headerStyle.Render("Global Tables:"))
		for name, table := range m.tables {
			fmt.Println(strings.Repeat(" ", basePadding+2) + headerStyle.Render(name))
			fmt.Println(table.FormatTable(false))
		}
	}
	// Display function tables
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
					fmt.Println(table.FormatTable(false))
				}
			}
		}
	}
}

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
		fmt.Printf("%s%s\n", strings.Repeat(" ", basePadding+4), errorStyle.Render(fmt.Sprintf("Error: %v", err.Error)))
	}
}

func (m *Manager) ShowSummary() {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	fmt.Println()
	var success, failures int
	for _, info := range m.outputs {
		if info.Status == "success" {
			success++
		} else if info.Status == "error" {
			failures++
		}
	}
	totalOps := fmt.Sprintf("Total Operations: %d", len(m.outputs))
	succeeded := fmt.Sprintf("Succeeded: %s", successStyle.Render(fmt.Sprintf("%d", success)))
	failed := fmt.Sprintf("Failed: %s", errorStyle.Render(fmt.Sprintf("%d", failures)))
	fmt.Println(infoStyle.Padding(0, basePadding).Render(fmt.Sprintf("%s, %s, %s", totalOps, succeeded, failed)))
	if m.unlimitedOutput {
		m.displayErrors()
	}
	fmt.Println()
}
