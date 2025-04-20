package main

import (
	"fmt"
	"os"
	"time"

	"github.com/tanq16/backhub/utils"
)

func main() {
	fmt.Println("Table Formatting Test")
	fmt.Println("====================")

	// Create a new table with headers
	headers := []string{"Repository", "Status", "Last Update", "Size", "Branch Count"}
	table := utils.NewTable(headers)

	// Add some sample data
	table.AddRow([]string{"github.com/user/repo1", "Success", "2025-04-20", "25MB", "3"})
	table.AddRow([]string{"github.com/user/large-project-with-a-very-long-name", "In Progress", "2025-04-19", "1.2GB", "12"})
	table.AddRow([]string{"github.com/org/repo3", "Failed", "2025-04-18", "756KB", "5"})
	table.AddRow([]string{"github.com/org/repo4", "Success", "2025-04-17", "128MB", "8"})

	// Add a row with very long content to test wrapping
	table.AddRow([]string{
		"github.com/org/repo-with-long-descriptive-name",
		"Completed with warnings: Some files couldn't be accessed due to permission issues. Check system logs for details.",
		"2025-04-16",
		"2.4GB",
		"15",
	})

	// Print the table with inner dividers
	fmt.Println("\nTable with inner dividers:")
	table.PrintTable(true)

	// Print the table without inner dividers
	fmt.Println("\nTable without inner dividers:")
	table.PrintTable(false)

	// Print as markdown
	fmt.Println("\nTable in Markdown format:")
	table.PrintMarkdownTable()

	// Save to file
	mdFile := "table_output.md"
	err := table.WriteMarkdownTableToFile(mdFile)
	if err != nil {
		fmt.Printf("Error writing markdown table to file: %v\n", err)
	} else {
		fmt.Printf("Markdown table written to %s\n", mdFile)
	}

	// Now demonstrate a table in the output manager
	fmt.Println("\nTable in Output Manager:")
	fmt.Println("=======================")

	// Create output manager
	mgr := utils.NewManager(5)
	mgr.SetUnlimitedOutput(true)
	mgr.StartDisplay()

	// Register a task
	mgr.Register("table-demo")
	mgr.SetMessage("table-demo", "Table demo task")

	// Create a table in the manager
	mgrTable := mgr.RegisterTable("status-table", []string{"Repository", "Status", "Time", "Success"})

	// Add rows
	mgrTable.AddRow([]string{"repo1", "Cloning", "2.5s", "Yes"})
	mgrTable.AddRow([]string{"repo2", "Updating", "1.3s", "Yes"})
	mgrTable.AddRow([]string{"repo3", "Failed", "4.2s", "No"})

	// Pause auto display to show the table
	mgr.Pause()
	fmt.Println() // Add a blank line
	mgr.DisplayTable("status-table", true)

	// Wait a moment
	time.Sleep(3 * time.Second)

	// Cleanup
	mgr.StopDisplay()

	// Try to remove the markdown file
	time.Sleep(500 * time.Millisecond)
	err = os.Remove(mdFile)
	if err != nil {
		fmt.Printf("Note: Could not remove %s: %v\n", mdFile, err)
	}
}
