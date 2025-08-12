# vtermtest Tasks

## Phase 1: MVP (Minimum Viable Product)

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

## Phase 2: Basic Functionality

### Enhanced Keys Support
- [x] Expand keys package
    - Add arrow keys (Up, Down, Left, Right)
    - Add common control keys (Ctrl+C, Ctrl+D)
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
    - Implement basic cleanup on Close()
    - Prevent zombie processes

## Phase 3: Usability Improvements

### Assertions (Testing Convenience)
- [ ] Implement assertion module
    - Create `assert.go`
    - Implement `AssertScreenEqual()` with basic retry
    - Implement `AssertLineEqual()`

### Unicode Support
- [ ] Unicode improvements
    - Integrate go-runewidth for CJK support
    - Test with emoji and wide characters

### Documentation
- [ ] Create basic documentation
    - Write simple README.md with quickstart
    - Add minimal inline comments
    - Create one complete example

## Phase 4: Polish & Stability

### Advanced Assertions
- [ ] Enhanced assertions
    - Implement `AssertScreenContains()`
    - Add retry logic with exponential backoff
    - Add configurable retry options

### Advanced Keys
- [ ] Complete keys support
    - Add all function keys (F1-F24)
    - Add Home, End, PageUp, PageDown
    - Add Alt key combinations
    - Add Delete key

### Configuration Options
- [ ] Add optional configurations
    - Add `WithAssertMaxAttempts()`
    - Add `WithAssertInitialDelay()`
    - Add `WithAssertBackoffFactor()`
    - Add `WithTrimTrailingSpaces()`
    - Add `WithEnterNewline()` and `WithBackspaceBS()`

### Platform Testing
- [ ] Platform verification
    - Test thoroughly on Linux
    - Test thoroughly on macOS
    - Document Windows limitations

## Phase 5: Advanced Features (Future)

### Testing Infrastructure
- [ ] Enhanced testing
    - Add golden/snapshot test support
    - Create comprehensive test suite
    - Add self-tests for all components

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
    - Finalize ARCHITECTURE.md
    - Fix all documentation inconsistencies
    - Add contribution guidelines

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

## Known Issues to Fix

- [ ] Fix documentation issues
    - Fix API inconsistency (Send/SendAll vs KeyPress)
    - Fix syntax error in README.md line 36
    - Fix error handling examples consistency
    - Correct variable assignment in README.md line 96

- [ ] Fix design issues
    - Add context cancellation support in reader goroutine
    - Improve error handling consistency
    - Address potential deadlock in mutex usage