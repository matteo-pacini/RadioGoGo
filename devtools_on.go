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

package main

import (
	"fmt"
	"os"

	"github.com/zi0p4tch0/radiogogo/devtools"
	"github.com/zi0p4tch0/radiogogo/models"

	tea "github.com/charmbracelet/bubbletea"
)

// runDevtoolsCtl runs `radiogogo ctl ...` and reports whether args requested it.
func runDevtoolsCtl(args []string) (exitCode int, handled bool) {
	if len(args) == 0 || args[0] != "ctl" {
		return 0, false
	}
	return devtools.RunCtl(args[1:], os.Stdout, os.Stderr), true
}

// setupDevtools opens the control socket and wraps model so it can be driven
// remotely. Call attach with the program before running it, and cleanup on exit.
// If the socket cannot be opened, a warning is printed and model is returned unwrapped.
func setupDevtools(model models.Model) (wrapped tea.Model, attach func(*tea.Program), cleanup func()) {
	srv, err := devtools.Listen(devtools.SocketPath())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: control socket disabled: %v\n", err)
		return model, func(*tea.Program) {}, func() {}
	}
	return devtools.Wrap(model), srv.Attach, srv.Close
}
