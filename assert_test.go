package vtermtest_test

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/c-bata/vtermtest"
)

func TestAssertions(t *testing.T) {
	ctx := context.Background()

	t.Run("AssertLineEqual", func(t *testing.T) {
		emu := vtermtest.New(5, 40).
			Command("echo", "line1\nline2\nline3").
			Env("LANG=C.UTF-8")

		if err := emu.Start(ctx); err != nil {
			t.Fatalf("failed to start: %v", err)
		}
		defer emu.Close()

		time.Sleep(100 * time.Millisecond)

		// Test passes
		emu.AssertLineEqual(t, 0, "line1")
		emu.AssertLineEqual(t, 1, "line2")
		emu.AssertLineEqual(t, 2, "line3")
	})

	t.Run("AssertScreenEqual", func(t *testing.T) {
		emu := vtermtest.New(4, 30).
			Command("echo", "Hello\nWorld").
			Env("LANG=C.UTF-8")

		if err := emu.Start(ctx); err != nil {
			t.Fatalf("failed to start: %v", err)
		}
		defer emu.Close()

		emu.AssertScreenEqual(t, `
Hello
World
`)
	})

	t.Run("AssertScreenContains", func(t *testing.T) {
		emu := vtermtest.New(5, 40).
			Command("sh", "-c", "echo 'The quick brown fox'").
			Env("LANG=C.UTF-8")

		if err := emu.Start(ctx); err != nil {
			t.Fatalf("failed to start: %v", err)
		}
		defer emu.Close()

		emu.AssertScreenContains(t, "quick brown")
		emu.AssertScreenContains(t, "fox")
	})
}

func TestAssertRetry(t *testing.T) {
	ctx := context.Background()

	// Test that assertions retry and eventually succeed
	emu := vtermtest.New(5, 40).
		Command("sh", "-c", "sleep 0.1 && echo 'delayed output'").
		Env("LANG=C.UTF-8").
		WithAssertMaxAttempts(10).
		WithAssertInitialDelay(50 * time.Millisecond)

	if err := emu.Start(ctx); err != nil {
		t.Fatalf("failed to start: %v", err)
	}
	defer emu.Close()

	// This should retry until the output appears
	emu.AssertScreenContains(t, "delayed output")
}

func TestAssertConfiguration(t *testing.T) {
	ctx := context.Background()

	// Test custom retry configuration
	emu := vtermtest.New(5, 40).
		Command("echo", "test").
		Env("LANG=C.UTF-8").
		WithAssertMaxAttempts(3).
		WithAssertInitialDelay(10 * time.Millisecond).
		WithAssertBackoffFactor(1.5)

	if err := emu.Start(ctx); err != nil {
		t.Fatalf("failed to start: %v", err)
	}
	defer emu.Close()

	// Should work with custom config
	emu.AssertScreenContains(t, "test")
}

// TestAssertFailure tests that assertions fail when they should
func TestAssertFailure(t *testing.T) {
	ctx := context.Background()

	t.Run("LineEqual fails on mismatch", func(t *testing.T) {
		// Use a custom test type to capture failures
		mockT := &mockTest{}
		
		emu := vtermtest.New(5, 40).
			Command("echo", "actual").
			Env("LANG=C.UTF-8").
			WithAssertMaxAttempts(1) // Don't retry for this test

		if err := emu.Start(ctx); err != nil {
			t.Fatalf("failed to start: %v", err)
		}
		defer emu.Close()

		time.Sleep(100 * time.Millisecond)

		// This should fail
		emu.AssertLineEqual(mockT, 0, "expected")

		if !mockT.failed {
			t.Error("AssertLineEqual should have failed")
		}
		if !strings.Contains(mockT.message, "mismatch") {
			t.Errorf("Error message should contain 'mismatch', got: %s", mockT.message)
		}
	})
}

// mockTest implements a minimal testing.T interface for testing failures
type mockTest struct {
	failed  bool
	message string
}

func (m *mockTest) Helper() {}

func (m *mockTest) Fatalf(format string, args ...interface{}) {
	m.failed = true
	m.message = fmt.Sprintf(format, args...)
}