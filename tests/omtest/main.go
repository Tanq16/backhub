package main

import (
	"fmt"
	"math/rand"
	"os"
	"sync"
	"time"

	"github.com/tanq16/backhub/utils"
)

func main() {
	// Determine if we should run in unlimited output mode
	unlimitedOutput := false
	if len(os.Args) > 1 && os.Args[1] == "--unlimited" {
		unlimitedOutput = true
		fmt.Println("Running in unlimited output mode (--unlimited flag provided)")
	} else {
		fmt.Println("Running in limited output mode (use --unlimited flag for unlimited output)")
	}

	// Create output manager with 3 lines max per task
	outputMgr := utils.NewManager(3)
	outputMgr.SetUnlimitedOutput(unlimitedOutput)

	// Start the display routine
	outputMgr.StartDisplay()
	defer outputMgr.StopDisplay()

	// Register a main task
	outputMgr.Register("main")
	outputMgr.SetMessage("main", "Starting tests...")

	// Sleep briefly to let the display initialize
	time.Sleep(100 * time.Millisecond)

	// Table test
	tableTestName := "table-test"
	outputMgr.Register(tableTestName)
	outputMgr.SetMessage(tableTestName, "Testing table functionality")

	// Create a table
	table := outputMgr.RegisterTable("demo-table", []string{"ID", "Name", "Status", "Duration"})

	// Add some rows
	table.AddRow([]string{"1", "Repository backup", "Completed", "1.2s"})
	table.AddRow([]string{"2", "Database migration", "In Progress", "5.7s"})
	table.AddRow([]string{"3", "File cleanup", "Pending", "N/A"})
	table.AddRow([]string{"4", "User authentication", "Failed", "3.1s"})
	table.AddRow([]string{"5", "Long text test with wrapping behavior to demonstrate how the table handles long content that needs to be wrapped", "Testing", "10.2s"})

	// Pause the auto-display so we can show the table
	outputMgr.Pause()
	fmt.Println("\nTable Demo:")
	outputMgr.DisplayTable("demo-table", true)
	time.Sleep(2 * time.Second)
	outputMgr.Resume()

	outputMgr.SetMessage(tableTestName, "Table functionality test complete")
	outputMgr.Complete(tableTestName)

	// Start concurrent worker simulations with a wait group
	wg := sync.WaitGroup{}
	workerCount := 5

	// Launch workers
	for i := 1; i <= workerCount; i++ {
		wg.Add(1)
		go runWorker(outputMgr, i, &wg)
	}

	// Add a worker that will generate an error
	wg.Add(1)
	go runErrorWorker(outputMgr, &wg)

	// Update main status
	outputMgr.SetMessage("main", fmt.Sprintf("Running %d concurrent workers...", workerCount+1))

	// Wait for workers to complete
	wg.Wait()

	// Final status
	outputMgr.SetMessage("main", "All tests completed")
	outputMgr.Complete("main")

	// Sleep to allow viewing the final state
	time.Sleep(1 * time.Second)

	// Show summary of results
	outputMgr.ShowSummary()

	// Keep the summary visible
	time.Sleep(2 * time.Second)
}

// Simulates a worker that completes successfully
func runWorker(mgr *utils.Manager, id int, wg *sync.WaitGroup) {
	defer wg.Done()

	workerName := fmt.Sprintf("worker-%d", id)
	mgr.Register(workerName)
	mgr.SetMessage(workerName, fmt.Sprintf("Worker %d starting", id))

	// Simulate work steps
	stepsCount := 3 + rand.Intn(5) // 3-7 steps
	for i := 1; i <= stepsCount; i++ {
		// Random sleep to simulate varied processing time
		sleepTime := 300 + rand.Intn(900)
		time.Sleep(time.Duration(sleepTime) * time.Millisecond)

		// Log a message about this step
		mgr.AddStreamLine(workerName, fmt.Sprintf("Step %d/%d: Processing (waited %dms)",
			i, stepsCount, sleepTime))

		// Update status message occasionally
		if i == stepsCount/2 {
			mgr.SetMessage(workerName, fmt.Sprintf("Worker %d at 50%% progress", id))
		}
	}

	// Complete task
	mgr.SetMessage(workerName, fmt.Sprintf("Worker %d completed", id))
	mgr.Complete(workerName)
}

// Simulates a worker that encounters an error
func runErrorWorker(mgr *utils.Manager, wg *sync.WaitGroup) {
	defer wg.Done()

	workerName := "error-worker"
	mgr.Register(workerName)
	mgr.SetMessage(workerName, "Error worker starting")

	// Simulate some work
	for i := 1; i <= 2; i++ {
		time.Sleep(500 * time.Millisecond)
		mgr.AddStreamLine(workerName, fmt.Sprintf("Error worker step %d", i))
	}

	// Simulate error condition
	mgr.AddStreamLine(workerName, "Encountered a critical error...")
	time.Sleep(300 * time.Millisecond)
	mgr.ReportError(workerName, fmt.Errorf("simulated error: operation failed"))
}
