package output

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

type FunctionOutput struct {
	Name        string
	Status      string
	Message     string
	StreamLines []string
	Complete    bool
	StartTime   time.Time
	LastUpdated time.Time
}

// Output manager
type Manager struct {
	outputs     map[string]*FunctionOutput
	mutex       sync.RWMutex
	numLines    int
	maxStreams  int               // Max stream lines per function
	tables      map[string]*Table // Tables that can be displayed
	doneCh      chan struct{}     // Channel to signal stopping the display
	displayTick time.Duration     // Interval between display updates
}

// Creates a new output manager
func NewManager(maxStreams int) *Manager {
	if maxStreams <= 0 {
		maxStreams = 5 // Default value
	}
	return &Manager{
		outputs:     make(map[string]*FunctionOutput),
		tables:      make(map[string]*Table),
		maxStreams:  maxStreams,
		doneCh:      make(chan struct{}),
		displayTick: 200 * time.Millisecond, // 200ms default update interval
	}
}

// Adds a new function to be tracked
func (m *Manager) Register(name string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.outputs[name] = &FunctionOutput{
		Name:        name,
		Status:      "pending",
		StreamLines: []string{},
		StartTime:   time.Now(),
		LastUpdated: time.Now(),
	}
}

// Sets the primary message for a function
func (m *Manager) SetMessage(name, message string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if info, exists := m.outputs[name]; exists {
		info.Message = message
		info.LastUpdated = time.Now()
	}
}

// Updates the status of a function
func (m *Manager) SetStatus(name, status string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if info, exists := m.outputs[name]; exists {
		info.Status = status
		info.LastUpdated = time.Now()
	}
}

// Retrieves the status of a function
func (m *Manager) GetStatus(name string) string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	if info, exists := m.outputs[name]; exists {
		return info.Status
	}
	return "unknown"
}

// Marks a function as complete
func (m *Manager) Complete(name string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if info, exists := m.outputs[name]; exists {
		info.Complete = true
		info.Status = "success"
		info.LastUpdated = time.Now()
	}
}

// Sets a function's status to error
func (m *Manager) ReportError(name string, err error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if info, exists := m.outputs[name]; exists {
		info.Complete = true
		info.Status = "error"
		info.Message = fmt.Sprintf("Error: %v", err)
		info.LastUpdated = time.Now()
	}
}

// Adds lines to a function's stream output
func (m *Manager) UpdateStreamOutput(name string, output []string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if info, exists := m.outputs[name]; exists {
		currentLen := len(info.StreamLines)
		if currentLen+len(output) > m.maxStreams {
			// Keep only the most recent lines up to maxStreams
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
		info.LastUpdated = time.Now()
	}
}

// Adds a single line to a function's stream output
func (m *Manager) AddStreamLine(name, line string) {
	m.UpdateStreamOutput(name, []string{line})
}

// Clear the entire screen
func (m *Manager) ClearOutput() {
	fmt.Print("\033[H\033[2J")
	m.numLines = 0
}

// Clears 'n' previous lines
func (m *Manager) ClearLines(n int) {
	if n <= 0 {
		return
	}
	fmt.Printf("\033[%dA\033[J", n)
	m.numLines = max(m.numLines-n, 0)
}

// Clears the output of a specific function
func (m *Manager) ClearFunction(name string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if info, exists := m.outputs[name]; exists {
		info.StreamLines = []string{}
		info.Message = ""
		info.Status = "pending"
		info.Complete = false
		info.LastUpdated = time.Now()
	}
}

// Returns a colored status indicator based on status
func (m *Manager) GetStatusDisplay(status string) string {
	switch status {
	case "success":
		return fmt.Sprintf("%s[%s]%s", Colors["green"], Colors["pass"], Colors["reset"])
	case "error":
		return fmt.Sprintf("%s[%s]%s", Colors["red"], Colors["fail"], Colors["reset"])
	case "warning":
		return fmt.Sprintf("%s[!]%s", Colors["yellow"], Colors["reset"])
	case "pending":
		return fmt.Sprintf("%s[…]%s", Colors["blue"], Colors["reset"])
	default:
		return fmt.Sprintf("%s[·]%s", Colors["grey"], Colors["reset"])
	}
}

// Adds a table to the manager
func (m *Manager) RegisterTable(name string, headers []string) *Table {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	table := NewTable(headers)
	m.tables[name] = table
	return table
}

// Retrieves a table by name
func (m *Manager) GetTable(name string) *Table {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.tables[name]
}

// Displays a specific table
func (m *Manager) DisplayTable(name string, innerDividers bool) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	if table, exists := m.tables[name]; exists {
		if m.numLines > 0 {
			fmt.Printf("\033[%dA\033[J", m.numLines)
		}
		tableStr := table.FormatTable(innerDividers)
		fmt.Print(tableStr)
		m.numLines = strings.Count(tableStr, "\n")
	}
}

// Removes a table from the manager
func (m *Manager) RemoveTable(name string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	delete(m.tables, name)
}

// Updates the console display with all function outputs
func (m *Manager) updateDisplay() {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	if m.numLines > 0 {
		fmt.Printf("\033[%dA\033[J", m.numLines)
	}
	lineCount := 0
	var keys []string
	for k := range m.outputs {
		keys = append(keys, k)
	}

	// First active functions, then completed, then pending in sorted order
	var activeKeys []string
	var completedKeys []string
	var pendingKeys []string
	for _, k := range keys {
		info := m.outputs[k]
		if info.Complete {
			completedKeys = append(completedKeys, k)
		} else if info.Status == "pending" && info.Message == "" {
			pendingKeys = append(pendingKeys, k)
		} else {
			activeKeys = append(activeKeys, k)
		}
	}

	// Print active functions
	for _, name := range activeKeys {
		info := m.outputs[name]
		statusDisplay := m.GetStatusDisplay(info.Status)
		// Calculate elapsed time
		elapsed := time.Since(info.StartTime).Round(time.Millisecond)
		elapsedStr := ""
		if elapsed > time.Second {
			elapsedStr = fmt.Sprintf(" [%s]", elapsed)
		}
		fmt.Printf("%s %s%s: %s\n", statusDisplay, name, elapsedStr, info.Message)
		lineCount++
		if len(info.StreamLines) > 0 {
			for _, line := range info.StreamLines {
				fmt.Printf("\t%s→ %s%s\n", Colors["green"], line, Colors["reset"])
				lineCount++
			}
		}
	}

	// Print pending functions
	for _, name := range pendingKeys {
		info := m.outputs[name]
		statusDisplay := m.GetStatusDisplay(info.Status)
		fmt.Printf("%s %s: Waiting...\n", statusDisplay, name)
		lineCount++
		if len(info.StreamLines) > 0 {
			for _, line := range info.StreamLines {
				fmt.Printf("\t%s→ %s%s\n", Colors["green"], line, Colors["reset"])
				lineCount++
			}
		}
	}

	// Print completed functions
	for _, name := range completedKeys {
		info := m.outputs[name]
		statusDisplay := m.GetStatusDisplay(info.Status)
		// Calculate total time
		totalTime := info.LastUpdated.Sub(info.StartTime).Round(time.Millisecond)
		timeStr := ""
		if totalTime > time.Millisecond {
			timeStr = fmt.Sprintf(" [completed in %s]", totalTime)
		}
		fmt.Printf("%s %s%s: %s\n", statusDisplay, name, timeStr, info.Message)
		lineCount++
		if len(info.StreamLines) > 0 {
			for _, line := range info.StreamLines {
				fmt.Printf("\t%s→ %s%s\n", Colors["green"], line, Colors["reset"])
				lineCount++
			}
		}
	}
	m.numLines = lineCount
}

// Starts the automatic display update goroutine
func (m *Manager) StartDisplay() {
	go func() {
		ticker := time.NewTicker(m.displayTick)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				m.updateDisplay()
			case <-m.doneCh:
				m.updateDisplay()
				return
			}
		}
	}()
}

// Stops the automatic display updates
func (m *Manager) StopDisplay() {
	close(m.doneCh)
}

// Sets the interval between display updates
func (m *Manager) SetUpdateInterval(interval time.Duration) {
	m.displayTick = interval
}

// Displays a final summary of all functions
func (m *Manager) ShowSummary() {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	m.updateDisplay()
	fmt.Println()

	// Count success and failures
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
	// Calculate average time if there are completed functions
	avgTime := time.Duration(0)
	if success+failures > 0 {
		avgTime = totalTime / time.Duration(success+failures)
	}
	fmt.Printf("%sTotal Operations: %d, Succeeded: %d, Failed: %d, Avg Time: %s%s\n",
		Colors["blue"], len(m.outputs), success, failures,
		avgTime.Round(time.Second), Colors["reset"])
}

// Removes a function from the manager
func (m *Manager) Remove(name string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	delete(m.outputs, name)
}

// Removes completed functions from the manager
func (m *Manager) RemoveCompleted() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	for name, info := range m.outputs {
		if info.Complete {
			delete(m.outputs, name)
		}
	}
}

// Shows outputs of all current functions (for a manual update)
func (m *Manager) Display() {
	m.updateDisplay()
}
