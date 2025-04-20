package main

import (
	"fmt"
	"math/rand"
	"os"
	"sync"
	"time"

	"github.com/tanq16/backhub/utils"
)

// Demo scenario: simulating a backup operation with multiple repositories
// Each repository will be processed in its own goroutine

func main() {
	// Check for unlimited output mode
	unlimitedOutput := false
	if len(os.Args) > 1 && os.Args[1] == "--unlimited" {
		unlimitedOutput = true
		fmt.Println("Running with unlimited output logging (--unlimited flag)")
	} else {
		fmt.Println("Running with limited output logging (max 3 lines per task)")
		fmt.Println("Use --unlimited flag for full logging")
	}
	fmt.Println()

	// Create the output manager
	outputMgr := utils.NewManager(3) // Max 3 lines per task by default
	outputMgr.SetUnlimitedOutput(unlimitedOutput)
	outputMgr.StartDisplay()
	defer outputMgr.StopDisplay()

	// Register the main coordination task
	outputMgr.Register("coordinator")
	outputMgr.SetMessage("coordinator", "Starting BackHub Visualization Demo")

	// Sample repositories to process
	repos := []struct {
		Name          string
		ShouldSucceed bool
		Size          string
		Duration      time.Duration
	}{
		{"github.com/user/small-repo", true, "15MB", 2 * time.Second},
		{"github.com/user/medium-repo", true, "120MB", 3 * time.Second},
		{"github.com/user/large-repo", true, "450MB", 4 * time.Second},
		{"github.com/org/private-repo", false, "85MB", 2 * time.Second},
		{"github.com/org/shared-repo", true, "250MB", 3 * time.Second},
	}

	// Create a table to track overall progress
	progressTable := outputMgr.RegisterTable("progress", []string{"Repository", "Status", "Size", "Duration"})

	// Initialize table with all repos as "Pending"
	for _, repo := range repos {
		progressTable.AddRow([]string{repo.Name, "Pending", repo.Size, "0s"})
	}

	// Show initial table
	outputMgr.Pause()
	fmt.Println("Initial state:")
	outputMgr.DisplayTable("progress", true)
	time.Sleep(2 * time.Second)
	outputMgr.Resume()

	// Start processing each repo in its own goroutine
	var wg sync.WaitGroup
	for i, repo := range repos {
		wg.Add(1)
		go processRepo(outputMgr, &wg, repo.Name, repo.ShouldSucceed, repo.Size, repo.Duration, i)
		time.Sleep(500 * time.Millisecond) // Stagger starts for better visualization
	}

	outputMgr.SetMessage("coordinator", "Processing 5 repositories...")

	// Wait for all processing to complete
	wg.Wait()

	// Show final status
	outputMgr.SetMessage("coordinator", "All repository processing has completed")
	outputMgr.Complete("coordinator")

	// Display final results table
	outputMgr.Pause()
	fmt.Println("\nFinal results:")
	outputMgr.DisplayTable("progress", true)
	time.Sleep(2 * time.Second)
	outputMgr.Resume()

	// Show summary
	time.Sleep(500 * time.Millisecond)
	outputMgr.ShowSummary()
	time.Sleep(3 * time.Second)
}

func processRepo(mgr *utils.Manager, wg *sync.WaitGroup, repoName string, shouldSucceed bool, size string, duration time.Duration, index int) {
	defer wg.Done()

	// Create a unique task name
	taskName := fmt.Sprintf("repo-%d", index)

	// Register with output manager
	mgr.Register(taskName)
	mgr.SetMessage(taskName, fmt.Sprintf("Processing %s", repoName))

	// Get a reference to the progress table
	progressTable := mgr.GetTable("progress")

	// Update table status to "In Progress"
	updateTableRow(progressTable, index, "In Progress", size, "0s")

	// Simulate the git clone/update operation
	steps := []string{
		"Connecting to GitHub...",
		fmt.Sprintf("Repository size: %s", size),
		"Checking remote refs...",
		"Downloading objects...",
		"Resolving deltas...",
		"Checking out files...",
	}

	// Process each step
	startTime := time.Now()
	stepDuration := duration / time.Duration(len(steps))

	for i, step := range steps {
		// Simulate some work
		time.Sleep(stepDuration)

		// Log progress
		progress := float64(i+1) / float64(len(steps)) * 100
		mgr.AddStreamLine(taskName, fmt.Sprintf("%s (%.0f%%)", step, progress))

		// Update table with current duration
		elapsed := time.Since(startTime).Round(100 * time.Millisecond)
		updateTableRow(progressTable, index, "In Progress", size, elapsed.String())

		// Add some randomized messages for more realistic output
		if rand.Float32() > 0.7 {
			extraMsg := getRandomProgressMessage()
			mgr.AddStreamLine(taskName, extraMsg)
		}
	}

	// Complete with success or error
	elapsed := time.Since(startTime).Round(100 * time.Millisecond)
	if shouldSucceed {
		mgr.SetMessage(taskName, fmt.Sprintf("Successfully processed %s", repoName))
		mgr.Complete(taskName)
		updateTableRow(progressTable, index, "Success", size, elapsed.String())
	} else {
		err := fmt.Errorf("failed to authenticate for repository %s", repoName)
		mgr.ReportError(taskName, err)
		updateTableRow(progressTable, index, "Failed", size, elapsed.String())
	}
}

// Helper function to update a specific row in the table
func updateTableRow(table *utils.Table, rowIndex int, status, size, duration string) {
	// This is a simplified version - in a real implementation you'd
	// need to handle table locks if multiple goroutines update it
	if rowIndex < len(table.Rows) {
		table.Rows[rowIndex][1] = status
		table.Rows[rowIndex][3] = duration
	}
}

// Returns random progress messages for variety
func getRandomProgressMessage() string {
	messages := []string{
		"Compressing objects...",
		"Network connection slower than expected",
		"Pruning redundant objects",
		"Optimizing local storage",
		"Received large pack file",
		"Processing reference changes",
		"Updating working copy",
		"Analyzing remote changes",
		"Verifying connectivity",
	}
	return messages[rand.Intn(len(messages))]
}
