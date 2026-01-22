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

package mocks

import "github.com/zi0p4tch0/radiogogo/common"

type MockPlaybackManagerService struct {
	NameResult                    string
	IsAvailableResult             bool
	NotAvailableErrorStringResult string
	IsPlayingResult               bool
	PlayStationFunc               func(station common.Station, volume int) error
	StopStationFunc               func() error
	VolumeMinResult               int
	VolumeDefaultResult           int
	VolumeMaxResult               int
	VolumeIsPercentageResult      bool
	CurrentStationResult          common.Station
}

func (m *MockPlaybackManagerService) IsAvailable() bool {
	return m.IsAvailableResult
}

func (m *MockPlaybackManagerService) Name() string {
	return m.NameResult
}

func (m *MockPlaybackManagerService) NotAvailableErrorString() string {
	return m.NotAvailableErrorStringResult
}

func (m *MockPlaybackManagerService) IsPlaying() bool {
	return m.IsPlayingResult
}

func (m *MockPlaybackManagerService) PlayStation(station common.Station, volume int) error {
	return m.PlayStationFunc(station, volume)
}

func (m *MockPlaybackManagerService) StopStation() error {
	return m.StopStationFunc()
}

func (m *MockPlaybackManagerService) VolumeMin() int {
	return m.VolumeMinResult
}

func (m *MockPlaybackManagerService) VolumeDefault() int {
	return m.VolumeDefaultResult
}

func (m *MockPlaybackManagerService) VolumeMax() int {
	return m.VolumeMaxResult
}

func (m *MockPlaybackManagerService) VolumeIsPercentage() bool {
	return m.VolumeIsPercentageResult
}

func (m *MockPlaybackManagerService) CurrentStation() common.Station {
	return m.CurrentStationResult
}
