package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/c-bata/vtermtest"
	"github.com/c-bata/vtermtest/keys"
)

func main() {
	ctx := context.Background()

	emu := vtermtest.New(10, 80).
		Command("go", "run", "./simple_example/main.go").
		Env("LANG=C.UTF-8", "TERM=xterm")

	if err := emu.Start(ctx); err != nil {
		log.Fatalf("failed to start emulator: %v", err)
	}
	defer emu.Close()

	// Wait for prompt
	time.Sleep(500 * time.Millisecond)
	emu.WaitStable(100*time.Millisecond, 2*time.Second)

	// Type "us<Tab>" (dev branch should handle this correctly)
	if err := emu.KeyPress(keys.Text("us"), keys.Tab); err != nil {
		log.Fatalf("failed to send keys: %v", err)
	}

	// Wait for completion
	time.Sleep(300 * time.Millisecond)
	emu.WaitStable(100*time.Millisecond, 2*time.Second)

	// Get screen
	screen, err := emu.GetScreenText()
	if err != nil {
		log.Fatalf("failed to get screen: %v", err)
	}

	fmt.Println(screen)
}
