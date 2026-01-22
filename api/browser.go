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

package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/zi0p4tch0/radiogogo/common"
	"github.com/zi0p4tch0/radiogogo/data"
)

type RadioBrowserService interface {
	// GetStations retrieves a list of radio stations from the RadioBrowser API based on the provided StationQuery, searchTerm, order, reverse, offset, limit and hideBroken parameters.
	// If stationQuery is not StationQueryAll, the searchTerm is used to filter the results.
	// The order parameter specifies the field to order the results by.
	// The reverse parameter specifies whether the results should be returned in reverse order.
	// The offset parameter specifies the number of results to skip before returning the remaining results.
	// The limit parameter specifies the maximum number of results to return.
	// The hideBroken parameter specifies whether to exclude broken stations from the results.
	// Returns a slice of Station structs and an error if any occurred.
	GetStations(
		stationQuery common.StationQuery,
		searchTerm string,
		order string,
		reverse bool,
		offset uint64,
		limit uint64,
		hideBroken bool,
	) ([]common.Station, error)
	// ClickStation sends a POST request to the RadioBrowser API to increment the click count of a given station.
	// It takes a Station struct as input and returns a ClickStationResponse struct and an error.
	ClickStation(station common.Station) (common.ClickStationResponse, error)
}

type RadioBrowserImpl struct {
	// The HTTP client used to make requests to the Radio Browser API.
	httpClient HTTPClientService
	// The base URL for the Radio Browser API.)
	baseUrl url.URL
}

// NewRadioBrowser returns a new instance of RadioBrowserService with the default HTTP client.
func NewRadioBrowser() (RadioBrowserService, error) {
	return NewRadioBrowserWithDependencies(http.DefaultClient)
}

// NewRadioBrowserWithDependencies creates a new instance of RadioBrowserService with the provided HTTP client.
// Returns an error if URL parsing fails.
func NewRadioBrowserWithDependencies(
	httpClient HTTPClientService,
) (RadioBrowserService, error) {
	browser := &RadioBrowserImpl{
		httpClient: httpClient,
	}

	url, err := url.Parse("https://all.api.radio-browser.info/json")
	if err != nil {
		return nil, err
	}
	browser.baseUrl = *url
	return browser, nil
}

func (radioBrowser *RadioBrowserImpl) GetStations(
	stationQuery common.StationQuery,
	searchTerm string,
	order string,
	reverse bool,
	offset uint64,
	limit uint64,
	hideBroken bool,
) ([]common.Station, error) {

	url := radioBrowser.baseUrl.JoinPath("/stations")
	if stationQuery != common.StationQueryAll {
		url = url.JoinPath("/" + string(stationQuery) + "/" + searchTerm)
	}

	query := url.Query()
	query.Set("order", order)
	query.Set("reverse", boolToString(reverse))
	query.Set("offset", uint64ToString(offset))
	query.Set("limit", uint64ToString(limit))
	query.Set("hidebroken", boolToString(hideBroken))
	url.RawQuery = query.Encode()

	headers := make(map[string]string)
	headers["User-Agent"] = data.UserAgent
	headers["Accept"] = "application/json"

	var stations []common.Station

	req, err := http.NewRequest("GET", url.String(), nil)
	if err != nil {
		return nil, err
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	result, err := radioBrowser.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer result.Body.Close()

	if result.StatusCode != 200 {
		return nil, fmt.Errorf("API request failed with status %d", result.StatusCode)
	}

	err = json.NewDecoder(result.Body).Decode(&stations)

	if err != nil {
		return nil, err
	}

	return stations, nil

}

func (radioBrowser *RadioBrowserImpl) ClickStation(station common.Station) (common.ClickStationResponse, error) {

	url := radioBrowser.baseUrl.JoinPath("/url/" + station.StationUuid.String())

	headers := make(map[string]string)
	headers["User-Agent"] = data.UserAgent
	headers["Accept"] = "application/json"

	req, err := http.NewRequest("POST", url.String(), nil)
	if err != nil {
		return common.ClickStationResponse{}, err
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	result, err := radioBrowser.httpClient.Do(req)
	if err != nil {
		return common.ClickStationResponse{}, err
	}

	defer result.Body.Close()

	if result.StatusCode != 200 {
		return common.ClickStationResponse{}, fmt.Errorf("API request failed with status %d", result.StatusCode)
	}

	var response common.ClickStationResponse
	err = json.NewDecoder(result.Body).Decode(&response)

	if err != nil {
		return common.ClickStationResponse{}, err
	}

	return response, nil
}
