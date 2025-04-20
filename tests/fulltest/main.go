package main

import (
	"errors"
	"flag"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/tanq16/backhub/utils"
)

// Example demonstrating the updated output manager functionality
func main() {
	// Parse command line flags
	debugMode := flag.Bool("debug", false, "Enable debug mode with unlimited output")
	flag.Parse()

	// Create output manager with up to 15 lines per function
	outputMgr := utils.NewManager(15)
	outputMgr.SetUnlimitedOutput(*debugMode)
	outputMgr.StartDisplay()
	defer outputMgr.StopDisplay()

	// Register main coordinator function
	outputMgr.Register("coordinator")
	outputMgr.SetMessage("coordinator", "Starting BackHub demo with updated output manager")

	// Create a global statistics table
	statsTable := outputMgr.RegisterTable("Statistics", []string{"Operation", "Count", "Status"})
	statsTable.AddRow([]string{"Repositories", "5", "Pending"})
	statsTable.AddRow([]string{"Files", "0", "Pending"})
	statsTable.AddRow([]string{"Bytes", "0", "Pending"})

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

	// Process repositories concurrently
	var wg sync.WaitGroup
	totalBytes := 0
	totalFiles := 0

	// Add our progress reporter
	wg.Add(1)
	go runProgressReporter(outputMgr, &wg)

	for i, repo := range repos {
		wg.Add(1)
		// Start each repository processing in its own goroutine
		go func(idx int, repo struct {
			Name          string
			ShouldSucceed bool
			Size          string
			Duration      time.Duration
		}) {
			defer wg.Done()

			// Create a unique task name
			taskName := fmt.Sprintf("repo-%d", idx)

			// Register with output manager
			outputMgr.Register(taskName)
			outputMgr.SetMessage(taskName, fmt.Sprintf("Processing %s", repo.Name))

			// Create a function-specific table
			repoTable := outputMgr.RegisterFunctionTable(taskName, fmt.Sprintf("%s-details", repo.Name),
				[]string{"Operation", "Status", "Duration"})

			// Simulate processing steps
			steps := []string{
				"Connecting to GitHub...",
				fmt.Sprintf("Repository size: %s", repo.Size),
				"Checking remote refs...",
				"Downloading objects...",
				"Resolving deltas...",
				"Checking out files...",
			}

			// Process each step
			bytesProcessed := 0
			filesProcessed := 0
			startTime := time.Now()
			stepDuration := repo.Duration / time.Duration(len(steps))

			for i, step := range steps {
				// Simulate work
				time.Sleep(stepDuration)

				// Process random data for statistics
				stepBytes := rand.Intn(10000) + 1000
				stepFiles := rand.Intn(5) + 1
				bytesProcessed += stepBytes
				filesProcessed += stepFiles

				// Update global statistics (in a real app, use mutex for this)
				totalBytes += stepBytes
				totalFiles += stepFiles
				statsTable.Rows[1][1] = fmt.Sprintf("%d", totalFiles)
				statsTable.Rows[2][1] = fmt.Sprintf("%d KB", totalBytes/1024)

				// Log progress
				progress := float64(i+1) / float64(len(steps)) * 100
				outputMgr.AddStreamLine(taskName, fmt.Sprintf("%s (%.0f%%)", step, progress))

				// Add to repo table
				elapsed := time.Since(startTime).Round(100 * time.Millisecond)
				repoTable.AddRow([]string{step, "Complete", elapsed.String()})

				// Add some randomized messages for more realistic output
				if rand.Float32() > 0.7 {
					extraMsg := getRandomProgressMessage()
					outputMgr.AddStreamLine(taskName, extraMsg)
				}
			}

			// Complete with success or error
			if repo.ShouldSucceed {
				outputMgr.SetMessage(taskName, fmt.Sprintf("Successfully processed %s", repo.Name))
				outputMgr.Complete(taskName)
			} else {
				err := errors.New("authentication failed: invalid token")
				outputMgr.ReportError(taskName, err)
			}

		}(i, repo)

		// Stagger starts for better visualization
		time.Sleep(500 * time.Millisecond)
	}

	// Update coordinator status
	outputMgr.SetMessage("coordinator", "Processing 5 repositories and running progress demo...")

	// Wait for all processing to complete
	wg.Wait()

	// Final update to the stats table
	statsTable.Rows[0][2] = "Complete"
	statsTable.Rows[1][2] = "Complete"
	statsTable.Rows[2][2] = "Complete"

	// Show completion message
	outputMgr.SetMessage("coordinator", "All repository processing has completed")
	outputMgr.Complete("coordinator")

	// Summary will be shown automatically when StopDisplay is called
	time.Sleep(500 * time.Millisecond)
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

// runProgressReporter simulates a task with a progress bar
func runProgressReporter(mgr *utils.Manager, wg *sync.WaitGroup) {
	defer wg.Done()

	progressName := "progress-reporter"
	mgr.Register(progressName)
	mgr.SetMessage(progressName, "Running task with progress reporting")

	// Total steps for our task
	totalSteps := 20

	// Simulate work with progress updates
	for step := 1; step <= totalSteps; step++ {
		// Calculate percentage
		percentage := float64(step) / float64(totalSteps) * 100

		// Sleep to simulate work
		sleepTime := 200 + rand.Intn(300)
		time.Sleep(time.Duration(sleepTime) * time.Millisecond)

		// Add progress message
		status := fmt.Sprintf("Step %d of %d", step, totalSteps)
		mgr.AddProgressBar(progressName, percentage, status)

		// Update main message occasionally
		if step == totalSteps/4 {
			mgr.SetMessage(progressName, "Progress reporter at 25%")
		} else if step == totalSteps/2 {
			mgr.SetMessage(progressName, "Progress reporter at 50%")
		} else if step == totalSteps*3/4 {
			mgr.SetMessage(progressName, "Progress reporter at 75%")
		}
	}

	// Complete the task
	mgr.SetMessage(progressName, "Progress reporter completed")
	mgr.Complete(progressName)
}
