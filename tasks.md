# vtermtest Tasks

## Phase 1: MVP (Minimum Viable Product) ✅

### Core Dependencies Setup
- [x] Manage dependencies
    - Add `github.com/creack/pty`
    - Add `github.com/mattn/go-libvterm`
    - Add `github.com/mattn/go-runewidth`

### Core Implementation (Essential)
- [x] Implement core emulator module
    - Create `emulator.go` with `Emulator` struct
    - Implement PTY lifecycle management (Start, Close)
    - Add libvterm integration
    - Implement reader goroutine for PTY output
    - Add basic synchronization with mutex

- [x] Implement minimal screen capture
    - Create `screen.go` with basic screen reading
    - Implement simple `GetScreenText()` method
    - Add basic trailing space trimming

- [x] Implement minimal keys package
    - Create `keys/keys.go` with basic key types
    - Add printable text support (`Text()`)
    - Implement essential keys only (Tab, Enter, Backspace)

### Basic Configuration
- [x] Implement essential builder methods
    - Add `Command()` for process execution
    - Add `New()` constructor with rows/cols

### Minimal Testing
- [x] Create basic example
    - Add simple `_examples/simple_example.go` demo
    - Create one working integration test
    - Verify PTY creation and basic I/O works

## Phase 2: Basic Functionality ✅

### Enhanced Keys Support
- [x] Expand keys package
    - Add arrow keys (Up, Down, Left, Right)
    - Add all control keys (Ctrl+A to Ctrl+Z)
    - Add function keys (F1-F12)

### Basic Stability
- [x] Add timing mechanisms
    - Implement simple `WaitStable()` method
    - Add basic timeout handling

### Basic Configuration Options
- [x] Add environment configuration
    - Implement `Env()` method
    - Implement `Dir()` method

### Error Handling
- [x] Basic error handling
    - Add error returns for all public methods
    - Implement improved cleanup on Close()
    - Prevent zombie processes

## Phase 3: Usability Improvements ✅

### Assertions (Testing Convenience)
- [x] Implement assertion module
    - Create `assert.go`
    - Implement `AssertScreenEqual()` with retry
    - Implement `AssertLineEqual()`
    - Implement `AssertScreenContains()`
    - Add exponential backoff retry logic
    - Add configurable retry options (`WithAssertMaxAttempts()`, `WithAssertInitialDelay()`, `WithAssertBackoffFactor()`)

### Unicode Support
- [x] Unicode improvements
    - Integrate go-runewidth for CJK support
    - Proper Unicode width handling in screen capture

### Documentation & Examples
- [x] Create examples and documentation
    - Update README.md with usage examples
    - Update ARCHITECTURE.md with Unicode handling
    - Create go-prompt example (`_examples/goprompt/`)
    - Create test examples with assertions

## Phase 4: Polish & Stability

### Advanced Keys
- [ ] Complete keys support
    - Add all function keys (F13-F24)
    - Add Home, End, PageUp, PageDown
    - Add Alt key combinations
    - Add Delete key

### Resize Support
- [ ] Add resize capability
    - Implement `Resize()` method
    - Update both PTY and libvterm dimensions

### CI/CD
- [ ] Setup continuous integration
    - Add GitHub Actions workflow
    - Configure test matrix
    - Add code coverage

### Documentation
- [ ] Complete documentation
    - Add contribution guidelines
    - Create more complex examples

### Export Capabilities
- [ ] Export features
    - Add `.cast` format export
    - Implement session recording
    - Add replay functionality

### Color Support
- [ ] Color and attributes
    - Design SGR attribute capture API
    - Add optional color information
    - Implement attribute comparison

### Performance
- [ ] Optimization
    - Profile and optimize screen capture
    - Reduce mutex contention
    - Add benchmarks

### Security
- [ ] Security hardening
    - Review command injection prevention
    - Sanitize environment variables
    - Add security documentation

## Completed Issues

### Fixed Documentation Issues
- [x] Updated ARCHITECTURE.md with go-runewidth integration
- [x] API is now consistent (KeyPress throughout)
- [x] Created working examples with proper error handling

### Fixed Design Issues
- [x] Improved error handling with error collection in Close()
- [x] Added proper mutex synchronization
- [x] Implemented TestingT interface for better testability