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
	"github.com/zi0p4tch0/radiogogo/common"
)

// PlaybackManagerService is an interface that defines methods for managing playback of a radio station.
type PlaybackManagerService interface {
	// Name returns the name of the playback manager.
	Name() string
	// IsAvailable returns true if the playback manager is available for use.
	IsAvailable() bool
	// NotAvailableErrorString returns a string that describes why the playback manager is not available.
	NotAvailableErrorString() string
	// IsPlaying returns true if a radio station is currently being played.
	IsPlaying() bool
	// PlayStation starts playing the specified radio station at the given volume.
	// If a radio station is already being played, it is stopped first.
	PlayStation(station common.Station, volume int) error
	// StopStation stops the currently playing radio station.
	// If no radio station is being played, this method does nothing.
	StopStation() error
	// VolumeMin returns the minimum volume level.
	VolumeMin() int
	// VolumeDefault returns the default volume level.
	VolumeDefault() int
	// VolumeMax returns the maximum volume level.
	VolumeMax() int
	// VolumeIsPercentage returns true if the volume is represented as a percentage.
	VolumeIsPercentage() bool
	// CurrentStation returns the station currently playing, or an empty Station if nothing is playing.
	CurrentStation() common.Station
}
