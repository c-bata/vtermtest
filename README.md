# vtermtest

Snapshot testing for interactive TUIs/REPLs (e.g., go-prompt) using a real PTY + libvterm.
Feed keystrokes (printable + special keys) and capture the rendered screen as plain text at any step.

## Usage

### Installation

```
go get github.com/c-bata/vtermtest
```

> [!NOTE]
> Requires CGO and a working libvterm toolchain.


### Quickstart (Assert Inline)

```go
package myapp_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/c-bata/vtermtest"
	"github.com/c-bata/vtermtest/keys"
)

func Test_GoPrompt_InlineAssert(t *testing.T) {
	// rows, cols
	emu := vtermtest.New(4, 50).
		.Command("go", "run", "./_examples/go_prompt_example.go").
		.Env("LANG=C.UTF-8")
	t.Cleanup(func() { _ = emu.Close() })

	if err := emu.Start(t.Context()); err != nil {
		t.Fatalf("start: %v", err)
	}

	// Type query + press Tab to trigger completion
	if err := emu.KeyPress(keys.Text("SELECT * FROM "), keys.Tab); err != nil {
		t.Fatalf("send: %v", err)
	}

	// Assert lines (row index starts at 0)
	emu.AssertLineEqual(t, 0, ">>> SELECT * FROM articles")
	emu.AssertLineEqual(t, 1, "                   users     user table")
	emu.AssertLineEqual(t, 2, "                   articles  articles table")
	emu.AssertLineEqual(t, 3, "") // empty line

	// Or assert the whole screen
	emu.AssertScreenEqual(t, `
>>> SELECT * FROM articles
                   users     user table
                   articles  articles table

`)
}
```

### Golden/Snapshot Test

```go
package myapp_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/c-bata/vtermtest"
	"github.com/c-bata/vtermtest/keys"
)

func TestSnapshotGoPrompt(t *testing.T) {
	emu := vtermtest.New(4, 50).
		.Command("go", "run", "_examples/go_prompt_example.go").
		.Env("LANG=C.UTF-8")
	t.Cleanup(func() { _ = emu.Close() })

	err := emu.Start(t.Context());

	_ = emu.KeyPress(keys.Text("SELECT * FROM "), keys.Tab);
	if !emu.WaitStable(50*time.Millisecond, 2*time.Second) {
		t.Fatalf("screen did not stabilize")
	}
    screen, err := emu.GetScreenText()

	// Golden comparison
	golden := filepath.Join("testdata", "sql_example.golden.txt")
	if update := os.Getenv("VTERMTEST_GOLDEN_UPDATE") == "1" {
		if err := os.WriteFile(golden, []byte(screen), 0o644); err != nil {
			t.Fatalf("write golden: %v", err)
		}
	} else {
		expect, err := os.ReadFile(golden)
		if string(expect) != screen {
			t.Fatalf("screen mismatch.\n--- want ---\n%s\n--- got ---\n%s", expect, screen)
		}
	}
}
```

### Keys API

```
// Text & printable
K.Text("use")

// Common keys
K.Tab
K.Enter
K.Backspace
K.Up, K.Down, K.Left, K.Right
K.Home, K.End, K.PageUp, K.PageDown
K.Delete

// Fn Keys
K.F(1) ... K.F(24)

// Ctrl keys
K.CtrlA, K.CtrlB, K.CtrlC
```

## Limitations

- This library currently focuses on characters; color/attributes are not included (by design).
- Tested on Linux/macOS; Windows support depends on your PTY backend and toolchain.
- Requires CGO and a libvterm toolchain supported by your OS.

## Acknowledgements

- [libvterm](https://www.leonerd.org.uk/code/libvterm/) — An abstract library implementation of a VT220/xterm/ECMA-48 terminal emulator.
- [mattn/go-libvterm](https://github.com/mattn/go-libvterm) — Go bindings used internally.
- [creack/pty](https://github.com/creack/pty) — PTY management.
