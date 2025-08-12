package keys

import (
	"bytes"
	"reflect"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected [][]byte
		wantErr  bool
	}{
		{
			name:     "simple text",
			input:    "hello",
			expected: [][]byte{Text("hello")},
		},
		{
			name:     "text with tab",
			input:    "hello<Tab>world",
			expected: [][]byte{Text("hello"), Tab, Text("world")},
		},
		{
			name:     "ctrl keys",
			input:    "<C-a><C-c><C-z>",
			expected: [][]byte{CtrlA, CtrlC, CtrlZ},
		},
		{
			name:     "alt keys",
			input:    "<A-a><A-f>",
			expected: [][]byte{Alt('a'), Alt('f')},
		},
		{
			name:     "function keys",
			input:    "<F1><F12><F24>",
			expected: [][]byte{F(1), F(12), F(24)},
		},
		{
			name:     "navigation keys",
			input:    "<Up><Down><Left><Right><Home><End>",
			expected: [][]byte{Up, Down, Left, Right, Home, End},
		},
		{
			name:     "special keys",
			input:    "<Enter><BS><Del><Esc><Space>",
			expected: [][]byte{Enter, Backspace, Delete, []byte{0x1B}, []byte{' '}},
		},
		{
			name:     "escaped angle bracket",
			input:    "hello<<world",
			expected: [][]byte{Text("hello<world")},
		},
		{
			name:  "complex example",
			input: "SELECT * FROM us<Tab><C-a>deleted<C-k>",
			expected: [][]byte{
				Text("SELECT * FROM us"),
				Tab,
				CtrlA,
				Text("deleted"),
				CtrlK,
			},
		},
		{
			name:  "vim-like example",
			input: "ihello world<Esc>:wq<Enter>",
			expected: [][]byte{
				Text("ihello world"),
				[]byte{0x1B},
				Text(":wq"),
				Enter,
			},
		},
		{
			name:     "case insensitive keys",
			input:    "<tab><ENTER><bs><DEL>",
			expected: [][]byte{Tab, Enter, Backspace, Delete},
		},
		{
			name:     "alternative names",
			input:    "<CR><Backspace><Escape>",
			expected: [][]byte{Enter, Backspace, []byte{0x1B}},
		},
		{
			name:    "unclosed bracket",
			input:   "hello<Tab",
			wantErr: true,
		},
		{
			name:    "unknown key",
			input:   "<Unknown>",
			wantErr: true,
		},
		{
			name:    "invalid ctrl key",
			input:   "<C-1>",
			wantErr: true,
		},
		{
			name:    "invalid function key",
			input:   "<F25>",
			wantErr: true,
		},
		{
			name:    "invalid function key format",
			input:   "<Fabc>",
			wantErr: true,
		},
		{
			name:     "empty input",
			input:    "",
			expected: nil,
		},
		{
			name:     "only special keys",
			input:    "<Tab><Enter>",
			expected: [][]byte{Tab, Enter},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Parse(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Parse() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Parse() unexpected error: %v", err)
				return
			}

			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("Parse() result mismatch")
				t.Errorf("Expected: %v", tt.expected)
				t.Errorf("Got:      %v", result)

				// More detailed comparison for debugging
				if len(result) != len(tt.expected) {
					t.Errorf("Length mismatch: expected %d, got %d", len(tt.expected), len(result))
				} else {
					for i := range result {
						if !bytes.Equal(result[i], tt.expected[i]) {
							t.Errorf("Mismatch at index %d: expected %v, got %v", i, tt.expected[i], result[i])
						}
					}
				}
			}
		})
	}
}

func TestParseSpecialKey(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []byte
		wantErr  bool
	}{
		{"tab", "tab", Tab, false},
		{"Tab", "Tab", Tab, false},
		{"TAB", "TAB", Tab, false},
		{"ctrl-a", "C-a", CtrlA, false},
		{"ctrl-z", "C-z", CtrlZ, false},
		{"alt-a", "A-a", Alt('a'), false},
		{"alt-f", "A-f", Alt('f'), false},
		{"f1", "F1", F(1), false},
		{"f24", "F24", F(24), false},
		{"enter", "enter", Enter, false},
		{"cr", "cr", Enter, false},
		{"backspace", "backspace", Backspace, false},
		{"bs", "bs", Backspace, false},
		{"delete", "delete", Delete, false},
		{"del", "del", Delete, false},
		{"escape", "escape", []byte{0x1B}, false},
		{"esc", "esc", []byte{0x1B}, false},
		{"space", "space", []byte{' '}, false},
		{"up", "up", Up, false},
		{"down", "down", Down, false},
		{"left", "left", Left, false},
		{"right", "right", Right, false},
		{"home", "home", Home, false},
		{"end", "end", End, false},
		{"pageup", "pageup", PageUp, false},
		{"pagedown", "pagedown", PageDown, false},

		// Error cases
		{"unknown", "unknown", nil, true},
		{"invalid-ctrl", "C-1", nil, true},
		{"invalid-alt", "A-1", nil, true},
		{"invalid-function", "F25", nil, true},
		{"invalid-function-format", "Fabc", nil, true},
		{"empty", "", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseSpecialKey(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("parseSpecialKey() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("parseSpecialKey() unexpected error: %v", err)
				return
			}

			if !bytes.Equal(result, tt.expected) {
				t.Errorf("parseSpecialKey() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

// Test that demonstrates the DSL usage examples from the specification
func TestDSLExamples(t *testing.T) {
	tests := []struct {
		name  string
		input string
		desc  string
	}{
		{
			name:  "go-prompt completion",
			input: "SELECT * FROM us<Tab>",
			desc:  "SQL completion example",
		},
		{
			name:  "vim-like editing",
			input: "ihello world<Esc>:wq<Enter>",
			desc:  "Vim insert mode, escape, save and quit",
		},
		{
			name:  "complex editing",
			input: "<C-a>deleted<C-k>new text<Enter>",
			desc:  "Go to beginning, type, kill line, type more, enter",
		},
		{
			name:  "escaped brackets",
			input: "echo <<angle bracket>>",
			desc:  "Text with literal angle brackets",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Parse(tt.input)
			if err != nil {
				t.Errorf("Parse(%q) failed: %v", tt.input, err)
				return
			}

			// Just verify it parses without error and produces some output
			if len(result) == 0 {
				t.Errorf("Parse(%q) produced no output", tt.input)
			}

			t.Logf("DSL: %q -> %d key sequences (%s)", tt.input, len(result), tt.desc)
		})
	}
}
