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
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/zi0p4tch0/radiogogo/models"

	tea "github.com/charmbracelet/bubbletea"
)

const (
	socketName         = "radiogogo-dev.sock"
	defaultWaitTimeout = 5 * time.Second
	watchBuffer        = 256
	maxRequestBytes    = 1 << 20
)

// ErrInstanceRunning is returned by Listen when another instance already
// accepts connections on the socket.
var ErrInstanceRunning = errors.New("another radiogogo instance owns the control socket")

// SocketPath returns $XDG_RUNTIME_DIR/radiogogo-dev.sock, or the same name in
// the OS temporary directory when XDG_RUNTIME_DIR is unset.
func SocketPath() string {
	dir := os.Getenv("XDG_RUNTIME_DIR")
	if dir == "" {
		dir = os.TempDir()
	}
	return filepath.Join(dir, socketName)
}

// sender is the part of *tea.Program the server uses.
type sender interface {
	Send(msg tea.Msg)
}

// Server accepts control connections and relays them to a program whose model
// was created with Wrap.
type Server struct {
	path     string
	listener net.Listener
	program  sender
	done     chan struct{}
	close    sync.Once

	clientsMu sync.Mutex
	clients   int
	nextID    atomic.Int64
}

// Listen claims the socket at path. A leftover file that nobody accepts
// connections on is replaced. If a live instance owns it, Listen returns
// ErrInstanceRunning.
func Listen(path string) (*Server, error) {
	if conn, err := net.DialTimeout("unix", path, 500*time.Millisecond); err == nil {
		conn.Close()
		return nil, ErrInstanceRunning
	}
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	listener, err := net.Listen("unix", path)
	if err != nil {
		return nil, err
	}
	// Both candidate directories are private to the user, so the window between
	// Listen and Chmod does not expose the socket to other users.
	if err := os.Chmod(path, 0o600); err != nil {
		listener.Close()
		return nil, err
	}
	return &Server{path: path, listener: listener, done: make(chan struct{})}, nil
}

// Attach starts serving connections for p. It must be called once, before
// p.Run, with the program whose model was created with Wrap.
func (s *Server) Attach(p *tea.Program) {
	s.attach(p)
}

func (s *Server) attach(p sender) {
	s.program = p
	go s.serve()
}

// Close stops accepting connections, ends open streams and removes the socket file.
func (s *Server) Close() {
	s.close.Do(func() {
		close(s.done)
		s.listener.Close()
		_ = os.Remove(s.path)
	})
}

func (s *Server) serve() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			return
		}
		go s.handle(conn)
	}
}

// changeClients adjusts the client count and tells the model. The mutex keeps
// the counts reaching the model in the same order as the changes.
func (s *Server) changeClients(delta int) {
	s.clientsMu.Lock()
	defer s.clientsMu.Unlock()
	s.clients += delta
	s.program.Send(models.RemoteClientsChangedMsg{Count: s.clients})
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()
	s.changeClients(1)
	defer s.changeClients(-1)

	scanner := bufio.NewScanner(conn)
	scanner.Buffer(make([]byte, 64*1024), maxRequestBytes)
	enc := json.NewEncoder(conn)

	for scanner.Scan() {
		var req Request
		if err := json.Unmarshal(scanner.Bytes(), &req); err != nil {
			if enc.Encode(Response{Error: "malformed request: " + err.Error()}) != nil {
				return
			}
			continue
		}
		if req.Op == "watch" {
			s.watch(conn, enc)
			return
		}
		if enc.Encode(s.do(req)) != nil {
			return
		}
	}
}

func (s *Server) do(req Request) Response {
	switch req.Op {
	case "key":
		msgs, err := parseKeys(req.Keys)
		if err != nil {
			return Response{Error: err.Error()}
		}
		s.sendKeys(msgs)
		return Response{OK: true}
	case "type":
		s.sendKeys(textKeys(req.Text))
		return Response{OK: true}
	case "screen":
		res, ok := s.snapshot()
		if !ok {
			return Response{Error: "radiogogo is shutting down"}
		}
		return Response{OK: true, Screen: res.screen}
	case "state":
		res, ok := s.snapshot()
		if !ok {
			return Response{Error: "radiogogo is shutting down"}
		}
		return Response{OK: true, State: newState(res.state)}
	case "wait":
		return s.wait(req)
	case "":
		return Response{Error: "missing op"}
	}
	return Response{Error: fmt.Sprintf("unknown op %q", req.Op)}
}

func (s *Server) sendKeys(msgs []tea.KeyMsg) {
	for _, msg := range msgs {
		s.program.Send(msg)
	}
}

func (s *Server) snapshot() (snapshotResult, bool) {
	reply := make(chan snapshotResult, 1)
	s.program.Send(snapshotReq{reply: reply})
	select {
	case res := <-reply:
		return res, true
	case <-s.done:
		return snapshotResult{}, false
	}
}

func (s *Server) wait(req Request) Response {
	if req.Text == "" {
		return Response{Error: "wait needs text"}
	}
	timeout := defaultWaitTimeout
	if req.TimeoutMS > 0 {
		timeout = time.Duration(req.TimeoutMS) * time.Millisecond
	}

	id := s.nextID.Add(1)
	reply := make(chan waitResult, 1)
	s.program.Send(waitReq{id: id, text: req.Text, reply: reply})

	var res waitResult
	select {
	case res = <-reply:
	case <-time.After(timeout):
		// The wrapper answers a cancel exactly once unless the wait already
		// matched, in which case that result is in the buffered channel.
		s.program.Send(waitCancel{id: id})
		select {
		case res = <-reply:
		case <-s.done:
			return Response{Error: "radiogogo is shutting down"}
		}
	case <-s.done:
		return Response{Error: "radiogogo is shutting down"}
	}

	if !res.matched {
		return Response{Error: fmt.Sprintf("timed out after %v waiting for %q", timeout, req.Text), Screen: res.screen}
	}
	return Response{OK: true, Screen: res.screen}
}

// watch streams events until the client disconnects or the server closes.
func (s *Server) watch(conn net.Conn, enc *json.Encoder) {
	id := s.nextID.Add(1)
	events := make(chan Event, watchBuffer)
	s.program.Send(watchReq{id: id, ch: events})
	defer s.program.Send(watchDrop{id: id})

	// Nothing more is read from a watching client, so EOF on the read side is
	// the only prompt signal that it went away.
	gone := make(chan struct{})
	go func() {
		_, _ = io.Copy(io.Discard, conn)
		close(gone)
	}()

	for {
		select {
		case e := <-events:
			if enc.Encode(e) != nil {
				return
			}
		case <-gone:
			return
		case <-s.done:
			return
		}
	}
}
