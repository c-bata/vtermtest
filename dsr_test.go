package vtermtest

import (
	"context"
	"testing"
	"time"

	"github.com/c-bata/vtermtest/keys"
)

func TestGetCursorPosition(t *testing.T) {
	emu := New(24, 80).Command("bash", "-c", "echo -n 'Hello World'")
	t.Cleanup(func() { _ = emu.Close() })

	if err := emu.Start(context.Background()); err != nil {
		t.Fatalf("start: %v", err)
	}

	// Wait for the command to output text
	if !emu.WaitStable(100*time.Millisecond, 2*time.Second) {
		t.Fatalf("screen did not stabilize")
	}

	// Get cursor position after "Hello World" is printed
	row, col, err := emu.GetCursorPosition()
	if err != nil {
		t.Fatalf("GetCursorPosition failed: %v", err)
	}

	// After "Hello World" (11 chars), cursor should be at column 12 (1-based)
	// Row depends on the prompt, but should be at least 1
	if row < 1 {
		t.Errorf("Expected row >= 1, got %d", row)
	}
	if col < 12 {
		t.Errorf("Expected col >= 12 (after 'Hello World'), got %d", col)
	}

	t.Logf("Cursor position: row=%d, col=%d", row, col)
}

func TestGetCursorPositionAfterMovement(t *testing.T) {
	emu := New(24, 80).Command("bash", "-c", "stty raw -echo; cat")
	t.Cleanup(func() { _ = emu.Close() })

	if err := emu.Start(context.Background()); err != nil {
		t.Fatalf("start: %v", err)
	}

	// Wait for initialization
	time.Sleep(100 * time.Millisecond)

	// Type some text
	if err := emu.KeyPress(keys.Text("test")); err != nil {
		t.Fatalf("send text: %v", err)
	}

	// Wait for text to appear
	time.Sleep(50 * time.Millisecond)

	// Get position after typing "test"
	row1, col1, err := emu.GetCursorPosition()
	if err != nil {
		t.Fatalf("GetCursorPosition after text: %v", err)
	}
	t.Logf("Position after 'test': row=%d, col=%d", row1, col1)

	// Move cursor left twice
	if err := emu.KeyPress(keys.Left, keys.Left); err != nil {
		t.Fatalf("send left keys: %v", err)
	}

	// Wait for cursor movement
	time.Sleep(50 * time.Millisecond)

	// Get position after moving left
	row2, col2, err := emu.GetCursorPosition()
	if err != nil {
		t.Fatalf("GetCursorPosition after movement: %v", err)
	}
	t.Logf("Position after Left x2: row=%d, col=%d", row2, col2)

	// Cursor should have moved left by 2 columns
	if row1 != row2 {
		t.Errorf("Row changed unexpectedly: %d -> %d", row1, row2)
	}
	if col2 != col1-2 {
		t.Errorf("Expected col to decrease by 2 (from %d to %d), got: %d", col1, col1-2, col2)
	}
}

func TestDSRSequenceInKeys(t *testing.T) {
	// Test that DSR sequence is correctly defined
	expected := []byte{0x1B, 0x5B, 0x36, 0x6E} // ESC[6n
	if string(keys.DSR) != string(expected) {
		t.Errorf("DSR sequence mismatch. Got %v, want %v", keys.DSR, expected)
	}
}