package mocks

import "radiogogo/common"

type MockRadioBrowserService struct {
	GetStationsFunc func(
		stationQuery common.StationQuery,
		searchTerm string,
		order string,
		reverse bool,
		offset uint64,
		limit uint64,
		hideBroken bool,
	) ([]common.Station, error)

	ClickStationFunc func(station common.Station) (common.ClickStationResponse, error)
}

func (m *MockRadioBrowserService) GetStations(
	stationQuery common.StationQuery,
	searchTerm string,
	order string,
	reverse bool,
	offset uint64,
	limit uint64,
	hideBroken bool,
) ([]common.Station, error) {
	return m.GetStationsFunc(stationQuery, searchTerm, order, reverse, offset, limit, hideBroken)
}

func (m *MockRadioBrowserService) ClickStation(station common.Station) (common.ClickStationResponse, error) {
	return m.ClickStationFunc(station)
}
