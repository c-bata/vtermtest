package vtermtest_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/c-bata/vtermtest"
	"github.com/c-bata/vtermtest/keys"
)

func TestEnvAndDir(t *testing.T) {
	ctx := context.Background()

	// Test Env
	t.Run("Environment Variables", func(t *testing.T) {
		emu := vtermtest.New(10, 80).
			Command("sh", "-c", "echo $TEST_VAR").
			Env("TEST_VAR=hello_world", "LANG=C.UTF-8")

		if err := emu.Start(ctx); err != nil {
			t.Fatalf("failed to start: %v", err)
		}
		defer emu.Close()

		time.Sleep(200 * time.Millisecond)
		emu.WaitStable(100*time.Millisecond, 2*time.Second)

		screen, err := emu.GetScreenText()
		if err != nil {
			t.Fatalf("failed to get screen: %v", err)
		}

		if !strings.Contains(screen, "hello_world") {
			t.Errorf("Expected 'hello_world' in output, got: %s", screen)
		}
	})

	// Test Dir
	t.Run("Working Directory", func(t *testing.T) {
		emu := vtermtest.New(10, 80).
			Command("pwd").
			Dir("/tmp")

		if err := emu.Start(ctx); err != nil {
			t.Fatalf("failed to start: %v", err)
		}
		defer emu.Close()

		time.Sleep(200 * time.Millisecond)
		emu.WaitStable(100*time.Millisecond, 2*time.Second)

		screen, err := emu.GetScreenText()
		if err != nil {
			t.Fatalf("failed to get screen: %v", err)
		}

		if !strings.Contains(screen, "/tmp") {
			t.Errorf("Expected '/tmp' in output, got: %s", screen)
		}
	})
}

func TestCtrlKeys(t *testing.T) {
	ctx := context.Background()

	emu := vtermtest.New(10, 80).
		Command("cat").
		Env("LANG=C.UTF-8")

	if err := emu.Start(ctx); err != nil {
		t.Fatalf("failed to start: %v", err)
	}
	defer emu.Close()

	// Send text followed by Ctrl+D (EOF)
	if err := emu.KeyPress(keys.Text("test"), keys.Enter, keys.CtrlD); err != nil {
		t.Fatalf("failed to send keys: %v", err)
	}

	time.Sleep(200 * time.Millisecond)
	emu.WaitStable(100*time.Millisecond, 2*time.Second)

	screen, err := emu.GetScreenText()
	if err != nil {
		t.Fatalf("failed to get screen: %v", err)
	}

	if !strings.Contains(screen, "test") {
		t.Errorf("Expected 'test' in output, got: %s", screen)
	}
}