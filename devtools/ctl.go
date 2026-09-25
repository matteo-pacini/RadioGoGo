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
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
)

const ctlUsage = `usage: radiogogo ctl <command>

commands:
  key <name>...                 press keys, e.g. key down down enter
  type <text>                   type text one character at a time
  screen                        print the screen as plain text
  state                         print the UI state as JSON
  wait <text> [--timeout <ms>]  wait until the screen contains text
  watch                         print one event per line until interrupted
`

// RunCtl runs `radiogogo ctl` with args (excluding "ctl") against the running
// instance and returns the process exit code: 0 on success, 1 when the
// instance reports an error or cannot be reached, 2 on invalid usage.
func RunCtl(args []string, stdout, stderr io.Writer) int {
	return runCtl(SocketPath(), args, stdout, stderr)
}

func runCtl(path string, args []string, stdout, stderr io.Writer) int {
	req, err := parseCtlArgs(args)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n\n%s", err, ctlUsage)
		return 2
	}

	conn, err := net.Dial("unix", path)
	if err != nil {
		fmt.Fprintf(stderr, "error: no radiogogo instance running (socket %s)\n", path)
		return 1
	}
	defer conn.Close()

	line, _ := json.Marshal(req)
	if _, err := conn.Write(append(line, '\n')); err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}

	if req.Op == "watch" {
		_, _ = io.Copy(stdout, conn)
		return 0
	}

	scanner := bufio.NewScanner(conn)
	scanner.Buffer(make([]byte, 64*1024), maxRequestBytes)
	if !scanner.Scan() {
		fmt.Fprintln(stderr, "error: connection closed without a response")
		return 1
	}
	var resp Response
	if err := json.Unmarshal(scanner.Bytes(), &resp); err != nil {
		fmt.Fprintf(stderr, "error: invalid response: %v\n", err)
		return 1
	}

	switch req.Op {
	case "screen", "wait":
		if resp.Screen != "" {
			fmt.Fprintln(stdout, resp.Screen)
		}
	case "state":
		if resp.State != nil {
			out, _ := json.MarshalIndent(resp.State, "", "  ")
			fmt.Fprintln(stdout, string(out))
		}
	}
	if !resp.OK {
		fmt.Fprintf(stderr, "error: %s\n", resp.Error)
		return 1
	}
	return 0
}

func parseCtlArgs(args []string) (Request, error) {
	if len(args) == 0 {
		return Request{}, errors.New("missing command")
	}
	cmd, rest := args[0], args[1:]
	switch cmd {
	case "key":
		if len(rest) == 0 {
			return Request{}, errors.New("key needs at least one key name")
		}
		return Request{Op: "key", Keys: rest}, nil
	case "type":
		if len(rest) == 0 {
			return Request{}, errors.New("type needs text")
		}
		return Request{Op: "type", Text: strings.Join(rest, " ")}, nil
	case "screen", "state", "watch":
		if len(rest) != 0 {
			return Request{}, fmt.Errorf("%s takes no arguments", cmd)
		}
		return Request{Op: cmd}, nil
	case "wait":
		req := Request{Op: "wait"}
		var text []string
		for i := 0; i < len(rest); i++ {
			if rest[i] != "--timeout" {
				text = append(text, rest[i])
				continue
			}
			if i+1 == len(rest) {
				return Request{}, errors.New("--timeout needs a value in milliseconds")
			}
			ms, err := strconv.Atoi(rest[i+1])
			if err != nil || ms <= 0 {
				return Request{}, fmt.Errorf("invalid --timeout %q", rest[i+1])
			}
			req.TimeoutMS = ms
			i++
		}
		if len(text) == 0 {
			return Request{}, errors.New("wait needs text")
		}
		req.Text = strings.Join(text, " ")
		return req, nil
	}
	return Request{}, fmt.Errorf("unknown command %q", cmd)
}
