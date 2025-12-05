package builtin

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"strings"
)

// StreamCallback is called for each line of output during command execution
type StreamCallback func(requestID, commandID, line string)

// ExecuteWithStreaming executes a command and streams output line by line
// The cmd should already be configured with all arguments before calling this function
func ExecuteWithStreaming(cmd *exec.Cmd, requestID, commandID string, callback StreamCallback) (string, error) {
	// Create pipes
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", fmt.Errorf("failed to create stdout pipe: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return "", fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		return "", err
	}

	// Use a channel to receive output lines
	outputChan := make(chan string)
	doneChan := make(chan bool)

	// Helper to read from pipe to channel
	readPipe := func(r io.Reader) {
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			outputChan <- scanner.Text()
		}
		doneChan <- true
	}

	go readPipe(stdout)
	go readPipe(stderr)

	// Close channel when both readers are done
	go func() {
		<-doneChan
		<-doneChan
		close(outputChan)
	}()

	var fullOutput strings.Builder

	// Stream each line as it comes
	for line := range outputChan {
		fullOutput.WriteString(line + "\n")

		// Send line to callback for real-time streaming
		if callback != nil {
			callback(requestID, commandID, line)
		}
	}

	// Wait for command to finish
	err = cmd.Wait()
	return fullOutput.String(), err
}
