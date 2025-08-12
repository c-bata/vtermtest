package keys

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

// Parse converts DSL string to key sequences.
// Example: "hello<Tab>world<C-c>" -> [Text("hello"), Tab, Text("world"), CtrlC]
//
// DSL notation:
//   - Regular text: typed as-is
//   - Special keys: <Tab> <Enter> <BS> <Del> <Esc> <Space>
//   - Arrow keys: <Up> <Down> <Left> <Right>
//   - Ctrl keys: <C-a> ... <C-z>
//   - Alt keys: <A-a> ... <A-z>
//   - Function keys: <F1> ... <F24>
//   - Navigation: <Home> <End> <PageUp> <PageDown>
//   - Escape: << for literal <
func Parse(dsl string) ([][]byte, error) {
	var result [][]byte
	var text strings.Builder

	for i := 0; i < len(dsl); i++ {
		if dsl[i] == '<' {
			// Check for escaped < (<<)
			if i+1 < len(dsl) && dsl[i+1] == '<' {
				text.WriteByte('<')
				i++ // Skip the second <
				continue
			}

			// Flush accumulated text
			if text.Len() > 0 {
				result = append(result, Text(text.String()))
				text.Reset()
			}

			// Find closing >
			end := strings.IndexByte(dsl[i+1:], '>')
			if end == -1 {
				return nil, fmt.Errorf("unclosed '<' at position %d", i)
			}

			keyName := dsl[i+1 : i+1+end]
			key, err := parseSpecialKey(keyName)
			if err != nil {
				return nil, fmt.Errorf("at position %d: %w", i, err)
			}

			result = append(result, key)
			i += end + 1 // Skip past the >
		} else {
			text.WriteByte(dsl[i])
		}
	}

	// Flush remaining text
	if text.Len() > 0 {
		result = append(result, Text(text.String()))
	}

	return result, nil
}

func parseSpecialKey(name string) ([]byte, error) {
	// Handle basic special keys
	switch strings.ToLower(name) {
	case "tab":
		return Tab, nil
	case "enter", "cr":
		return Enter, nil
	case "bs", "backspace":
		return Backspace, nil
	case "del", "delete":
		return Delete, nil
	case "esc", "escape":
		return []byte{0x1B}, nil
	case "space":
		return []byte{' '}, nil
	case "up":
		return Up, nil
	case "down":
		return Down, nil
	case "left":
		return Left, nil
	case "right":
		return Right, nil
	case "home":
		return Home, nil
	case "end":
		return End, nil
	case "pageup":
		return PageUp, nil
	case "pagedown":
		return PageDown, nil
	}

	// Handle Ctrl-X format (C-a, C-b, etc.)
	if strings.HasPrefix(strings.ToLower(name), "c-") && len(name) == 3 {
		ch := unicode.ToLower(rune(name[2]))
		if ch >= 'a' && ch <= 'z' {
			return []byte{byte(ch - 'a' + 1)}, nil
		}
		return nil, fmt.Errorf("invalid ctrl key: <%s>", name)
	}

	// Handle Alt-X format (A-a, A-b, etc.)
	if strings.HasPrefix(strings.ToLower(name), "a-") && len(name) == 3 {
		ch := rune(name[2])
		// Only allow letters for Alt combinations
		if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') {
			return Alt(ch), nil
		}
		return nil, fmt.Errorf("invalid alt key: <%s>", name)
	}

	// Handle Function keys (F1-F24)
	if strings.HasPrefix(strings.ToUpper(name), "F") {
		numStr := name[1:]
		n, err := strconv.Atoi(numStr)
		if err != nil {
			return nil, fmt.Errorf("invalid function key: <%s>", name)
		}
		key := F(n)
		if key == nil {
			return nil, fmt.Errorf("function key out of range: <%s> (valid: F1-F24)", name)
		}
		return key, nil
	}

	return nil, fmt.Errorf("unknown key: <%s>", name)
}
