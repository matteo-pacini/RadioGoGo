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

package playback

import (
	"errors"
	"net/url"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/zi0p4tch0/radiogogo/common"
)

// mockProcess implements the Process interface for testing.
type mockProcess struct {
	pid        int
	killErr    error
	signalErr  error
	waitErr    error
	killCalled bool
	signalSig  os.Signal
}

func (p *mockProcess) Kill() error {
	p.killCalled = true
	return p.killErr
}

func (p *mockProcess) Signal(sig os.Signal) error {
	p.signalSig = sig
	return p.signalErr
}

func (p *mockProcess) Wait() (*os.ProcessState, error) {
	return nil, p.waitErr
}

func (p *mockProcess) Pid() int {
	return p.pid
}

// mockCmd implements the Cmd interface for testing.
type mockCmd struct {
	startErr error
	runErr   error
	process  *mockProcess
}

func (c *mockCmd) Start() error {
	return c.startErr
}

func (c *mockCmd) Run() error {
	return c.runErr
}

func (c *mockCmd) Process() Process {
	return c.process
}

func (c *mockCmd) SetStderr(w *os.File) {}
func (c *mockCmd) SetStdout(w *os.File) {}

// mockExecutor implements CommandExecutor for testing.
type mockExecutor struct {
	lookPathResults map[string]error
	commandFunc     func(name string, args ...string) Cmd
	commandCalls    [][]string
}

func newMockExecutor() *mockExecutor {
	return &mockExecutor{
		lookPathResults: make(map[string]error),
		commandCalls:    make([][]string, 0),
	}
}

func (e *mockExecutor) LookPath(file string) (string, error) {
	if err, ok := e.lookPathResults[file]; ok {
		return "", err
	}
	return "/usr/bin/" + file, nil
}

func (e *mockExecutor) Command(name string, args ...string) Cmd {
	e.commandCalls = append(e.commandCalls, append([]string{name}, args...))
	if e.commandFunc != nil {
		return e.commandFunc(name, args...)
	}
	return &mockCmd{process: &mockProcess{pid: 12345}}
}

// testStation creates a test station with the given URL.
func testStation(streamURL string) common.Station {
	u, _ := url.Parse(streamURL)
	return common.Station{
		StationUuid: uuid.New(),
		Name:        "Test Station",
		Url:         common.RadioGoGoURL{URL: *u},
	}
}

func TestFFPlayPlaybackManager(t *testing.T) {
	t.Run("NewFFPlaybackManager returns a valid manager", func(t *testing.T) {
		manager := NewFFPlaybackManager()
		assert.NotNil(t, manager)
	})

	t.Run("Name returns ffplay", func(t *testing.T) {
		manager := NewFFPlaybackManagerWithExecutor(newMockExecutor())
		assert.Equal(t, "ffplay", manager.Name())
	})

	t.Run("VolumeMin returns 0", func(t *testing.T) {
		manager := NewFFPlaybackManagerWithExecutor(newMockExecutor())
		assert.Equal(t, 0, manager.VolumeMin())
	})

	t.Run("VolumeDefault returns 80", func(t *testing.T) {
		manager := NewFFPlaybackManagerWithExecutor(newMockExecutor())
		assert.Equal(t, 80, manager.VolumeDefault())
	})

	t.Run("VolumeMax returns 100", func(t *testing.T) {
		manager := NewFFPlaybackManagerWithExecutor(newMockExecutor())
		assert.Equal(t, 100, manager.VolumeMax())
	})

	t.Run("VolumeIsPercentage returns false", func(t *testing.T) {
		manager := NewFFPlaybackManagerWithExecutor(newMockExecutor())
		assert.False(t, manager.VolumeIsPercentage())
	})

	t.Run("NotAvailableErrorString returns descriptive message", func(t *testing.T) {
		manager := NewFFPlaybackManagerWithExecutor(newMockExecutor())
		errStr := manager.NotAvailableErrorString()
		assert.Contains(t, errStr, "ffplay")
		assert.Contains(t, errStr, "ffmpeg")
		assert.Contains(t, errStr, "PATH")
	})

	t.Run("IsPlaying returns false when nothing is playing", func(t *testing.T) {
		manager := NewFFPlaybackManagerWithExecutor(newMockExecutor())
		assert.False(t, manager.IsPlaying())
	})

	t.Run("StopStation returns nil when nothing is playing", func(t *testing.T) {
		manager := NewFFPlaybackManagerWithExecutor(newMockExecutor())
		err := manager.StopStation()
		assert.NoError(t, err)
	})

	t.Run("RecordingNotAvailableErrorString returns descriptive message", func(t *testing.T) {
		manager := NewFFPlaybackManagerWithExecutor(newMockExecutor())
		errStr := manager.RecordingNotAvailableErrorString()
		assert.Contains(t, errStr, "ffmpeg")
		assert.Contains(t, errStr, "PATH")
	})

	t.Run("IsRecording returns false when nothing is recording", func(t *testing.T) {
		manager := NewFFPlaybackManagerWithExecutor(newMockExecutor())
		assert.False(t, manager.IsRecording())
	})

	t.Run("CurrentRecordingPath returns empty when not recording", func(t *testing.T) {
		manager := NewFFPlaybackManagerWithExecutor(newMockExecutor())
		assert.Empty(t, manager.CurrentRecordingPath())
	})

	t.Run("StopRecording returns empty path when not recording", func(t *testing.T) {
		manager := NewFFPlaybackManagerWithExecutor(newMockExecutor())
		path, err := manager.StopRecording()
		assert.NoError(t, err)
		assert.Empty(t, path)
	})

	t.Run("StartRecording returns error when not playing", func(t *testing.T) {
		manager := NewFFPlaybackManagerWithExecutor(newMockExecutor())
		err := manager.StartRecording("test.mp3")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no station is playing")
	})
}

func TestFFPlayPlaybackManager_IsAvailable(t *testing.T) {
	t.Run("returns true when ffplay is in PATH", func(t *testing.T) {
		executor := newMockExecutor()
		manager := NewFFPlaybackManagerWithExecutor(executor)
		assert.True(t, manager.IsAvailable())
	})

	t.Run("returns false when ffplay is not in PATH", func(t *testing.T) {
		executor := newMockExecutor()
		executor.lookPathResults["ffplay"] = errors.New("not found")
		manager := NewFFPlaybackManagerWithExecutor(executor)
		assert.False(t, manager.IsAvailable())
	})
}

func TestFFPlayPlaybackManager_IsRecordingAvailable(t *testing.T) {
	t.Run("returns true when ffmpeg is in PATH", func(t *testing.T) {
		executor := newMockExecutor()
		manager := NewFFPlaybackManagerWithExecutor(executor)
		assert.True(t, manager.IsRecordingAvailable())
	})

	t.Run("returns false when ffmpeg is not in PATH", func(t *testing.T) {
		executor := newMockExecutor()
		executor.lookPathResults["ffmpeg"] = errors.New("not found")
		manager := NewFFPlaybackManagerWithExecutor(executor)
		assert.False(t, manager.IsRecordingAvailable())
	})
}

func TestFFPlayPlaybackManager_PlayStation(t *testing.T) {
	t.Run("starts playback successfully", func(t *testing.T) {
		executor := newMockExecutor()
		executor.commandFunc = func(name string, args ...string) Cmd {
			return &mockCmd{process: &mockProcess{pid: 12345}}
		}
		manager := NewFFPlaybackManagerWithExecutor(executor)
		station := testStation("http://example.com/stream")

		err := manager.PlayStation(station, 80)

		assert.NoError(t, err)
		assert.True(t, manager.IsPlaying())
		assert.Equal(t, station.StationUuid, manager.CurrentStation().StationUuid)
	})

	t.Run("passes correct arguments to ffplay", func(t *testing.T) {
		executor := newMockExecutor()
		manager := NewFFPlaybackManagerWithExecutor(executor)
		station := testStation("http://example.com/stream")

		err := manager.PlayStation(station, 75)

		assert.NoError(t, err)
		assert.Len(t, executor.commandCalls, 1)
		call := executor.commandCalls[0]
		assert.Equal(t, "ffplay", call[0])
		assert.Contains(t, call, "-nodisp")
		assert.Contains(t, call, "-volume")
		assert.Contains(t, call, "75")
		assert.Contains(t, call, "http://example.com/stream")
	})

	t.Run("returns error when start fails", func(t *testing.T) {
		executor := newMockExecutor()
		executor.commandFunc = func(name string, args ...string) Cmd {
			return &mockCmd{startErr: errors.New("start failed")}
		}
		manager := NewFFPlaybackManagerWithExecutor(executor)
		station := testStation("http://example.com/stream")

		err := manager.PlayStation(station, 80)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "start failed")
		assert.False(t, manager.IsPlaying())
	})

	t.Run("stops previous station before starting new one", func(t *testing.T) {
		executor := newMockExecutor()
		process := &mockProcess{pid: 12345}
		executor.commandFunc = func(name string, args ...string) Cmd {
			return &mockCmd{process: process}
		}
		manager := NewFFPlaybackManagerWithExecutor(executor)

		station1 := testStation("http://example.com/stream1")
		station2 := testStation("http://example.com/stream2")

		err := manager.PlayStation(station1, 80)
		assert.NoError(t, err)
		assert.True(t, manager.IsPlaying())

		err = manager.PlayStation(station2, 80)
		assert.NoError(t, err)
		assert.True(t, manager.IsPlaying())
		assert.Equal(t, station2.StationUuid, manager.CurrentStation().StationUuid)
		assert.True(t, process.killCalled)
	})

	t.Run("handles volume at boundaries", func(t *testing.T) {
		testCases := []struct {
			name   string
			volume int
		}{
			{"minimum volume", 0},
			{"maximum volume", 100},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				executor := newMockExecutor()
				manager := NewFFPlaybackManagerWithExecutor(executor)
				station := testStation("http://example.com/stream")

				err := manager.PlayStation(station, tc.volume)

				assert.NoError(t, err)
			})
		}
	})
}

func TestFFPlayPlaybackManager_StopStation(t *testing.T) {
	t.Run("stops playing station", func(t *testing.T) {
		executor := newMockExecutor()
		process := &mockProcess{pid: 12345}
		executor.commandFunc = func(name string, args ...string) Cmd {
			return &mockCmd{process: process}
		}
		manager := NewFFPlaybackManagerWithExecutor(executor)
		station := testStation("http://example.com/stream")

		_ = manager.PlayStation(station, 80)
		err := manager.StopStation()

		assert.NoError(t, err)
		assert.False(t, manager.IsPlaying())
		assert.True(t, process.killCalled)
	})

	t.Run("clears current station", func(t *testing.T) {
		executor := newMockExecutor()
		manager := NewFFPlaybackManagerWithExecutor(executor)
		station := testStation("http://example.com/stream")

		_ = manager.PlayStation(station, 80)
		_ = manager.StopStation()

		assert.Equal(t, common.Station{}, manager.CurrentStation())
	})

	t.Run("returns error when kill fails", func(t *testing.T) {
		executor := newMockExecutor()
		process := &mockProcess{pid: 12345, killErr: errors.New("kill failed")}
		executor.commandFunc = func(name string, args ...string) Cmd {
			return &mockCmd{process: process}
		}
		manager := NewFFPlaybackManagerWithExecutor(executor)
		station := testStation("http://example.com/stream")

		_ = manager.PlayStation(station, 80)
		err := manager.StopStation()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "kill failed")
	})

	t.Run("stops recording when stopping station", func(t *testing.T) {
		executor := newMockExecutor()
		process := &mockProcess{pid: 12345}
		executor.commandFunc = func(name string, args ...string) Cmd {
			return &mockCmd{process: process}
		}
		manager := NewFFPlaybackManagerWithExecutor(executor)
		station := testStation("http://example.com/stream")

		_ = manager.PlayStation(station, 80)
		_ = manager.StartRecording("/tmp/test.mp3")
		err := manager.StopStation()

		assert.NoError(t, err)
		assert.False(t, manager.IsPlaying())
		assert.False(t, manager.IsRecording())
	})
}

func TestFFPlayPlaybackManager_Recording(t *testing.T) {
	t.Run("starts recording when playing", func(t *testing.T) {
		executor := newMockExecutor()
		executor.commandFunc = func(name string, args ...string) Cmd {
			return &mockCmd{process: &mockProcess{pid: 12345}}
		}
		manager := NewFFPlaybackManagerWithExecutor(executor)
		station := testStation("http://example.com/stream")

		_ = manager.PlayStation(station, 80)
		err := manager.StartRecording("/tmp/test.mp3")

		assert.NoError(t, err)
		assert.True(t, manager.IsRecording())
		assert.Equal(t, "/tmp/test.mp3", manager.CurrentRecordingPath())
	})

	t.Run("passes correct arguments to ffmpeg", func(t *testing.T) {
		executor := newMockExecutor()
		manager := NewFFPlaybackManagerWithExecutor(executor)
		station := testStation("http://example.com/stream")

		_ = manager.PlayStation(station, 80)
		_ = manager.StartRecording("/tmp/test.mp3")

		// Find the ffmpeg call
		var ffmpegCall []string
		for _, call := range executor.commandCalls {
			if call[0] == "ffmpeg" {
				ffmpegCall = call
				break
			}
		}

		assert.NotNil(t, ffmpegCall)
		assert.Contains(t, ffmpegCall, "-y")
		assert.Contains(t, ffmpegCall, "-i")
		assert.Contains(t, ffmpegCall, "http://example.com/stream")
		assert.Contains(t, ffmpegCall, "-c")
		assert.Contains(t, ffmpegCall, "copy")
		assert.Contains(t, ffmpegCall, "/tmp/test.mp3")
	})

	t.Run("returns error when not playing", func(t *testing.T) {
		executor := newMockExecutor()
		manager := NewFFPlaybackManagerWithExecutor(executor)

		err := manager.StartRecording("/tmp/test.mp3")

		assert.Error(t, err)
		assert.False(t, manager.IsRecording())
	})

	t.Run("returns error when start fails", func(t *testing.T) {
		executor := newMockExecutor()
		callCount := 0
		executor.commandFunc = func(name string, args ...string) Cmd {
			callCount++
			if name == "ffmpeg" {
				return &mockCmd{startErr: errors.New("ffmpeg failed")}
			}
			return &mockCmd{process: &mockProcess{pid: 12345}}
		}
		manager := NewFFPlaybackManagerWithExecutor(executor)
		station := testStation("http://example.com/stream")

		_ = manager.PlayStation(station, 80)
		err := manager.StartRecording("/tmp/test.mp3")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ffmpeg failed")
		assert.False(t, manager.IsRecording())
	})

	t.Run("stops previous recording before starting new one", func(t *testing.T) {
		executor := newMockExecutor()
		process := &mockProcess{pid: 12345}
		executor.commandFunc = func(name string, args ...string) Cmd {
			return &mockCmd{process: process}
		}
		manager := NewFFPlaybackManagerWithExecutor(executor)
		station := testStation("http://example.com/stream")

		_ = manager.PlayStation(station, 80)
		_ = manager.StartRecording("/tmp/test1.mp3")
		_ = manager.StartRecording("/tmp/test2.mp3")

		assert.Equal(t, "/tmp/test2.mp3", manager.CurrentRecordingPath())
	})
}

func TestFFPlayPlaybackManager_StopRecording(t *testing.T) {
	t.Run("stops recording and returns path", func(t *testing.T) {
		executor := newMockExecutor()
		process := &mockProcess{pid: 12345}
		executor.commandFunc = func(name string, args ...string) Cmd {
			return &mockCmd{process: process}
		}
		manager := NewFFPlaybackManagerWithExecutor(executor)
		station := testStation("http://example.com/stream")

		_ = manager.PlayStation(station, 80)
		_ = manager.StartRecording("/tmp/test.mp3")
		path, err := manager.StopRecording()

		assert.NoError(t, err)
		assert.Equal(t, "/tmp/test.mp3", path)
		assert.False(t, manager.IsRecording())
		assert.Empty(t, manager.CurrentRecordingPath())
	})

	t.Run("returns empty path when not recording", func(t *testing.T) {
		executor := newMockExecutor()
		manager := NewFFPlaybackManagerWithExecutor(executor)

		path, err := manager.StopRecording()

		assert.NoError(t, err)
		assert.Empty(t, path)
	})

	t.Run("sends interrupt signal on Unix", func(t *testing.T) {
		executor := newMockExecutor()
		process := &mockProcess{pid: 12345}
		executor.commandFunc = func(name string, args ...string) Cmd {
			return &mockCmd{process: process}
		}
		manager := NewFFPlaybackManagerWithExecutor(executor)
		station := testStation("http://example.com/stream")

		_ = manager.PlayStation(station, 80)
		_ = manager.StartRecording("/tmp/test.mp3")
		_, _ = manager.StopRecording()

		// On Unix, Signal should be called with Interrupt
		assert.Equal(t, os.Interrupt, process.signalSig)
	})

	t.Run("falls back to kill when signal fails", func(t *testing.T) {
		executor := newMockExecutor()
		process := &mockProcess{pid: 12345, signalErr: errors.New("signal failed")}
		executor.commandFunc = func(name string, args ...string) Cmd {
			return &mockCmd{process: process}
		}
		manager := NewFFPlaybackManagerWithExecutor(executor)
		station := testStation("http://example.com/stream")

		_ = manager.PlayStation(station, 80)
		_ = manager.StartRecording("/tmp/test.mp3")
		path, err := manager.StopRecording()

		assert.NoError(t, err)
		assert.Equal(t, "/tmp/test.mp3", path)
		assert.True(t, process.killCalled)
	})
}

func TestFFPlayPlaybackManager_CurrentStation(t *testing.T) {
	t.Run("returns empty station when not playing", func(t *testing.T) {
		executor := newMockExecutor()
		manager := NewFFPlaybackManagerWithExecutor(executor)

		assert.Equal(t, common.Station{}, manager.CurrentStation())
	})

	t.Run("returns current station when playing", func(t *testing.T) {
		executor := newMockExecutor()
		manager := NewFFPlaybackManagerWithExecutor(executor)
		station := testStation("http://example.com/stream")

		_ = manager.PlayStation(station, 80)

		result := manager.CurrentStation()
		assert.Equal(t, station.StationUuid, result.StationUuid)
		assert.Equal(t, station.Name, result.Name)
	})
}

func TestFFPlayPlaybackManager_VolumeEdgeCases(t *testing.T) {
	t.Run("handles negative volume", func(t *testing.T) {
		executor := newMockExecutor()
		manager := NewFFPlaybackManagerWithExecutor(executor)
		station := testStation("http://example.com/stream")

		// ffplay accepts negative volume (treated as 0)
		err := manager.PlayStation(station, -10)

		assert.NoError(t, err)
		call := executor.commandCalls[0]
		assert.Contains(t, call, "-10")
	})

	t.Run("handles volume above 100", func(t *testing.T) {
		executor := newMockExecutor()
		manager := NewFFPlaybackManagerWithExecutor(executor)
		station := testStation("http://example.com/stream")

		// ffplay may accept volumes above 100 (can clip/distort)
		err := manager.PlayStation(station, 200)

		assert.NoError(t, err)
		call := executor.commandCalls[0]
		assert.Contains(t, call, "200")
	})
}

func TestFFPlayPlaybackManager_URLEdgeCases(t *testing.T) {
	t.Run("handles URL with authentication", func(t *testing.T) {
		executor := newMockExecutor()
		manager := NewFFPlaybackManagerWithExecutor(executor)
		station := testStation("http://user:pass@example.com/stream")

		err := manager.PlayStation(station, 80)

		assert.NoError(t, err)
		call := executor.commandCalls[0]
		assert.Contains(t, call, "http://user:pass@example.com/stream")
	})

	t.Run("handles URL with query parameters", func(t *testing.T) {
		executor := newMockExecutor()
		manager := NewFFPlaybackManagerWithExecutor(executor)
		station := testStation("http://example.com/stream?token=abc123&quality=high")

		err := manager.PlayStation(station, 80)

		assert.NoError(t, err)
		call := executor.commandCalls[0]
		// The URL is passed as the last argument
		urlArg := call[len(call)-1]
		assert.Contains(t, urlArg, "token=abc123")
		assert.Contains(t, urlArg, "quality=high")
	})

	t.Run("handles URL with port", func(t *testing.T) {
		executor := newMockExecutor()
		manager := NewFFPlaybackManagerWithExecutor(executor)
		station := testStation("http://example.com:8080/stream")

		err := manager.PlayStation(station, 80)

		assert.NoError(t, err)
		call := executor.commandCalls[0]
		assert.Contains(t, call, "http://example.com:8080/stream")
	})

	t.Run("handles HTTPS URL", func(t *testing.T) {
		executor := newMockExecutor()
		manager := NewFFPlaybackManagerWithExecutor(executor)
		station := testStation("https://secure.example.com/stream")

		err := manager.PlayStation(station, 80)

		assert.NoError(t, err)
		call := executor.commandCalls[0]
		assert.Contains(t, call, "https://secure.example.com/stream")
	})

	t.Run("handles URL with special characters in path", func(t *testing.T) {
		executor := newMockExecutor()
		manager := NewFFPlaybackManagerWithExecutor(executor)
		station := testStation("http://example.com/stream/live%20radio")

		err := manager.PlayStation(station, 80)

		assert.NoError(t, err)
		call := executor.commandCalls[0]
		// The URL is passed as the last argument
		urlArg := call[len(call)-1]
		assert.Contains(t, urlArg, "live%20radio")
	})
}

func TestFFPlayPlaybackManager_WaitErrors(t *testing.T) {
	t.Run("StopStation returns error when Wait fails", func(t *testing.T) {
		executor := newMockExecutor()
		process := &mockProcess{pid: 12345, waitErr: errors.New("wait failed")}
		executor.commandFunc = func(name string, args ...string) Cmd {
			return &mockCmd{process: process}
		}
		manager := NewFFPlaybackManagerWithExecutor(executor)
		station := testStation("http://example.com/stream")

		_ = manager.PlayStation(station, 80)
		err := manager.StopStation()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "wait failed")
	})

	t.Run("StopRecording handles kill fallback error", func(t *testing.T) {
		executor := newMockExecutor()
		process := &mockProcess{
			pid:       12345,
			signalErr: errors.New("signal failed"),
			killErr:   errors.New("kill also failed"),
		}
		executor.commandFunc = func(name string, args ...string) Cmd {
			return &mockCmd{process: process}
		}
		manager := NewFFPlaybackManagerWithExecutor(executor)
		station := testStation("http://example.com/stream")

		_ = manager.PlayStation(station, 80)
		_ = manager.StartRecording("/tmp/test.mp3")
		_, err := manager.StopRecording()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "kill also failed")
	})
}

func TestFFPlayPlaybackManager_ConcurrentOperations(t *testing.T) {
	t.Run("multiple PlayStations stops previous correctly", func(t *testing.T) {
		executor := newMockExecutor()
		killCount := 0
		executor.commandFunc = func(name string, args ...string) Cmd {
			return &mockCmd{process: &mockProcess{
				pid: 12345,
				killErr: nil,
			}}
		}
		manager := NewFFPlaybackManagerWithExecutor(executor)

		station1 := testStation("http://example.com/stream1")
		station2 := testStation("http://example.com/stream2")
		station3 := testStation("http://example.com/stream3")

		_ = manager.PlayStation(station1, 80)
		_ = manager.PlayStation(station2, 80)
		_ = manager.PlayStation(station3, 80)

		// Still playing after all switches
		assert.True(t, manager.IsPlaying())
		assert.Equal(t, station3.StationUuid, manager.CurrentStation().StationUuid)
		_ = killCount // silence unused warning
	})

	t.Run("double StopStation is safe", func(t *testing.T) {
		executor := newMockExecutor()
		manager := NewFFPlaybackManagerWithExecutor(executor)
		station := testStation("http://example.com/stream")

		_ = manager.PlayStation(station, 80)
		err1 := manager.StopStation()
		err2 := manager.StopStation()

		assert.NoError(t, err1)
		assert.NoError(t, err2)
		assert.False(t, manager.IsPlaying())
	})

	t.Run("double StopRecording is safe", func(t *testing.T) {
		executor := newMockExecutor()
		executor.commandFunc = func(name string, args ...string) Cmd {
			return &mockCmd{process: &mockProcess{pid: 12345}}
		}
		manager := NewFFPlaybackManagerWithExecutor(executor)
		station := testStation("http://example.com/stream")

		_ = manager.PlayStation(station, 80)
		_ = manager.StartRecording("/tmp/test.mp3")
		path1, err1 := manager.StopRecording()
		path2, err2 := manager.StopRecording()

		assert.NoError(t, err1)
		assert.Equal(t, "/tmp/test.mp3", path1)
		assert.NoError(t, err2)
		assert.Empty(t, path2)
	})
}

func TestFFPlayPlaybackManager_RecordingPaths(t *testing.T) {
	t.Run("handles path with spaces", func(t *testing.T) {
		executor := newMockExecutor()
		executor.commandFunc = func(name string, args ...string) Cmd {
			return &mockCmd{process: &mockProcess{pid: 12345}}
		}
		manager := NewFFPlaybackManagerWithExecutor(executor)
		station := testStation("http://example.com/stream")

		_ = manager.PlayStation(station, 80)
		err := manager.StartRecording("/tmp/my recordings/test.mp3")

		assert.NoError(t, err)
		assert.Equal(t, "/tmp/my recordings/test.mp3", manager.CurrentRecordingPath())
	})

	t.Run("handles absolute Windows-style path", func(t *testing.T) {
		executor := newMockExecutor()
		executor.commandFunc = func(name string, args ...string) Cmd {
			return &mockCmd{process: &mockProcess{pid: 12345}}
		}
		manager := NewFFPlaybackManagerWithExecutor(executor)
		station := testStation("http://example.com/stream")

		_ = manager.PlayStation(station, 80)
		err := manager.StartRecording("C:\\Users\\Test\\Music\\recording.mp3")

		assert.NoError(t, err)
		assert.Equal(t, "C:\\Users\\Test\\Music\\recording.mp3", manager.CurrentRecordingPath())
	})

	t.Run("handles empty output path", func(t *testing.T) {
		executor := newMockExecutor()
		executor.commandFunc = func(name string, args ...string) Cmd {
			return &mockCmd{process: &mockProcess{pid: 12345}}
		}
		manager := NewFFPlaybackManagerWithExecutor(executor)
		station := testStation("http://example.com/stream")

		_ = manager.PlayStation(station, 80)
		err := manager.StartRecording("")

		// ffmpeg will fail with empty path, but StartRecording doesn't validate
		assert.NoError(t, err)
		assert.Equal(t, "", manager.CurrentRecordingPath())
	})

	t.Run("handles unicode path", func(t *testing.T) {
		executor := newMockExecutor()
		executor.commandFunc = func(name string, args ...string) Cmd {
			return &mockCmd{process: &mockProcess{pid: 12345}}
		}
		manager := NewFFPlaybackManagerWithExecutor(executor)
		station := testStation("http://example.com/stream")

		_ = manager.PlayStation(station, 80)
		err := manager.StartRecording("/tmp/録音/test.mp3")

		assert.NoError(t, err)
		assert.Equal(t, "/tmp/録音/test.mp3", manager.CurrentRecordingPath())
	})
}
