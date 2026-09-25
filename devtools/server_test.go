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
	"io"
	"net"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/zi0p4tch0/radiogogo/models"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// socketPathForTest returns a short path; Unix socket paths are limited to
// about 100 bytes and t.TempDir can exceed that.
func socketPathForTest(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "rgg")
	require.NoError(t, err)
	t.Cleanup(func() { os.RemoveAll(dir) })
	return filepath.Join(dir, socketName)
}

func TestSocketPath(t *testing.T) {
	t.Setenv("XDG_RUNTIME_DIR", "/run/user/1000")
	assert.Equal(t, "/run/user/1000/radiogogo-dev.sock", SocketPath())

	t.Setenv("XDG_RUNTIME_DIR", "")
	assert.Equal(t, filepath.Join(os.TempDir(), socketName), SocketPath())
}

func TestListen(t *testing.T) {
	t.Run("fresh start uses mode 0600 and close removes the file", func(t *testing.T) {
		path := socketPathForTest(t)
		srv, err := Listen(path)
		require.NoError(t, err)

		info, err := os.Stat(path)
		require.NoError(t, err)
		assert.Equal(t, os.ModeSocket|0o600, info.Mode())

		srv.Close()
		_, err = os.Stat(path)
		assert.ErrorIs(t, err, os.ErrNotExist)
	})

	t.Run("stale socket file is replaced", func(t *testing.T) {
		path := socketPathForTest(t)
		stale, err := net.ListenUnix("unix", &net.UnixAddr{Name: path, Net: "unix"})
		require.NoError(t, err)
		stale.SetUnlinkOnClose(false)
		stale.Close()
		_, err = os.Stat(path)
		require.NoError(t, err, "stale file should remain")

		srv, err := Listen(path)
		require.NoError(t, err)
		srv.Close()
	})

	t.Run("live instance is detected", func(t *testing.T) {
		path := socketPathForTest(t)
		first, err := Listen(path)
		require.NoError(t, err)
		defer first.Close()

		_, err = Listen(path)
		assert.ErrorIs(t, err, ErrInstanceRunning)
	})
}

// startProgram runs a real program around a fake model and serves it.
func startProgram(t *testing.T) string {
	t.Helper()
	path := socketPathForTest(t)
	srv, err := Listen(path)
	require.NoError(t, err)

	p := tea.NewProgram(Wrap(newFakeModel()), tea.WithInput(nil), tea.WithOutput(io.Discard), tea.WithoutRenderer())
	srv.Attach(p)
	finished := make(chan struct{})
	go func() {
		_, _ = p.Run()
		close(finished)
	}()
	t.Cleanup(func() {
		p.Quit()
		<-finished
		srv.Close()
	})
	return path
}

type testClient struct {
	conn    net.Conn
	scanner *bufio.Scanner
}

func dial(t *testing.T, path string) *testClient {
	t.Helper()
	conn, err := net.Dial("unix", path)
	require.NoError(t, err)
	t.Cleanup(func() { conn.Close() })
	return &testClient{conn: conn, scanner: bufio.NewScanner(conn)}
}

func (c *testClient) raw(t *testing.T, line string) Response {
	t.Helper()
	_, err := c.conn.Write([]byte(line + "\n"))
	require.NoError(t, err)
	require.True(t, c.scanner.Scan(), "no response")
	var resp Response
	require.NoError(t, json.Unmarshal(c.scanner.Bytes(), &resp))
	return resp
}

func (c *testClient) do(t *testing.T, req Request) Response {
	t.Helper()
	line, err := json.Marshal(req)
	require.NoError(t, err)
	return c.raw(t, string(line))
}

func TestServer_Ops(t *testing.T) {
	path := startProgram(t)
	c := dial(t, path)

	resp := c.do(t, Request{Op: "key", Keys: []string{"a", "enter"}})
	assert.True(t, resp.OK, resp.Error)
	assert.Equal(t, "typed:aenter", c.do(t, Request{Op: "screen"}).Screen)

	resp = c.do(t, Request{Op: "key", Keys: []string{"b", "hyperspace"}})
	assert.False(t, resp.OK)
	assert.Contains(t, resp.Error, "hyperspace")
	assert.Equal(t, "typed:aenter", c.do(t, Request{Op: "screen"}).Screen, "no key from a rejected request is delivered")

	assert.True(t, c.do(t, Request{Op: "type", Text: "x y"}).OK)
	assert.Equal(t, "typed:aenterx y", c.do(t, Request{Op: "screen"}).Screen)

	resp = c.do(t, Request{Op: "state"})
	require.True(t, resp.OK)
	assert.Equal(t, &State{View: "search", Width: 120, Height: 40}, resp.State)

	resp = c.do(t, Request{Op: "fly"})
	assert.False(t, resp.OK)
	assert.Contains(t, resp.Error, `"fly"`)

	resp = c.raw(t, "{not json")
	assert.False(t, resp.OK)
	assert.Contains(t, resp.Error, "malformed")

	assert.True(t, c.do(t, Request{Op: "state"}).OK, "connection stays usable after errors")
}

func TestServer_Wait(t *testing.T) {
	path := startProgram(t)

	t.Run("resolves when another client changes the screen", func(t *testing.T) {
		waiter, typist := dial(t, path), dial(t, path)
		result := make(chan Response, 1)
		go func() {
			line, _ := json.Marshal(Request{Op: "wait", Text: "typed:z", TimeoutMS: 5000})
			waiter.conn.Write(append(line, '\n'))
			waiter.scanner.Scan()
			var resp Response
			json.Unmarshal(waiter.scanner.Bytes(), &resp)
			result <- resp
		}()
		time.Sleep(50 * time.Millisecond)
		typist.do(t, Request{Op: "type", Text: "z"})

		resp := <-result
		assert.True(t, resp.OK, resp.Error)
		assert.Equal(t, "typed:z", resp.Screen)
	})

	t.Run("times out with the current screen", func(t *testing.T) {
		c := dial(t, path)
		start := time.Now()
		resp := c.do(t, Request{Op: "wait", Text: "never", TimeoutMS: 200})
		assert.False(t, resp.OK)
		assert.Contains(t, resp.Error, "timed out")
		assert.Equal(t, "typed:z", resp.Screen)
		assert.GreaterOrEqual(t, time.Since(start), 200*time.Millisecond)
	})

	t.Run("requires text", func(t *testing.T) {
		assert.False(t, dial(t, path).do(t, Request{Op: "wait"}).OK)
	})
}

func TestServer_Watch(t *testing.T) {
	path := startProgram(t)
	watcher := dial(t, path)
	_, err := watcher.conn.Write([]byte(`{"op":"watch"}` + "\n"))
	require.NoError(t, err)

	c := dial(t, path)
	// The watch registration and the keys travel on different connections, so
	// keep pressing until the registered stream reports one.
	for range 20 {
		c.do(t, Request{Op: "key", Keys: []string{"q"}})
		require.NoError(t, watcher.conn.SetReadDeadline(time.Now().Add(100*time.Millisecond)))
		for watcher.scanner.Scan() {
			var e Event
			require.NoError(t, json.Unmarshal(watcher.scanner.Bytes(), &e))
			if e.Name == "tea.KeyMsg" {
				return
			}
		}
		watcher.scanner = bufio.NewScanner(watcher.conn)
	}
	t.Fatal("no key event on the watch stream")
}

// recordingSender captures messages sent to the program.
type recordingSender struct {
	mu   sync.Mutex
	msgs []tea.Msg
}

func (r *recordingSender) Send(msg tea.Msg) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.msgs = append(r.msgs, msg)
}

func (r *recordingSender) counts() []int {
	r.mu.Lock()
	defer r.mu.Unlock()
	var counts []int
	for _, msg := range r.msgs {
		if m, ok := msg.(models.RemoteClientsChangedMsg); ok {
			counts = append(counts, m.Count)
		}
	}
	return counts
}

func TestServer_ClientCount(t *testing.T) {
	srv, err := Listen(socketPathForTest(t))
	require.NoError(t, err)
	defer srv.Close()
	rec := &recordingSender{}
	srv.attach(rec)

	waitCounts := func(want ...int) {
		t.Helper()
		assert.Eventually(t, func() bool { return assert.ObjectsAreEqual(want, rec.counts()) }, time.Second, 5*time.Millisecond, "got %v", rec.counts())
	}

	first, err := net.Dial("unix", srv.path)
	require.NoError(t, err)
	waitCounts(1)
	second, err := net.Dial("unix", srv.path)
	require.NoError(t, err)
	waitCounts(1, 2)
	first.Close()
	waitCounts(1, 2, 1)
	second.Close()
	waitCounts(1, 2, 1, 0)
}
