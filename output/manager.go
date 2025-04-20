package output

import (
	"fmt"
	"strings"
	"sync"
)

// Output of a function
type FunctionOutput struct {
	Name        string
	Status      string
	Message     string
	StreamLines []string
	Complete    bool
}

// Output manager
type Manager struct {
	outputs    map[string]*FunctionOutput
	mutex      sync.RWMutex
	numLines   int
	maxStreams int               // Max stream lines per function
	tables     map[string]*Table // Tables that can be displayed
}

// Creates a new output manager
func NewManager(maxStreams int) *Manager {
	if maxStreams <= 0 {
		maxStreams = 5 // Default value
	}
	return &Manager{
		outputs:    make(map[string]*FunctionOutput),
		tables:     make(map[string]*Table),
		maxStreams: maxStreams,
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
	}
}

// Sets the primary message for a function
func (m *Manager) SetMessage(name, message string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if info, exists := m.outputs[name]; exists {
		info.Message = message
	}
}

// Updates the status of a function
func (m *Manager) SetStatus(name, status string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if info, exists := m.outputs[name]; exists {
		info.Status = status
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

// Shows outputs of all current functions
func (m *Manager) Display() {
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
	// Print active functions
	for _, name := range keys {
		info := m.outputs[name]
		// Skip completed functions with no streams if desired
		// if info.Complete && len(info.StreamLines) == 0 {
		//    continue
		// }
		statusDisplay := m.GetStatusDisplay(info.Status)
		fmt.Printf("%s %s: %s\n", statusDisplay, name, info.Message)
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
