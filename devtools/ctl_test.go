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
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func runCtlForTest(path string, args ...string) (code int, stdout, stderr string) {
	var out, errOut bytes.Buffer
	code = runCtl(path, args, &out, &errOut)
	return code, out.String(), errOut.String()
}

func TestParseCtlArgs(t *testing.T) {
	tests := []struct {
		args []string
		want Request
	}{
		{[]string{"key", "down", "enter"}, Request{Op: "key", Keys: []string{"down", "enter"}}},
		{[]string{"type", "hello", "world"}, Request{Op: "type", Text: "hello world"}},
		{[]string{"screen"}, Request{Op: "screen"}},
		{[]string{"state"}, Request{Op: "state"}},
		{[]string{"watch"}, Request{Op: "watch"}},
		{[]string{"wait", "Now", "playing"}, Request{Op: "wait", Text: "Now playing"}},
		{[]string{"wait", "x", "--timeout", "300"}, Request{Op: "wait", Text: "x", TimeoutMS: 300}},
		{[]string{"wait", "--timeout", "300", "x"}, Request{Op: "wait", Text: "x", TimeoutMS: 300}},
	}
	for _, tt := range tests {
		got, err := parseCtlArgs(tt.args)
		require.NoError(t, err, tt.args)
		assert.Equal(t, tt.want, got)
	}

	for _, args := range [][]string{
		nil, {"fly"}, {"key"}, {"type"}, {"screen", "x"},
		{"wait"}, {"wait", "x", "--timeout"}, {"wait", "x", "--timeout", "soon"},
	} {
		_, err := parseCtlArgs(args)
		assert.Error(t, err, args)
	}
}

func TestRunCtl(t *testing.T) {
	path := startProgram(t)

	code, _, _ := runCtlForTest(path, "type", "ab")
	assert.Equal(t, 0, code)

	code, out, _ := runCtlForTest(path, "screen")
	assert.Equal(t, 0, code)
	assert.Equal(t, "typed:ab\n", out)

	code, out, _ = runCtlForTest(path, "state")
	assert.Equal(t, 0, code)
	var state State
	require.NoError(t, json.Unmarshal([]byte(out), &state))
	assert.Equal(t, "search", state.View)

	code, _, errOut := runCtlForTest(path, "key", "hyperspace")
	assert.Equal(t, 1, code)
	assert.Contains(t, errOut, "hyperspace")

	code, out, errOut = runCtlForTest(path, "wait", "never", "--timeout", "100")
	assert.Equal(t, 1, code)
	assert.Contains(t, errOut, "timed out")
	assert.Equal(t, "typed:ab\n", out)

	code, _, _ = runCtlForTest(path, "bogus")
	assert.Equal(t, 2, code)
}

func TestRunCtl_NoInstance(t *testing.T) {
	code, _, errOut := runCtlForTest(socketPathForTest(t), "screen")
	assert.Equal(t, 1, code)
	assert.Contains(t, errOut, "no radiogogo instance running")
}
