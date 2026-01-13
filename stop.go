package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

const pidFile = "server.pid"

func main() {
	fmt.Println("--- Stopping Server ---")

	// Read PID file
	data, err := ioutil.ReadFile(pidFile)
	if err != nil {
		fmt.Printf("Could not read %s. Is the server running?\nError: %v\n", pidFile, err)
		return
	}

	pidStr := strings.TrimSpace(string(data))
	pid, err := strconv.Atoi(pidStr)
	if err != nil {
		fmt.Printf("Invalid PID in file: %s\n", pidStr)
		return
	}

	fmt.Printf("Found Server PID: %d. Terminating...\n", pid)

	// Kill process (Windows)
	// taskkill /PID <pid> /F
	cmd := exec.Command("taskkill", "/PID", pidStr, "/F")
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Failed to stop process. It might have already exited.\nOutput: %s\nError: %v\n", string(output), err)
	} else {
		fmt.Println("Server stopped successfully.")
		fmt.Printf("Output: %s\n", string(output))
	}

	// Remove PID file
	err = os.Remove(pidFile)
	if err != nil {
		fmt.Printf("Warning: Could not delete %s: %v\n", pidFile, err)
	} else {
		fmt.Printf("Removed %s\n", pidFile)
	}

	fmt.Println("Done.")
	fmt.Println("Press Enter to exit...")
	fmt.Scanln()
}
