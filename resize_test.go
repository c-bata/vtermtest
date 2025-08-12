package vtermtest_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/c-bata/vtermtest"
	"github.com/c-bata/vtermtest/keys"
)

func TestResize(t *testing.T) {
	ctx := context.Background()

	emu := vtermtest.New(5, 20).
		Command("sh", "-c", "while true; do echo '==========1234567890=========='; sleep 0.1; done").
		Env("LANG=C.UTF-8")

	if err := emu.Start(ctx); err != nil {
		t.Fatalf("failed to start: %v", err)
	}
	defer emu.Close()

	// Wait for initial output
	time.Sleep(200 * time.Millisecond)
	emu.WaitStable(100*time.Millisecond, 1*time.Second)

	// Get initial screen (20 columns wide)
	screen, err := emu.GetScreenText()
	if err != nil {
		t.Fatalf("failed to get screen: %v", err)
	}

	// Each line should be truncated to 20 chars
	lines := strings.Split(screen, "\n")
	for i, line := range lines {
		if len(line) > 20 {
			t.Errorf("Line %d is longer than 20 chars: %q (len=%d)", i, line, len(line))
		}
	}

	// Resize to wider terminal
	if err := emu.Resize(5, 40); err != nil {
		t.Fatalf("failed to resize: %v", err)
	}

	// Wait for new output after resize
	time.Sleep(200 * time.Millisecond)
	emu.WaitStable(100*time.Millisecond, 1*time.Second)

	// Get screen after resize
	screen, err = emu.GetScreenText()
	if err != nil {
		t.Fatalf("failed to get screen after resize: %v", err)
	}

	// Now lines can be up to 40 chars
	lines = strings.Split(screen, "\n")
	hasLongLine := false
	for _, line := range lines {
		if len(line) > 20 {
			hasLongLine = true
			break
		}
	}

	if !hasLongLine {
		t.Error("After resize to 40 columns, expected longer lines")
	}
}

func TestResizeInteractive(t *testing.T) {
	ctx := context.Background()

	emu := vtermtest.New(10, 40).
		Command("sh").
		Env("LANG=C.UTF-8")

	if err := emu.Start(ctx); err != nil {
		t.Fatalf("failed to start: %v", err)
	}
	defer emu.Close()

	// Send a command
	if err := emu.KeyPress(keys.Text("echo 'Short'"), keys.Enter); err != nil {
		t.Fatal(err)
	}

	emu.AssertScreenContains(t, "Short")

	// Resize smaller
	if err := emu.Resize(5, 20); err != nil {
		t.Fatalf("failed to resize: %v", err)
	}

	// Send another command after resize
	if err := emu.KeyPress(keys.Text("echo 'After resize'"), keys.Enter); err != nil {
		t.Fatal(err)
	}

	emu.AssertScreenContains(t, "After resize")
}