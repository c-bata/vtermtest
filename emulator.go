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

	"github.com/creack/pty"
	libvterm "github.com/mattn/go-libvterm"
)

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
}

func New(rows, cols uint16) *Emulator {
	return &Emulator{
		rows:       rows,
		cols:       cols,
		readerDone: make(chan struct{}),
	}
}

func (e *Emulator) Command(path string, args ...string) *Emulator {
	e.commandPath = path
	e.commandArgs = args
	return e
}

func (e *Emulator) Env(env ...string) *Emulator {
	e.env = append(e.env, env...)
	return e
}

func (e *Emulator) Dir(dir string) *Emulator {
	e.dir = dir
	return e
}

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

func (e *Emulator) WaitStable(quiet, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)

	for {
		e.mu.Lock()
		lastActivity := e.lastActivity
		e.mu.Unlock()

		if time.Since(lastActivity) >= quiet {
			return true
		}

		if time.Now().After(deadline) {
			return false
		}

		time.Sleep(10 * time.Millisecond)
	}
}