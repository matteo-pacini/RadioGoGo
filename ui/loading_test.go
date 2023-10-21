package ui

import (
	"bytes"
	"io"
	"net/http"
	"radiogogo/api"
	"radiogogo/mocks"
	"testing"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

func TestLoadingModel_Init(t *testing.T) {

	mockDNSLookupService := mocks.MockDNSLookupService{
		LookupIPFunc: func(host string) ([]string, error) {
			return []string{"127.0.0.1"}, nil
		},
	}

	mocksHttpClient := mocks.MockHttpClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(bytes.NewReader([]byte("[]"))),
			}, nil
		},
	}

	t.Run("starts the spinner", func(t *testing.T) {

		browser, _ := api.NewRadioBrowserWithDependencies(&mockDNSLookupService, &mocksHttpClient)
		assert.NotNil(t, browser)

		model := NewLoadingModel(browser, "text")

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

		browser, _ := api.NewRadioBrowserWithDependencies(&mockDNSLookupService, &mocksHttpClient)
		assert.NotNil(t, browser)

		model := NewLoadingModel(browser, "text")

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

		mockHttpClient := mocks.MockHttpClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				return nil, io.EOF
			},
		}

		browser, _ := api.NewRadioBrowserWithDependencies(&mockDNSLookupService, &mockHttpClient)
		assert.NotNil(t, browser)

		model := NewLoadingModel(browser, "text")

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
