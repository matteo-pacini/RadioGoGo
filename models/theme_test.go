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

package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zi0p4tch0/radiogogo/config"
)

func TestNewTheme(t *testing.T) {
	cfg := config.NewDefaultConfig()
	theme := NewTheme(cfg)

	t.Run("creates non-nil styles", func(t *testing.T) {
		// Verify all style fields are set
		assert.NotEmpty(t, theme.SecondaryColor)
	})

	t.Run("stores secondary color", func(t *testing.T) {
		assert.Equal(t, cfg.Theme.SecondaryColor, theme.SecondaryColor)
	})
}

func TestStyleBottomBar(t *testing.T) {
	cfg := config.NewDefaultConfig()
	theme := NewTheme(cfg)

	t.Run("empty commands", func(t *testing.T) {
		result := theme.StyleBottomBar([]string{})
		assert.Equal(t, "", result)
	})

	t.Run("single command", func(t *testing.T) {
		result := theme.StyleBottomBar([]string{"q: quit"})
		assert.Contains(t, result, "q: quit")
	})

	t.Run("multiple commands", func(t *testing.T) {
		commands := []string{"q: quit", "enter: play", "↑/↓: move"}
		result := theme.StyleBottomBar(commands)

		for _, cmd := range commands {
			assert.Contains(t, result, cmd)
		}
	})
}

func TestStyleTwoRowBottomBar(t *testing.T) {
	cfg := config.NewDefaultConfig()
	theme := NewTheme(cfg)

	t.Run("renders both rows", func(t *testing.T) {
		primary := []string{"q: quit", "enter: play"}
		secondary := []string{"b: bookmark", "h: hide"}

		result := theme.StyleTwoRowBottomBar(primary, secondary)

		// Should contain all commands from both rows
		for _, cmd := range primary {
			assert.Contains(t, result, cmd)
		}
		for _, cmd := range secondary {
			assert.Contains(t, result, cmd)
		}

		// Should have a newline separating the two rows
		assert.Contains(t, result, "\n")
	})
}
