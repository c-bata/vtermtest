// Package vtermtest provides snapshot testing for interactive TUIs/REPLs using a real PTY and libvterm.
package vtermtest

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/c-bata/vtermtest/keys"
	"github.com/creack/pty"
	libvterm "github.com/mattn/go-libvterm"
)

// Emulator represents a terminal emulator for testing interactive programs.
// It creates a PTY, launches a process, and uses libvterm to emulate terminal behavior.
type Emulator struct {
	rows uint16
	cols uint16

	cmd  *exec.Cmd
	ptmx *os.File

	vt     *libvterm.VTerm
	screen *libvterm.Screen

	mu           sync.Mutex
	lastActivity time.Time
	readerDone   chan struct{}

	commandPath string
	commandArgs []string
	env         []string
	dir         string

	assertCfg assertConfig

	// Raw bytes collection
	collectRawBytes bool
	rawBytes        []byte
}

// New creates a new Emulator with the specified terminal dimensions.
// rows and cols specify the terminal size in characters.
func New(rows, cols uint16) *Emulator {
	return &Emulator{
		rows:       rows,
		cols:       cols,
		readerDone: make(chan struct{}),
	}
}

// EnableRawBytesCollection enables collection of raw bytes from PTY.
// When enabled, all bytes read from PTY are stored and can be retrieved with GetRawBytes().
func (e *Emulator) EnableRawBytesCollection() *Emulator {
	e.collectRawBytes = true
	return e
}

// Command sets the command to execute. Returns self for method chaining.
func (e *Emulator) Command(path string, args ...string) *Emulator {
	e.commandPath = path
	e.commandArgs = args
	return e
}

// Env adds environment variables. Multiple calls append variables.
// Format: "KEY=value". Returns self for method chaining.
func (e *Emulator) Env(env ...string) *Emulator {
	e.env = append(e.env, env...)
	return e
}

// Dir sets the working directory for the command. Returns self for method chaining.
func (e *Emulator) Dir(dir string) *Emulator {
	e.dir = dir
	return e
}

// Start launches the command in a PTY and begins terminal emulation.
// The context can be used to control the lifetime of the process.
func (e *Emulator) Start(ctx context.Context) error {
	if e.commandPath == "" {
		return errors.New("no command specified")
	}

	e.cmd = exec.CommandContext(ctx, e.commandPath, e.commandArgs...)
	if len(e.env) > 0 {
		e.cmd.Env = append(os.Environ(), e.env...)
	}
	if e.dir != "" {
		e.cmd.Dir = e.dir
	}

	ptmx, err := pty.StartWithSize(e.cmd, &pty.Winsize{
		Rows: e.rows,
		Cols: e.cols,
	})
	if err != nil {
		return err
	}
	e.ptmx = ptmx

	e.vt = libvterm.New(int(e.rows), int(e.cols))
	e.screen = e.vt.ObtainScreen()
	e.screen.Reset(true)

	go e.readLoop()

	return nil
}

func (e *Emulator) readLoop() {
	defer close(e.readerDone)
	buf := make([]byte, 4096)

	for {
		n, err := e.ptmx.Read(buf)
		if n > 0 {
			e.mu.Lock()
			// Collect raw bytes if enabled
			if e.collectRawBytes {
				e.rawBytes = append(e.rawBytes, buf[:n]...)
			}
			_, writeErr := e.vt.Write(buf[:n])
			if writeErr == nil {
				e.screen.Flush()
			}
			e.lastActivity = time.Now()
			e.mu.Unlock()
		}
		if err != nil {
			if err != io.EOF {
				// Log error if needed
			}
			break
		}
	}
}

// Close terminates the process and cleans up resources.
// It closes the PTY, kills the process if still running, and waits for cleanup.
func (e *Emulator) Close() error {
	var errs []error

	// Close PTY
	if e.ptmx != nil {
		if err := e.ptmx.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	// Kill process if still running
	if e.cmd != nil && e.cmd.Process != nil {
		if err := e.cmd.Process.Kill(); err != nil {
			// Process might already be dead, which is OK
			if !strings.Contains(err.Error(), "process already finished") {
				errs = append(errs, err)
			}
		}
		// Wait for process to exit
		if err := e.cmd.Wait(); err != nil {
			// Ignore "signal: killed" errors
			if !strings.Contains(err.Error(), "signal: killed") {
				errs = append(errs, err)
			}
		}
	}

	// Wait for reader goroutine to finish
	select {
	case <-e.readerDone:
	case <-time.After(2 * time.Second):
		errs = append(errs, errors.New("timeout waiting for reader to finish"))
	}

	// Close libvterm
	if e.vt != nil {
		if err := e.vt.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return errors.New(fmt.Sprintf("close errors: %v", errs))
	}
	return nil
}

// KeyPress sends keystrokes to the terminal.
// Use the keys package for special keys (e.g., keys.Tab, keys.Enter).
func (e *Emulator) KeyPress(keys ...[]byte) error {
	if e.ptmx == nil {
		return errors.New("emulator not started")
	}

	for _, key := range keys {
		if _, err := e.ptmx.Write(key); err != nil {
			return err
		}
	}
	return nil
}

// KeyPressString sends keystrokes using DSL notation.
// Example: "hello<Tab>world<C-c>" sends "hello", Tab key, "world", then Ctrl-C.
// Special DSL: <WaitStable> waits for screen to stabilize.
// See keys.Parse for supported notation.
func (e *Emulator) KeyPressString(dsl string) error {
	return e.KeyPressStringWithOptions(dsl, keys.DefaultParseOptions())
}

// KeyPressStringWithOptions sends keystrokes using DSL notation with custom tag delimiters.
// Example with options {TagStart: '[', TagEnd: ']'}: "hello[Tab]world[C-c]"
func (e *Emulator) KeyPressStringWithOptions(dsl string, opts keys.ParseOptions) error {
	parsedKeys, err := keys.ParseWithOptions(dsl, opts)
	if err != nil {
		return fmt.Errorf("parse DSL: %w", err)
	}

	for _, key := range parsedKeys {
		keyStr := string(key)
		if keyStr == "__WAITSTABLE__" {
			if !e.WaitStable(100*time.Millisecond, 5*time.Second) {
				return fmt.Errorf("screen did not stabilize")
			}
		} else if strings.HasPrefix(keyStr, "__WAITFOR__") {
			text := keyStr[11:] // Remove "__WAITFOR__" prefix
			if err := e.WaitFor(text, 5*time.Second); err != nil {
				return err
			}
		} else {
			if err := e.KeyPress(key); err != nil {
				return err
			}
		}
	}
	return nil
}

// WaitStable waits until the screen output is stable (no changes for 'quiet' duration).
// Returns true if stable within timeout, false if timeout exceeded.
// quiet: duration of inactivity to consider stable
// timeout: maximum time to wait
func (e *Emulator) WaitStable(quiet, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	var lastScreen string
	var stableStart time.Time

	// Get initial screen content
	screen, err := e.GetScreenText()
	if err != nil {
		return false
	}
	lastScreen = screen
	stableStart = time.Now()

	for {
		if time.Now().After(deadline) {
			return false
		}

		time.Sleep(10 * time.Millisecond)

		// Get current screen content
		currentScreen, err := e.GetScreenText()
		if err != nil {
			return false
		}

		if currentScreen == lastScreen {
			// Screen content hasn't changed
			if time.Since(stableStart) >= quiet {
				return true
			}
		} else {
			// Screen content changed, reset stable timer
			lastScreen = currentScreen
			stableStart = time.Now()
		}
	}
}

// WaitFor waits until the specified text appears on the screen.
// Returns error if text doesn't appear within timeout.
// timeout: maximum time to wait for the text to appear
func (e *Emulator) WaitFor(text string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	var lastScreen string

	for {
		screen, err := e.GetScreenText()
		if err != nil {
			return fmt.Errorf("failed to get screen text: %w", err)
		}

		lastScreen = screen
		if strings.Contains(screen, text) {
			return nil
		}

		if time.Now().After(deadline) {
			return fmt.Errorf("text %q not found within timeout\nCurrent screen content:\n%s", text, lastScreen)
		}

		time.Sleep(50 * time.Millisecond)
	}
}

// Resize changes the terminal size dynamically.
// Both PTY and libvterm are resized to match the new dimensions.
func (e *Emulator) Resize(rows, cols uint16) error {
	if e.ptmx == nil {
		return errors.New("emulator not started")
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	// Update internal dimensions
	e.rows = rows
	e.cols = cols

	// Resize PTY
	if err := pty.Setsize(e.ptmx, &pty.Winsize{
		Rows: rows,
		Cols: cols,
	}); err != nil {
		return fmt.Errorf("failed to resize PTY: %w", err)
	}

	// Resize libvterm
	if e.vt != nil {
		e.vt.SetSize(int(rows), int(cols))
	}

	// Mark as activity to trigger any waiting operations
	e.lastActivity = time.Now()

	return nil
}

// GetRawBytes returns the raw bytes collected from PTY.
// Raw bytes collection must be enabled with EnableRawBytesCollection().
// Returns a copy of the collected bytes.
func (e *Emulator) GetRawBytes() []byte {
	e.mu.Lock()
	defer e.mu.Unlock()

	if !e.collectRawBytes {
		return nil
	}

	// Return a copy to prevent external modification
	result := make([]byte, len(e.rawBytes))
	copy(result, e.rawBytes)
	return result
}
