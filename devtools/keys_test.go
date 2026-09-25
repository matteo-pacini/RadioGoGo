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
	"testing"

	"github.com/zi0p4tch0/radiogogo/config"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseKey_RoundTripsEveryNamedKey(t *testing.T) {
	require.Greater(t, len(keyNames), 50)
	for name := range keyNames {
		if name == "space" {
			continue
		}
		msg, err := parseKey(name)
		require.NoError(t, err, name)
		assert.Equal(t, name, msg.String())
	}
}

func TestParseKey(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"j", "j"},
		{"B", "B"},
		{"9", "9"},
		{"é", "é"},
		{"enter", "enter"},
		{"ctrl+k", "ctrl+k"},
		{"alt+x", "alt+x"},
		{"alt+enter", "alt+enter"},
		{"space", " "},
		{" ", " "},
		{"+", "+"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg, err := parseKey(tt.name)
			require.NoError(t, err)
			assert.Equal(t, tt.want, msg.String())
		})
	}
}

func TestParseKey_Rejects(t *testing.T) {
	for _, name := range []string{"hyperspace", "", "alt+", "runes", "ctrl+nope"} {
		_, err := parseKey(name)
		assert.Error(t, err, name)
	}
}

func TestParseKeys_AllOrNothing(t *testing.T) {
	msgs, err := parseKeys([]string{"down", "hyperspace"})
	assert.ErrorContains(t, err, "hyperspace")
	assert.Nil(t, msgs)
}

func TestParseKey_DefaultKeybindings(t *testing.T) {
	kb := config.NewDefaultConfig().Keybindings
	for _, key := range []string{
		kb.Quit, kb.Search, kb.Record, kb.BookmarkToggle, kb.BookmarksView,
		kb.HideStation, kb.ManageHidden, kb.ChangeLanguage, kb.VolumeDown,
		kb.VolumeUp, kb.NavigateDown, kb.NavigateUp, kb.StopPlayback, kb.Vote,
	} {
		msg, err := parseKey(key)
		require.NoError(t, err, key)
		assert.Equal(t, key, msg.String())
	}
}

func TestTextKeys(t *testing.T) {
	msgs := textKeys("a b")
	require.Len(t, msgs, 3)
	assert.Equal(t, tea.KeyRunes, msgs[0].Type)
	assert.Equal(t, tea.KeySpace, msgs[1].Type)
	assert.Equal(t, "a", msgs[0].String())
	assert.Equal(t, " ", msgs[1].String())
}
