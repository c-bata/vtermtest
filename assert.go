package vtermtest

import (
	"fmt"
	"strings"
	"time"
)

// Default retry configuration
const (
	defaultMaxAttempts    = 6
	defaultInitialDelay   = 20 * time.Millisecond
	defaultBackoffFactor  = 2.0
)

// TestingT is the subset of testing.T used by assertions
type TestingT interface {
	Helper()
	Fatalf(format string, args ...interface{})
}

// AssertLineEqual asserts that a specific line equals the expected string.
// It retries with exponential backoff until the assertion passes or max attempts is reached.
func (e *Emulator) AssertLineEqual(t TestingT, row int, want string) {
	t.Helper()
	
	e.assertWithRetry(t, func() error {
		got, err := e.GetLine(row)
		if err != nil {
			return fmt.Errorf("failed to get line %d: %v", row, err)
		}
		
		if got != want {
			return fmt.Errorf("line %d mismatch:\nwant: %q\ngot:  %q", row, want, got)
		}
		return nil
	})
}

// AssertScreenEqual asserts that the entire screen matches the expected string.
// Leading/trailing whitespace in want is trimmed, and empty lines at the start are ignored.
func (e *Emulator) AssertScreenEqual(t TestingT, want string) {
	t.Helper()
	
	// Normalize expected output
	want = strings.TrimSpace(want)
	
	e.assertWithRetry(t, func() error {
		got, err := e.GetScreenText()
		if err != nil {
			return fmt.Errorf("failed to get screen: %v", err)
		}
		
		// Normalize actual output
		got = strings.TrimSpace(got)
		
		if got != want {
			return fmt.Errorf("screen mismatch:\n--- want ---\n%s\n--- got ---\n%s", want, got)
		}
		return nil
	})
}

// AssertScreenContains asserts that the screen contains the given substring.
func (e *Emulator) AssertScreenContains(t TestingT, substr string) {
	t.Helper()
	
	e.assertWithRetry(t, func() error {
		got, err := e.GetScreenText()
		if err != nil {
			return fmt.Errorf("failed to get screen: %v", err)
		}
		
		if !strings.Contains(got, substr) {
			return fmt.Errorf("screen does not contain %q:\n%s", substr, got)
		}
		return nil
	})
}

// assertWithRetry implements the retry logic with exponential backoff
func (e *Emulator) assertWithRetry(t TestingT, check func() error) {
	t.Helper()
	
	maxAttempts := e.getMaxAttempts()
	delay := e.getInitialDelay()
	backoffFactor := e.getBackoffFactor()
	
	var lastErr error
	
	for attempt := 0; attempt < maxAttempts; attempt++ {
		if err := check(); err == nil {
			return // Success
		} else {
			lastErr = err
		}
		
		// Don't sleep after the last attempt
		if attempt < maxAttempts-1 {
			time.Sleep(delay)
			delay = time.Duration(float64(delay) * backoffFactor)
		}
	}
	
	// All attempts failed
	if lastErr != nil {
		t.Fatalf("assertion failed after %d attempts: %v", maxAttempts, lastErr)
	}
}

// Configuration methods for retry behavior
type assertConfig struct {
	maxAttempts    int
	initialDelay   time.Duration
	backoffFactor  float64
}

// Add to Emulator struct (in emulator.go):
// assertCfg assertConfig

func (e *Emulator) getMaxAttempts() int {
	if e.assertCfg.maxAttempts > 0 {
		return e.assertCfg.maxAttempts
	}
	return defaultMaxAttempts
}

func (e *Emulator) getInitialDelay() time.Duration {
	if e.assertCfg.initialDelay > 0 {
		return e.assertCfg.initialDelay
	}
	return defaultInitialDelay
}

func (e *Emulator) getBackoffFactor() float64 {
	if e.assertCfg.backoffFactor > 0 {
		return e.assertCfg.backoffFactor
	}
	return defaultBackoffFactor
}

// WithAssertMaxAttempts sets the maximum number of retry attempts for assertions
func (e *Emulator) WithAssertMaxAttempts(n int) *Emulator {
	e.assertCfg.maxAttempts = n
	return e
}

// WithAssertInitialDelay sets the initial delay between retry attempts
func (e *Emulator) WithAssertInitialDelay(d time.Duration) *Emulator {
	e.assertCfg.initialDelay = d
	return e
}

// WithAssertBackoffFactor sets the backoff multiplier for retry delays
func (e *Emulator) WithAssertBackoffFactor(f float64) *Emulator {
	e.assertCfg.backoffFactor = f
	return e
}