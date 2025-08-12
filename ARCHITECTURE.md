# ARCHITECTURE

This document explains how **vtermtest** works internally and why it’s designed this way.
Scope: **deterministic snapshot testing** for interactive TUIs/REPLs (e.g., go-prompt) using a **real PTY** + **terminal emulation** to turn bytes into a **rendered screen** you can assert against.

## Goals & Non-Goals

**Goals**

* Launch a target program under a real PTY with a fixed window size.
* Send keystrokes (printable + special keys) via a clear, typed keys API.
* Use libvterm to emulate the terminal and obtain the rendered screen.
* Provide stable, deterministic assertions suitable for golden/snapshot tests.
* Keep the public surface small and obvious (no DSL ambiguity).
* Be tool-agnostic: works with go-prompt and other TUIs.

**Non-Goals**

* Video/animation rendering (GIF/MP4).
* Full fidelity of colors/SGR attributes in the base snapshot (characters first; attributes optional later).
* Replaying mouse or complex device control (may be added separately).

## Package layout

```
vtermtest/
  README.md
  ARCHITECTURE.md
  emulator.go          // core
  screen.go            // snapshot/stringification
  assert.go            // adaptive asserts
  keys/
    keys.go            // Event, helpers (Text, Ctrl, Alt, Tab, Enter, ...)
```

## Modules & responsibilities

### 1) Core `Emulator`

* **Process/PTY lifecycle**
  * `Command(path, args...)`, `Env(...)`, `Dir(...)`, `Start(ctx)`, `Close()`.
  * Uses `creack/pty` to start the child process attached to a PTY.
  * Sets window size on PTY creation; mirrors size in libvterm.

* **Terminal emulation**
  * Creates a `libvterm` instance of the same size.
  * A **reader goroutine** consumes PTY bytes, writes them to libvterm, and calls `Flush()`.

* **Screen readback**
  * `SnapshotString()` walks rows×cols, extracts runes, and trims **trailing spaces** per line (reduces noise).

* **Timing**
  * `WaitStable(quiet, timeout)` checks inactivity via a `lastActivity` timestamp updated by the reader.
  * Assertions also implement their **own adaptive waits** (details below).

* **Sync model**
  * A `sync.Mutex` protects libvterm state and `lastActivity`.


### 2) `keys` package

* A tiny, dependency-free layer returning **final byte sequences**:
  * `K.Text("...")`
  * `K.Tab`, `K.Enter`, `K.Backspace`, navigation keys, `K.F(n)`.
  * `K.CtrlC`, `K.AltX`.
* Public senders:
  * `Emulator.KeyPress(ev ...keys.Key) error`
* **No DSL**: everything is explicit and IDE-discoverable.

## Lifecycle

1. **Construct**: `emu := vtermtest.New(rows, cols)`.
2. **Configure**:
   * `Command("go", "run", "./_examples/app.go")`
   * `Env("TERM=xterm-256color", "LANG=C.UTF-8")`
3. **Start**:
   * `pty.Start(cmd)`, `pty.Setsize(...)`
   * init `libvterm` with matching size
   * spawn reader goroutine:

     ```go
     for {
       n, err := ptmx.Read(buf)
       if n > 0 {
         lock()
         vtermState.Write(buf[:n])
         screen.Flush()
         lastActivity = time.Now()
         unlock()
       }
       if err != nil { break }
     }
     ```
4. **Interact**: `Send/SendAll` with `keys.Event`.
5. **Assert**: take snapshots or call built-in assertions.
6. **Close**: close PTY, kill child (best-effort), wait for reader to finish.

## Screen model & snapshot

* `GetScreenText()`:
  * Iterates rows and cols, converts each cell’s rune (fallback to space for 0).
  * Trims **right-side spaces** per line.
  * Joins with `\n`.
* Rationale for trimming:
  * Most TUIs clear the right edge; trimming prevents noisy diffs without losing meaningful content.
* (Future) optional “no-trim” mode and attribute capture (SGR) behind flags.

## Stability & assertions

### Explicit waiting (API)

* `WaitStable(quiet, timeout)` is exposed for manual orchestration when needed.

### **Adaptive assertions (recommended)**

All `Assert*` methods **retry automatically** so you don’t have to call `WaitStable` first:

* `AssertLineEqual(t, row, want string)`
* `AssertScreenEqual(t, want string)`
* `AssertScreenContains(t, substr string)`

**Strategy**

* Try `GetScreenText()` and check the predicate.
* If it fails, sleep and retry **up to N attempts**, using exponential backoff.
* If any attempt passes, return success; else print a helpful diff.

**Defaults**

* `maxAttempts = 6`
* `initialDelay = 20ms`
* backoff factor `×2` (20, 40, 80, 160, 320, 640ms; total \~1.26s)
* configurable via options on `Emulator`:
  * `WithAssertMaxAttempts(n int)`
  * `WithAssertInitialDelay(d time.Duration)`
  * `WithAssertBackoffFactor(f float64)`

**Pseudo-code**

```go
func (e *Emulator) assertEventually(t *testing.T, check func(screen string) error) {
    delay := e.initialDelay
    for i := 0; i < e.maxAttempts; i++ {
        screen := e.SnapshotString()
        if err := check(screen); err == nil {
            return
        }
        time.Sleep(delay)
        delay = time.Duration(float64(delay) * e.backoffFactor)
    }
    // last attempt
    screen := e.SnapshotString()
    if err := check(screen); err != nil {
        t.Fatalf("assertion failed after %d attempts\n%s", e.maxAttempts, prettyDiff(err, screen))
    }
}
```

## Determinism & environment

* Fix the size: `New(24, 80)` and keep it constant.
* Set `TERM=xterm-256color` (unless your app needs another).
* Use a stable UTF-8 locale: `LANG=C.UTF-8`.
* Avoid time-dependent output, or inject a fake clock in your app.
* Be aware of IME/composition if your app processes raw input; tests should feed deterministic bytes via `keys`.

## Error handling & cleanup

* `Start(ctx)` returns errors early (PTY, exec, sizing, libvterm init).
* `Close()` closes PTY, kills process (if still running), and waits for the reader to exit on `EOF`.
* Reader loop treats `io.EOF` as normal termination; other read errors end the loop.

## Public API surface (stable)

```go
// Construction & config
func New(rows, cols uint16) *Emulator
func (e *Emulator) Command(path string, args ...string) *Emulator
func (e *Emulator) Env(env ...string) *Emulator
func (e *Emulator) Dir(dir string) *Emulator

// Lifecycle
func (e *Emulator) Start(ctx context.Context) error
func (e *Emulator) Close() error

// Sending keys (typed API only)
func (e *Emulator) Send(ev keys.Event) error
func (e *Emulator) SendAll(evs ...keys.Event) error

// Timing
func (e *Emulator) WaitStable(quiet, timeout time.Duration) bool
func (e *Emulator) Resize(rows, cols uint16) error

// Screen readback
func (e *Emulator) GetScreenText() string

// Assertions (with adaptive waits)
func (e *Emulator) AssertLineEqual(t *testing.T, row int, want string)
func (e *Emulator) AssertScreenEqual(t *testing.T, want string)
func (e *Emulator) AssertScreenContains(t *testing.T, substr string)

// Options (suggested)
func (e *Emulator) WithAssertMaxAttempts(n int) *Emulator
func (e *Emulator) WithAssertInitialDelay(d time.Duration) *Emulator
func (e *Emulator) WithAssertBackoffFactor(f float64) *Emulator
func (e *Emulator) WithTrimTrailingSpaces(enabled bool) *Emulator
func (e *Emulator) WithEnterNewline(useLF bool) *Emulator      // default CR
func (e *Emulator) WithBackspaceBS(useBS bool) *Emulator       // default DEL
```

## Unicode, width & wrapping

* libvterm handles most ECMA-48/ANSI cursor movement, erasing, and wrapping.
* For visual width–sensitive tests (CJK/fullwidth/combining), we use `github.com/mattn/go-runewidth` to calculate proper display widths:
  * Full-width characters (CJK, emoji) are correctly measured as 2 columns
  * Zero-width characters (combining marks) are handled appropriately
  * The `GetScreenText()` method applies runewidth calculations when extracting text to ensure column positions align with visual expectations
* Consistent UTF-8 locale (`LANG=C.UTF-8`) and fixed terminal size are still recommended for deterministic tests.

## Portability

* **Linux/macOS**: primary targets (CGO + libvterm available).
* **Windows**: possible with CGO + suitable PTY backend; considered best-effort in v0 (contributions welcome).

## Testing vtermtest itself

* Self-tests run a tiny demo binary (echo, wraps, cursor ops).
* Golden files verify:
  * navigation, erase sequences, wraps
  * Enter/Backspace variants
  * assertion retries (simulate delayed output)
* CI sets `TERM` and `LANG` and uses a fixed size.

## Future work

* Optional **attributes** (SGR/bold/color) in snapshots.
* Built-in **pretty diff** for large screens (side-by-side).
* `.cast` export for replay/debugging.
* Windows setup docs & helpers.
* Terminal “profiles” (xterm-256color, linux, screen-256color) with per-profile key maps.
