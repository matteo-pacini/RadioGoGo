// Copyright (c) 2023 Matteo Pacini
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
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"radiogogo/data"
	"radiogogo/models"

	"github.com/google/uuid"
)

type RadioBrowser struct {
	// The HTTP client used to make requests to the Radio Browser API.
	httpClient HTTPClient
	// The base URL for the Radio Browser API.)
	baseUrl url.URL
}

// DefaultRadioBrowser returns a new instance of RadioBrowser with the default DNS lookup service and HTTP client.
func NewDefaultRadioBrowser() (*RadioBrowser, error) {
	return NewRadioBrowser(
		&DefaultDNSLookupService{},
		http.DefaultClient,
	)
}

// NewRadioBrowser creates a new instance of RadioBrowser struct with the provided DNSLookupService and HTTPClient.
// It returns a pointer to the created instance and an error if any.
// The function performs a DNS lookup for "all.api.radio-browser.info" and selects a random IP address from the returned list.
// It then constructs a base URL using the selected IP address and sets it as the baseUrl of the created instance.
func NewRadioBrowser(
	dnsLookupService DNSLookupService,
	httpClient HTTPClient,
) (*RadioBrowser, error) {
	browser := &RadioBrowser{
		httpClient: httpClient,
	}
	ips, err := dnsLookupService.LookupIP("all.api.radio-browser.info")
	if err != nil {
		return nil, err
	}

	randomIp := ips[rand.Intn(len(ips))]

	if net.ParseIP(randomIp).To4() == nil {
		randomIp = "[" + randomIp + "]"
	}

	url, err := url.Parse("http://" + randomIp + "/json")
	if err != nil {
		return nil, err
	}
	browser.baseUrl = *url
	return browser, nil
}

type StationQuery string

const (
	StationQueryAll                StationQuery = ""
	StationQueryByUuid             StationQuery = "byuuid"
	StationQueryByName             StationQuery = "byname"
	StationQueryByNameExact        StationQuery = "bynameexact"
	StationQueryByCodec            StationQuery = "bycodec"
	StationQueryByCodecExact       StationQuery = "bycodecexact"
	StationQueryByCountry          StationQuery = "bycountry"
	StationQueryByCountryExact     StationQuery = "bycountryexact"
	StationQueryByCountryCodeExact StationQuery = "bycountrycodeexact"
	StationQueryByState            StationQuery = "bystate"
	StationQueryByStateExact       StationQuery = "bystateexact"
	StationQueryByLanguage         StationQuery = "bylanguage"
	StationQueryByLanguageExact    StationQuery = "bylanguageexact"
	StationQueryByTag              StationQuery = "bytag"
	StationQueryByTagExact         StationQuery = "bytagexact"
)

// GetStations retrieves a list of radio stations from the RadioBrowser API based on the provided StationQuery, searchTerm, order, reverse, offset, limit and hideBroken parameters.
// If stationQuery is not StationQueryAll, the searchTerm is used to filter the results.
// The order parameter specifies the field to order the results by.
// The reverse parameter specifies whether the results should be returned in reverse order.
// The offset parameter specifies the number of results to skip before returning the remaining results.
// The limit parameter specifies the maximum number of results to return.
// The hideBroken parameter specifies whether to exclude broken stations from the results.
// Returns a slice of Station structs and an error if any occurred.
func (radioBrowser *RadioBrowser) GetStations(
	stationQuery StationQuery,
	searchTerm string,
	order string,
	reverse bool,
	offset uint64,
	limit uint64,
	hideBroken bool,
) ([]models.Station, error) {

	url := radioBrowser.baseUrl.JoinPath("/stations")
	if stationQuery != StationQueryAll {
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

	var stations []models.Station

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

	err = json.NewDecoder(result.Body).Decode(&stations)

	if err != nil {
		return nil, err
	}

	return stations, nil

}

// ClickStationResponse represents the response returned by the API when a user clicks on a station.
type ClickStationResponse struct {
	// Ok indicates whether the request was successful or not.
	Ok bool `json:"ok"`

	// Message contains an optional message returned by the server.
	Message string `json:"message"`

	// StationUuid is the unique identifier of the station.
	StationUuid uuid.UUID `json:"stationuuid"`

	// Name is the name of the station.
	Name string `json:"name"`

	// Url is the URL of the station's stream.
	Url models.RadioGoGoURL `json:"url"`
}

// ClickStation sends a POST request to the RadioBrowser API to increment the click count of a given station.
// It takes a Station struct as input and returns a ClickStationResponse struct and an error.
func (radioBrowser *RadioBrowser) ClickStation(station models.Station) (ClickStationResponse, error) {

	// POST json/url/stationuuid

	url := radioBrowser.baseUrl.JoinPath("/url/" + station.StationUuid.String())

	headers := make(map[string]string)
	headers["User-Agent"] = data.UserAgent
	headers["Accept"] = "application/json"

	req, err := http.NewRequest("POST", url.String(), nil)
	if err != nil {
		return ClickStationResponse{}, err
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	result, err := radioBrowser.httpClient.Do(req)
	if err != nil {
		return ClickStationResponse{}, err
	}

	defer result.Body.Close()

	var response ClickStationResponse
	err = json.NewDecoder(result.Body).Decode(&response)

	if err != nil {
		return ClickStationResponse{}, err
	}

	return response, nil
}
