package vtermtest_test

import (
	"context"
	"testing"
	"time"

	"github.com/c-bata/vtermtest"
	"github.com/c-bata/vtermtest/keys"
)

func TestBasicEmulator(t *testing.T) {
	ctx := context.Background()
	
	emu := vtermtest.New(10, 80).
		Command("go", "run", "./_examples/simple_example.go").
		Env("LANG=C.UTF-8", "TERM=xterm")
	
	if err := emu.Start(ctx); err != nil {
		t.Fatalf("failed to start emulator: %v", err)
	}
	defer emu.Close()

	// Wait for initial prompt to appear
	time.Sleep(500 * time.Millisecond)
	if !emu.WaitStable(100*time.Millisecond, 2*time.Second) {
		t.Fatal("initial prompt did not appear")
	}

	// Get initial screen
	screen, err := emu.GetScreenText()
	if err != nil {
		t.Fatalf("failed to get screen: %v", err)
	}
	t.Logf("Initial screen:\n%s", screen)

	// Send name
	if err := emu.KeyPress(keys.Text("Alice"), keys.Enter); err != nil {
		t.Fatalf("failed to send name: %v", err)
	}

	// Wait for response
	if !emu.WaitStable(50*time.Millisecond, 2*time.Second) {
		t.Fatal("response did not stabilize")
	}

	// Get screen after name input
	screen, err = emu.GetScreenText()
	if err != nil {
		t.Fatalf("failed to get screen: %v", err)
	}
	t.Logf("After name input:\n%s", screen)

	// Send number
	if err := emu.KeyPress(keys.Text("42"), keys.Enter); err != nil {
		t.Fatalf("failed to send number: %v", err)
	}

	// Wait for final output
	if !emu.WaitStable(50*time.Millisecond, 2*time.Second) {
		t.Fatal("final output did not stabilize")
	}

	// Get final screen
	screen, err = emu.GetScreenText()
	if err != nil {
		t.Fatalf("failed to get final screen: %v", err)
	}
	t.Logf("Final screen:\n%s", screen)

	// Basic assertions
	if !contains(screen, "Hello, Alice!") {
		t.Error("Expected to see 'Hello, Alice!' in output")
	}
	if !contains(screen, "You entered: 42") {
		t.Error("Expected to see 'You entered: 42' in output")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[len(s)-len(substr):] == substr || 
		   len(s) >= len(substr) && s[:len(substr)] == substr ||
		   len(s) >= len(substr) && containsMiddle(s, substr)
}

func containsMiddle(s, substr string) bool {
	for i := 1; i < len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}