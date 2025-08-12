package vtermtest

import (
	"strings"

	libvterm "github.com/mattn/go-libvterm"
	"github.com/mattn/go-runewidth"
)

// GetScreenText returns the entire terminal screen as a string.
// Lines are trimmed of trailing spaces and joined with newlines.
func (e *Emulator) GetScreenText() (string, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.screen == nil {
		return "", nil
	}

	lines := make([]string, e.rows)
	for row := 0; row < int(e.rows); row++ {
		line := e.getLine(row)
		lines[row] = strings.TrimRight(line, " ")
	}

	return strings.Join(lines, "\n"), nil
}

func (e *Emulator) getLine(row int) string {
	var line strings.Builder
	currentCol := 0

	for col := 0; col < int(e.cols); {
		pos := libvterm.NewPos(row, col)
		cell, err := e.screen.GetCell(pos)
		
		if err != nil || cell == nil {
			line.WriteRune(' ')
			currentCol++
			col++
			continue
		}

		chars := cell.Chars()
		if len(chars) == 0 || chars[0] == 0 {
			line.WriteRune(' ')
			currentCol++
			col++
			continue
		}

		r := chars[0]
		line.WriteRune(r)
		
		width := runewidth.RuneWidth(r)
		if width == 0 {
			width = 1
		}
		
		currentCol += width
		col += width
	}

	return line.String()
}

// GetLine returns a specific line from the terminal screen.
// Row index starts at 0. Trailing spaces are trimmed.
func (e *Emulator) GetLine(row int) (string, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.screen == nil || row >= int(e.rows) {
		return "", nil
	}

	line := e.getLine(row)
	return strings.TrimRight(line, " "), nil
}