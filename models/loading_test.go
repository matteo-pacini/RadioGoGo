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

package models

import (
	"io"
	"testing"

	"github.com/zi0p4tch0/radiogogo/common"

	"github.com/zi0p4tch0/radiogogo/mocks"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

func TestLoadingModel_Init(t *testing.T) {

	t.Run("starts the spinner", func(t *testing.T) {

		mockBrowser := mocks.MockRadioBrowserService{}
		model := NewLoadingModel(Theme{}, &mockBrowser, common.StationQueryAll, "text")

		cmd := model.Init()
		assert.NotNil(t, cmd)

		var batchMsg tea.BatchMsg = cmd().(tea.BatchMsg)

		found := false
		for _, msg := range batchMsg {
			currentMsg := msg()
			if _, ok := currentMsg.(spinner.TickMsg); ok {
				found = true
				break
			}
		}

		assert.True(t, found)

	})

	t.Run("searches for stations and broadcasts switchToStationsModelMsg on success", func(t *testing.T) {

		mockBrowser := mocks.MockRadioBrowserService{
			GetStationsFunc: func(stationQuery common.StationQuery, searchTerm string, order string, reverse bool, offset uint64, limit uint64, hideBroken bool) ([]common.Station, error) {
				return []common.Station{}, nil
			},
			ClickStationFunc: func(station common.Station) (common.ClickStationResponse, error) {
				return common.ClickStationResponse{}, nil
			},
		}

		model := NewLoadingModel(Theme{}, &mockBrowser, common.StationQueryAll, "text")

		cmd := model.Init()
		assert.NotNil(t, cmd)

		var batchMsg tea.BatchMsg = cmd().(tea.BatchMsg)

		found := false
		for _, msg := range batchMsg {
			currentMsg := msg()
			if _, ok := currentMsg.(switchToStationsModelMsg); ok {
				found = true
				break
			}
		}

		assert.True(t, found)

	})

	t.Run("searches for stations and broadcasts switchToErrorModelMsg on error", func(t *testing.T) {

		mockBrowser := mocks.MockRadioBrowserService{
			GetStationsFunc: func(stationQuery common.StationQuery, searchTerm string, order string, reverse bool, offset uint64, limit uint64, hideBroken bool) ([]common.Station, error) {
				return nil, io.EOF
			},
			ClickStationFunc: func(station common.Station) (common.ClickStationResponse, error) {
				return common.ClickStationResponse{}, io.EOF
			},
		}

		model := NewLoadingModel(Theme{}, &mockBrowser, common.StationQueryAll, "text")

		cmd := model.Init()
		assert.NotNil(t, cmd)

		var batchMsg tea.BatchMsg = cmd().(tea.BatchMsg)

		found := false
		for _, msg := range batchMsg {
			currentMsg := msg()
			if _, ok := currentMsg.(switchToErrorModelMsg); ok {
				found = true
				break
			}
		}

		assert.True(t, found)

	})

}
