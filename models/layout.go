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

import "strings"

// LayoutConfig holds terminal dimensions
type LayoutConfig struct {
	TerminalWidth  int
	TerminalHeight int
}

// LayoutResult contains calculated heights for each component
type LayoutResult struct {
	HeaderHeight    int
	ContentHeight   int
	BottomBarHeight int
	FillerHeight    int
}

// CalculateFillerHeight computes the filler height needed to fill the terminal.
// headerContentHeight should be lipgloss.Height of the already-concatenated header + content.
// bottomBarHeight is 1 for single-row bar, 2 for two-row bar.
//
// The formula accounts for how lipgloss.Height counts trailing newlines:
// - When filler (newlines) is followed by bottomBar content, the last newline "merges"
// - So filler of N newlines contributes N visual lines (not N+1)
// - Total = headerContentHeight + fillerHeight + bottomBarHeight - 1 (one merge)
// - Therefore: fillerHeight = terminalHeight - headerContentHeight - bottomBarHeight + 1
func CalculateFillerHeight(terminalHeight, headerContentHeight, bottomBarHeight int) int {
	filler := terminalHeight - headerContentHeight - bottomBarHeight + 1
	if filler < 0 {
		filler = 0
	}
	return filler
}

// CalculateLayout computes heights for single-row bottom bar states (search, loading, error)
// DEPRECATED: Use CalculateFillerHeight instead for accurate calculations
func CalculateLayout(config LayoutConfig, contentLines int) LayoutResult {
	headerHeight := 1
	bottomBarHeight := 1
	// Use the new formula: terminal - actualHeight - bottomBar + 1
	// Here we approximate actualHeight as headerHeight + contentLines, but this is imprecise
	// because it doesn't account for trailing newline merge effects
	fillerHeight := config.TerminalHeight - headerHeight - contentLines - bottomBarHeight + 1
	if fillerHeight < 0 {
		fillerHeight = 0
	}
	return LayoutResult{
		HeaderHeight:    headerHeight,
		ContentHeight:   config.TerminalHeight - headerHeight - bottomBarHeight,
		BottomBarHeight: bottomBarHeight,
		FillerHeight:    fillerHeight,
	}
}

// CalculateStationsLayout computes heights for two-row bottom bar (stations state)
// DEPRECATED: Use CalculateFillerHeight instead for accurate calculations
func CalculateStationsLayout(config LayoutConfig, contentLines int) LayoutResult {
	headerHeight := 1
	bottomBarHeight := 2
	fillerHeight := config.TerminalHeight - headerHeight - contentLines - bottomBarHeight + 1
	if fillerHeight < 0 {
		fillerHeight = 0
	}
	return LayoutResult{
		HeaderHeight:    headerHeight,
		ContentHeight:   config.TerminalHeight - headerHeight - bottomBarHeight,
		BottomBarHeight: bottomBarHeight,
		FillerHeight:    fillerHeight,
	}
}

// RenderFiller returns the exact newlines needed for filler space
func RenderFiller(height int) string {
	if height <= 0 {
		return ""
	}
	return strings.Repeat("\n", height)
}
