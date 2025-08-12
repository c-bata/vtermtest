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

### Command-line Interface

```
$ vtermtest-cli --command kube-prompt --keys "get p<Tab>" --rows 12
kube-prompt v1.0.11 (rev-ac5964a)
Please use `exit` or `Ctrl-D` to exit this program.
>>> get p
          persistentvolumeclaims
          persistentvolumes
          pod
          podsecuritypolicies
          podtemplates
          pvc


$
```

```
$ vtermtest-cli --help
vtermtest-cli - Terminal emulator testing tool

USAGE:
    vtermtest-cli --command "COMMAND" [OPTIONS]

OPTIONS:
    --command STRING    Command to execute (required)
    --keys STRING       Key sequence in DSL format
    --rows INT          Terminal rows (default: 24)
    --cols INT          Terminal columns (default: 80)
    --output FILE       Output file (default: stdout)
    --timeout DURATION  Total timeout for command execution (default: 30s)
    --stable-duration DURATION  Duration screen must remain unchanged (default: 200ms)
    --stable-timeout DURATION   Timeout for screen stabilization (default: 10s)
    --env STRING        Environment variables (KEY=VALUE,...)
    --dir STRING        Working directory
    --delimiter STRING  DSL tag delimiters (default: "<>")

KEY DSL:
    Text: hello world
    Keys: <Tab> <Enter> <BS> <Del> <Esc> <Space> <Up> <Down> <Left> <Right>
    Ctrl: <C-a> ... <C-z>  Alt: <A-a> ... <A-z>  Fn: <F1> ... <F24>
    Nav: <Home> <End> <PageUp> <PageDown>
    Wait: <WaitStable> <WaitFor text>
    Escape: << (literal <)
```

### Quickstart (Go API)

```go
package myapp_test

import (
	"context"
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
		Command("go", "run", "./_examples/go_prompt_example.go").
		Env("LANG=C.UTF-8")
	t.Cleanup(func() { _ = emu.Close() })

	if err := emu.Start(context.Background()); err != nil {
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
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/c-bata/vtermtest"
	"github.com/c-bata/vtermtest/keys"
)

func TestSnapshotGoPrompt(t *testing.T) {
	emu := vtermtest.New(4, 50).
		Command("go", "run", "_examples/go_prompt_example.go").
		Env("LANG=C.UTF-8")
	t.Cleanup(func() { _ = emu.Close() })

	if err := emu.Start(context.Background()); err != nil {
		t.Fatalf("start: %v", err)
	}

	_ = emu.KeyPress(keys.Text("SELECT * FROM "), keys.Tab);
	if !emu.WaitStable(50*time.Millisecond, 2*time.Second) {
		t.Fatalf("screen did not stabilize")
	}
	screen, err := emu.GetScreenText()
	if err != nil {
		t.Fatalf("get screen: %v", err)
	}

	// Golden comparison
	golden := filepath.Join("testdata", "sql_example.golden.txt")
	if os.Getenv("VTERMTEST_GOLDEN_UPDATE") == "1" {
		if err := os.WriteFile(golden, []byte(screen), 0o644); err != nil {
			t.Fatalf("write golden: %v", err)
		}
	} else {
		expect, err := os.ReadFile(golden)
		if err != nil {
			t.Fatalf("read golden: %v", err)
		}
		if string(expect) != screen {
			t.Fatalf("screen mismatch.\n--- want ---\n%s\n--- got ---\n%s", expect, screen)
		}
	}
}
```

### Keys API

#### Keys

```go
// Text & printable
keys.Text("use")

// Common keys
keys.Tab
keys.Enter
keys.Backspace
keys.Up, keys.Down, keys.Left, keys.Right
keys.Home, keys.End, keys.PageUp, keys.PageDown
keys.Delete

// Fn Keys
keys.F(1) ... keys.F(24)

// Ctrl keys
keys.CtrlA, keys.CtrlB, keys.CtrlC
```

#### DSL

```go
// Simple text with special keys
emu.KeyPressString("SELECT * FROM us<Tab>")

// Vim-like operations
emu.KeyPressString("ihello world<Esc>:wq<Enter>")

// Complex editing
emu.KeyPressString("<C-a>deleted<C-k>new text<Enter>")

// Escaped angle brackets
emu.KeyPressString("echo <<literal angle bracket>>")
```

**DSL Notation:**
- Regular text: typed as-is
- Special keys: `<Tab>` `<Enter>` `<BS>` `<Del>` `<Esc>` `<Space>`
- Arrow keys: `<Up>` `<Down>` `<Left>` `<Right>`
- Ctrl keys: `<C-a>` ... `<C-z>`
- Alt keys: `<A-a>` ... `<A-z>`
- Function keys: `<F1>` ... `<F24>`
- Navigation: `<Home>` `<End>` `<PageUp>` `<PageDown>`
- Escape: `<<` for literal `<`

## Limitations

- This library currently focuses on characters; color/attributes are not included (by design).
- Tested on Linux/macOS; Windows support depends on your PTY backend and toolchain.
- Requires CGO and a libvterm toolchain supported by your OS.

## Acknowledgements

- [libvterm](https://www.leonerd.org.uk/code/libvterm/) — An abstract library implementation of a VT220/xterm/ECMA-48 terminal emulator.
- [mattn/go-libvterm](https://github.com/mattn/go-libvterm) — Go bindings used internally.
- [creack/pty](https://github.com/creack/pty) — PTY management.
