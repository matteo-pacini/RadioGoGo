//go:build devtools

// Copyright (c) 2023-2026 Matteo Pacini
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package devtools

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// keyNames maps every name produced by tea.KeyType.String() back to its type,
// so key names round-trip with the keybinding format in the config file.
var keyNames = func() map[string]tea.KeyType {
	names := map[string]tea.KeyType{"space": tea.KeySpace}
	// Bounds comfortably cover BubbleTea's negative special keys and the
	// 0-127 control range; unknown values stringify to "".
	for i := -256; i <= 256; i++ {
		kt := tea.KeyType(i)
		if kt == tea.KeyRunes {
			continue
		}
		if s := kt.String(); s != "" {
			names[s] = kt
		}
	}
	return names
}()

// parseKey converts a key name such as "enter", "ctrl+c", "alt+x" or "j" into
// the tea.KeyMsg a real keypress would produce.
func parseKey(name string) (tea.KeyMsg, error) {
	base, alt := name, false
	if rest, ok := strings.CutPrefix(name, "alt+"); ok && rest != "" {
		base, alt = rest, true
	}
	if kt, ok := keyNames[base]; ok {
		msg := tea.KeyMsg{Type: kt, Alt: alt}
		if kt == tea.KeySpace {
			msg.Runes = []rune{' '}
		}
		return msg, nil
	}
	if r := []rune(base); len(r) == 1 {
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: r, Alt: alt}, nil
	}
	return tea.KeyMsg{}, fmt.Errorf("unknown key %q", name)
}

// parseKeys parses all names, failing without a partial result if any is invalid.
func parseKeys(names []string) ([]tea.KeyMsg, error) {
	msgs := make([]tea.KeyMsg, 0, len(names))
	for _, name := range names {
		msg, err := parseKey(name)
		if err != nil {
			return nil, err
		}
		msgs = append(msgs, msg)
	}
	return msgs, nil
}

// textKeys returns one keypress per rune of text.
func textKeys(text string) []tea.KeyMsg {
	msgs := make([]tea.KeyMsg, 0, len(text))
	for _, r := range text {
		if r == ' ' {
			msgs = append(msgs, tea.KeyMsg{Type: tea.KeySpace, Runes: []rune{' '}})
			continue
		}
		msgs = append(msgs, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	return msgs
}
