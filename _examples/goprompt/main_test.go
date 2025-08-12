package main

import (
	"context"
	"testing"

	"github.com/c-bata/vtermtest"
	"github.com/c-bata/vtermtest/keys"
)

func TestGoPromptCompletion(t *testing.T) {
	emu := vtermtest.New(10, 80).
		Command("go", "run", "./simple_example/main.go").
		Env("LANG=C.UTF-8", "TERM=xterm")

	if err := emu.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	defer emu.Close()

	// Wait for prompt to appear
	emu.AssertScreenContains(t, ">>>")

	if err := emu.KeyPress(keys.Text("u"), keys.Tab); err != nil {
		t.Fatal(err)
	}

	emu.AssertLineEqual(t, 0, ">>> users")
	emu.AssertLineEqual(t, 1, "      users    user table")
}

func TestDsl(t *testing.T) {
	emu := vtermtest.New(10, 80).
		Command("go", "run", "./simple_example/main.go").
		Env("LANG=C.UTF-8", "TERM=xterm")

	if err := emu.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	defer emu.Close()

	// Wait for prompt to appear
	emu.AssertScreenContains(t, ">>>")

	if err := emu.KeyPressString("u<Tab>"); err != nil {
		t.Fatal(err)
	}

	emu.AssertLineEqual(t, 0, ">>> users")
	emu.AssertLineEqual(t, 1, "      users    user table")
}
