package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func TestTUIHang(t *testing.T) {
	// Set mock editor to 'true' so it exits immediately with success
	os.Setenv("EDITOR", "true")

	// Setup bubbletea logging to debug.log
	f, err := tea.LogToFile("debug.log", "debug")
	if err != nil {
		t.Fatalf("failed to setup log file: %v", err)
	}
	defer f.Close()

	// Initialize the model
	m := initialModel()

	pr, pw := io.Pipe()
	var outBuf bytes.Buffer

	p := tea.NewProgram(m, tea.WithInput(pr), tea.WithOutput(&outBuf))
	defer pw.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	errChan := make(chan error, 1)
	go func() {
		_, err := p.Run()
		errChan <- err
	}()

	// Wait for slides to load and initial renders
	time.Sleep(500 * time.Millisecond)

	// Send initial WindowSizeMsg to avoid 0x0 sizing panic
	p.Send(tea.WindowSizeMsg{Width: 80, Height: 24})
	time.Sleep(500 * time.Millisecond)

	// Send key strokes to navigate to the last slide
	// Let's press 'right' arrow key multiple times to make sure we reach the end
	fmt.Println("Test: Navigating to the end...")
	for i := 0; i < 5; i++ {
		pw.Write([]byte("\x1b[C")) // right arrow
		time.Sleep(200 * time.Millisecond)
	}

	// Test: Simulate editor finishing and slide reloading (which runs when Vim exits)
	fmt.Println("Test: Simulating editor finished...")
	p.Send(editorFinishedMsg{err: nil})
	time.Sleep(500 * time.Millisecond)

	// Now wait for a few seconds on the last slide, just like the user described
	fmt.Println("Test: Waiting on the last slide for 4 seconds...")
	time.Sleep(4 * time.Second)

	// Try navigating back
	fmt.Println("Test: Navigating back (left arrow)...")
	pw.Write([]byte("\x1b[D")) // left arrow
	time.Sleep(500 * time.Millisecond)

	// Quit the program
	fmt.Println("Test: Sending quit key 'q'...")
	pw.Write([]byte("q"))

	select {
	case err := <-errChan:
		if err != nil {
			t.Fatalf("Program exited with error: %v", err)
		}
		fmt.Println("Test: Finished successfully and responsive!")
	case <-ctx.Done():
		t.Fatal("Test: Program timed out! The application hung on the last slide!")
	}
}
