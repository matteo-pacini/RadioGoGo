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
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/stretchr/testify/assert"
)

func TestRenderFiller(t *testing.T) {

	t.Run("positive height", func(t *testing.T) {
		filler := RenderFiller(5)
		assert.Equal(t, 5, strings.Count(filler, "\n"))
	})

	t.Run("zero height", func(t *testing.T) {
		filler := RenderFiller(0)
		assert.Equal(t, "", filler)
	})

	t.Run("negative height", func(t *testing.T) {
		filler := RenderFiller(-5)
		assert.Equal(t, "", filler)
	})

	t.Run("single line", func(t *testing.T) {
		filler := RenderFiller(1)
		assert.Equal(t, "\n", filler)
	})

}

func TestCalculateFillerHeight(t *testing.T) {

	t.Run("basic calculation", func(t *testing.T) {
		// terminal=40, headerContent=10, bottomBar=1
		// filler = 40 - 10 - 1 + 1 = 30
		filler := CalculateFillerHeight(40, 10, 1)
		assert.Equal(t, 30, filler)
	})

	t.Run("with two-row bottom bar", func(t *testing.T) {
		// terminal=40, headerContent=10, bottomBar=2
		// filler = 40 - 10 - 2 + 1 = 29
		filler := CalculateFillerHeight(40, 10, 2)
		assert.Equal(t, 29, filler)
	})

	t.Run("no filler needed", func(t *testing.T) {
		// terminal=40, headerContent=39, bottomBar=1
		// filler = 40 - 39 - 1 + 1 = 1
		filler := CalculateFillerHeight(40, 39, 1)
		assert.Equal(t, 1, filler)
	})

	t.Run("content exceeds space", func(t *testing.T) {
		// terminal=40, headerContent=50, bottomBar=1
		// filler = 40 - 50 - 1 + 1 = -10 -> clamped to 0
		filler := CalculateFillerHeight(40, 50, 1)
		assert.Equal(t, 0, filler)
	})

}

// TestActualLayoutHeight tests that when we simulate the actual view construction,
// the final lipgloss.Height equals terminal height
func TestActualLayoutHeight(t *testing.T) {

	t.Run("single-row bottom bar layout", func(t *testing.T) {
		terminalHeight := 40

		// Simulate header (ends with \n like real header)
		header := "radiogogo v1.0\n"

		// Simulate content (some text, may or may not end with \n)
		content := "Search\nInput\nSelector\n"

		// Build view like model.go does
		view := header + content
		headerContentHeight := lipgloss.Height(view)

		// Calculate filler
		bottomBarHeight := 1
		fillerHeight := CalculateFillerHeight(terminalHeight, headerContentHeight, bottomBarHeight)
		view += RenderFiller(fillerHeight)

		// Add bottom bar (no trailing newline)
		bottomBar := "q: quit | tab: cycle"
		view += bottomBar

		// Verify total height
		assert.Equal(t, terminalHeight, lipgloss.Height(view))
	})

	t.Run("two-row bottom bar layout", func(t *testing.T) {
		terminalHeight := 40

		// Simulate header
		header := "radiogogo v1.0 (●) ffplay\n"

		// Simulate content
		content := "Station 1\nStation 2\nStation 3\n"

		// Build view
		view := header + content
		headerContentHeight := lipgloss.Height(view)

		// Calculate filler
		bottomBarHeight := 2
		fillerHeight := CalculateFillerHeight(terminalHeight, headerContentHeight, bottomBarHeight)
		view += RenderFiller(fillerHeight)

		// Add two-row bottom bar
		bottomBar := "q: quit | enter: play\nb: bookmark | h: hide"
		view += bottomBar

		// Verify total height
		assert.Equal(t, terminalHeight, lipgloss.Height(view))
	})

	t.Run("content without trailing newline", func(t *testing.T) {
		terminalHeight := 40

		header := "header\n"
		content := "content" // No trailing \n

		view := header + content
		headerContentHeight := lipgloss.Height(view)

		bottomBarHeight := 1
		fillerHeight := CalculateFillerHeight(terminalHeight, headerContentHeight, bottomBarHeight)
		view += RenderFiller(fillerHeight)

		bottomBar := "bottom"
		view += bottomBar

		assert.Equal(t, terminalHeight, lipgloss.Height(view))
	})

	t.Run("various terminal sizes", func(t *testing.T) {
		sizes := []int{31, 40, 50, 60, 100}

		for _, terminalHeight := range sizes {
			header := "header\n"
			content := "line1\nline2\nline3\n"

			view := header + content
			headerContentHeight := lipgloss.Height(view)

			// Single-row bar
			fillerHeight := CalculateFillerHeight(terminalHeight, headerContentHeight, 1)
			finalView := view + RenderFiller(fillerHeight) + "bottom"
			assert.Equal(t, terminalHeight, lipgloss.Height(finalView),
				"single-row bar with terminal height %d", terminalHeight)

			// Two-row bar
			fillerHeight = CalculateFillerHeight(terminalHeight, headerContentHeight, 2)
			finalView = view + RenderFiller(fillerHeight) + "bottom1\nbottom2"
			assert.Equal(t, terminalHeight, lipgloss.Height(finalView),
				"two-row bar with terminal height %d", terminalHeight)
		}
	})

}

// TestLipglossHeightBehavior documents how lipgloss.Height works
func TestLipglossHeightBehavior(t *testing.T) {

	t.Run("empty string", func(t *testing.T) {
		assert.Equal(t, 1, lipgloss.Height(""))
	})

	t.Run("single char", func(t *testing.T) {
		assert.Equal(t, 1, lipgloss.Height("a"))
	})

	t.Run("char with newline", func(t *testing.T) {
		// "a\n" = 2 lines (a on line 1, empty line 2)
		assert.Equal(t, 2, lipgloss.Height("a\n"))
	})

	t.Run("two lines no trailing", func(t *testing.T) {
		assert.Equal(t, 2, lipgloss.Height("a\nb"))
	})

	t.Run("two lines with trailing", func(t *testing.T) {
		assert.Equal(t, 3, lipgloss.Height("a\nb\n"))
	})

	t.Run("just newlines", func(t *testing.T) {
		assert.Equal(t, 2, lipgloss.Height("\n"))
		assert.Equal(t, 3, lipgloss.Height("\n\n"))
		assert.Equal(t, 4, lipgloss.Height("\n\n\n"))
	})

}
