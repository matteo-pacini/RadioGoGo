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
