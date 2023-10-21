package mocks

import "radiogogo/common"

type MockPlaybackManagerService struct {
	IsAvailableResult bool
	IsPlayingResult   bool
	PlayStationFunc   func(station common.Station, volume int) error
	StopStationFunc   func() error
}

func (m *MockPlaybackManagerService) IsAvailable() bool {
	return m.IsAvailableResult
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
