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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFFPlayPlaybackManager(t *testing.T) {

	t.Run("NewFFPlaybackManager returns a valid manager", func(t *testing.T) {
		manager := NewFFPlaybackManager()
		assert.NotNil(t, manager)
	})

	t.Run("Name returns ffplay", func(t *testing.T) {
		manager := &FFPlayPlaybackManager{}
		assert.Equal(t, "ffplay", manager.Name())
	})

	t.Run("VolumeMin returns 0", func(t *testing.T) {
		manager := &FFPlayPlaybackManager{}
		assert.Equal(t, 0, manager.VolumeMin())
	})

	t.Run("VolumeDefault returns 80", func(t *testing.T) {
		manager := &FFPlayPlaybackManager{}
		assert.Equal(t, 80, manager.VolumeDefault())
	})

	t.Run("VolumeMax returns 100", func(t *testing.T) {
		manager := &FFPlayPlaybackManager{}
		assert.Equal(t, 100, manager.VolumeMax())
	})

	t.Run("VolumeIsPercentage returns false", func(t *testing.T) {
		manager := &FFPlayPlaybackManager{}
		assert.False(t, manager.VolumeIsPercentage())
	})

	t.Run("NotAvailableErrorString returns descriptive message", func(t *testing.T) {
		manager := &FFPlayPlaybackManager{}
		errStr := manager.NotAvailableErrorString()
		assert.Contains(t, errStr, "ffplay")
		assert.Contains(t, errStr, "ffmpeg")
		assert.Contains(t, errStr, "PATH")
	})

	t.Run("IsPlaying returns false when nothing is playing", func(t *testing.T) {
		manager := &FFPlayPlaybackManager{}
		assert.False(t, manager.IsPlaying())
	})

	t.Run("StopStation returns nil when nothing is playing", func(t *testing.T) {
		manager := &FFPlayPlaybackManager{}
		err := manager.StopStation()
		assert.NoError(t, err)
	})

	t.Run("RecordingNotAvailableErrorString returns descriptive message", func(t *testing.T) {
		manager := &FFPlayPlaybackManager{}
		errStr := manager.RecordingNotAvailableErrorString()
		assert.Contains(t, errStr, "ffmpeg")
		assert.Contains(t, errStr, "PATH")
	})

	t.Run("IsRecording returns false when nothing is recording", func(t *testing.T) {
		manager := &FFPlayPlaybackManager{}
		assert.False(t, manager.IsRecording())
	})

	t.Run("CurrentRecordingPath returns empty when not recording", func(t *testing.T) {
		manager := &FFPlayPlaybackManager{}
		assert.Empty(t, manager.CurrentRecordingPath())
	})

	t.Run("StopRecording returns empty path when not recording", func(t *testing.T) {
		manager := &FFPlayPlaybackManager{}
		path, err := manager.StopRecording()
		assert.NoError(t, err)
		assert.Empty(t, path)
	})

	t.Run("StartRecording returns error when not playing", func(t *testing.T) {
		manager := &FFPlayPlaybackManager{}
		err := manager.StartRecording("test.mp3")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no station is playing")
	})
}
