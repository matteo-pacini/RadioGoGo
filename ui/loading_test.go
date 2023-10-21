package ui

import (
	"io"
	"radiogogo/common"
	"radiogogo/mocks"
	"testing"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

func TestLoadingModel_Init(t *testing.T) {

	t.Run("starts the spinner", func(t *testing.T) {

		mockBrowser := mocks.MockRadioBrowserService{}
		model := NewLoadingModel(&mockBrowser, "text")

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

		model := NewLoadingModel(&mockBrowser, "text")

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

		model := NewLoadingModel(&mockBrowser, "text")

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
