package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/c-bata/vtermtest"
	"github.com/c-bata/vtermtest/keys"
)

func main() {
	var (
		rows      = flag.Int("rows", 24, "Terminal rows (height)")
		cols      = flag.Int("cols", 80, "Terminal columns (width)")
		command   = flag.String("command", "", "Command to execute (required)")
		keySeq    = flag.String("keys", "", "Key sequence in DSL format (e.g., 'hello<Tab>world<Enter>')")
		output    = flag.String("output", "", "Output file (default: stdout)")
		timeout   = flag.Duration("timeout", 5*time.Second, "Timeout for screen stabilization")
		quiet     = flag.Duration("quiet", 100*time.Millisecond, "Quiet period to consider screen stable")
		env       = flag.String("env", "", "Environment variables (comma-separated KEY=VALUE pairs)")
		dir       = flag.String("dir", "", "Working directory")
		delimiter = flag.String("delimiter", "<>", "DSL tag delimiters (2 characters, e.g., '<>', '[]', '{}')")
		help      = flag.Bool("help", false, "Show help message")
	)

	flag.Parse()

	if *help {
		showHelp()
		return
	}

	if *command == "" {
		fmt.Fprintf(os.Stderr, "Error: --command is required\n\n")
		showHelp()
		os.Exit(1)
	}

	if *rows <= 0 || *cols <= 0 {
		fmt.Fprintf(os.Stderr, "Error: rows and cols must be positive integers\n")
		os.Exit(1)
	}

	// Parse command
	cmdParts := parseCommand(*command)
	if len(cmdParts) == 0 {
		fmt.Fprintf(os.Stderr, "Error: invalid command format\n")
		os.Exit(1)
	}

	// Create emulator
	emu := vtermtest.New(uint16(*rows), uint16(*cols))
	emu.Command(cmdParts[0], cmdParts[1:]...)

	// Set environment variables
	if *env != "" {
		envVars := parseEnvVars(*env)
		emu.Env(envVars...)
	}

	// Set working directory
	if *dir != "" {
		emu.Dir(*dir)
	}

	// Start emulator
	ctx := context.Background()
	if err := emu.Start(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Error starting emulator: %v\n", err)
		os.Exit(1)
	}
	defer emu.Close()

	// Wait for initial screen to stabilize
	if !emu.WaitStable(*quiet, *timeout) {
		fmt.Fprintf(os.Stderr, "Warning: initial screen did not stabilize within timeout\n")
	}

	// Send key sequences if provided
	if *keySeq != "" {
		// Parse delimiter
		tagStart, tagEnd, err := parseDelimiter(*delimiter)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing delimiter: %v\n", err)
			os.Exit(1)
		}

		// Create parse options
		opts := keys.ParseOptions{
			TagStart: tagStart,
			TagEnd:   tagEnd,
		}

		if err := emu.KeyPressStringWithOptions(*keySeq, opts); err != nil {
			fmt.Fprintf(os.Stderr, "Error sending keys: %v\n", err)
			os.Exit(1)
		}
	}

	// Wait for final screen to stabilize
	if !emu.WaitStable(*quiet, *timeout) {
		fmt.Fprintf(os.Stderr, "Warning: final screen did not stabilize within timeout\n")
	}

	// Get screen content
	screen, err := emu.GetScreenText()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting screen content: %v\n", err)
		os.Exit(1)
	}

	// Output result
	if *output == "" {
		fmt.Print(screen)
	} else {
		if err := os.WriteFile(*output, []byte(screen), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing to file: %v\n", err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "Screen content written to: %s\n", *output)
	}
}

func showHelp() {
	fmt.Printf(`vtermtest-cli - Terminal emulator testing tool

USAGE:
    vtermtest-cli --command "COMMAND" [OPTIONS]

OPTIONS:
    --command STRING    Command to execute (required)
    --keys STRING       Key sequence in DSL format
    --rows INT          Terminal rows (default: 24)
    --cols INT          Terminal columns (default: 80)
    --output FILE       Output file (default: stdout)
    --timeout DURATION  Screen stabilization timeout (default: 5s)
    --quiet DURATION    Stabilization quiet period (default: 100ms)
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

EXAMPLES:
    vtermtest-cli --command "echo hello"
    vtermtest-cli --command "sh -c 'read x; echo \$x'" --keys "test<Enter>"
    vtermtest-cli --command "vim" --keys "ihello<Esc>:wq<Enter>" --output screen.txt
    vtermtest-cli --command "sh -c 'sleep 1; echo Ready'" --keys "<WaitFor Ready>"
    vtermtest-cli --command "echo test" --keys "[WaitFor test]" --delimiter "[]"
`)
}

func parseCommand(cmd string) []string {
	// Simple command parsing - split by spaces but respect quotes
	var parts []string
	var current strings.Builder
	inQuotes := false
	quoteChar := byte(0)

	for i := 0; i < len(cmd); i++ {
		ch := cmd[i]

		if !inQuotes && (ch == '"' || ch == '\'') {
			inQuotes = true
			quoteChar = ch
		} else if inQuotes && ch == quoteChar {
			inQuotes = false
			quoteChar = 0
		} else if !inQuotes && ch == ' ' {
			if current.Len() > 0 {
				parts = append(parts, current.String())
				current.Reset()
			}
		} else {
			current.WriteByte(ch)
		}
	}

	if current.Len() > 0 {
		parts = append(parts, current.String())
	}

	return parts
}

func parseEnvVars(env string) []string {
	if env == "" {
		return nil
	}

	parts := strings.Split(env, ",")
	var result []string

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" && strings.Contains(part, "=") {
			result = append(result, part)
		}
	}

	return result
}

func parseDelimiter(delimiter string) (rune, rune, error) {
	if len(delimiter) != 2 {
		return 0, 0, fmt.Errorf("delimiter must be exactly 2 characters, got %d: %q", len(delimiter), delimiter)
	}

	runes := []rune(delimiter)
	if len(runes) != 2 {
		return 0, 0, fmt.Errorf("delimiter must contain exactly 2 Unicode characters: %q", delimiter)
	}

	return runes[0], runes[1], nil
}
