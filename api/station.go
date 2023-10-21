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
	"fmt"
	"net/http"
	"radiogogo/data"
	"time"

	"github.com/google/uuid"
)

// BoolFromlInt represents a boolean value that is converted from an integer value (0 or 1).
type BoolFromlInt bool

type Station struct {
	// A globally unique identifier for the change of the station information
	ChangeUuid uuid.UUID `json:"changeuuid"`
	// A globally unique identifier for the station
	StationUuid uuid.UUID `json:"stationuuid"`
	// The name of the station
	Name string `json:"name"`
	// The stream URL provided by the user
	Url RadioGoGoURL `json:"url"`
	// An automatically "resolved" stream URL. Things resolved are playlists (M3U/PLS/ASX...),
	// HTTP redirects (Code 301/302).
	// This link is especially usefull if you use this API from a platform that is not able to
	// do a resolve on its own (e.g. JavaScript in browser) or you just don't want to invest
	// the time in decoding playlists yourself.
	UrlResolved RadioGoGoURL `json:"url_resolved"`
	// URL to an icon or picture that represents the stream. (PNG, JPG)
	Favicon RadioGoGoURL `json:"favicon"`
	// Tags of the stream with more information about it (string, multivalue, split by comma).
	Tags string `json:"tags"`
	// Official countrycodes as in ISO 3166-1 alpha-2
	CountryCode string `json:"countrycode"`
	// Full name of the entity where the station is located inside the country
	State string `json:"state"`
	// Languages that are spoken in this stream.
	Languages string `json:"language"`
	// Languages that are spoken in this stream by code ISO 639-2/B.
	LanguagesCodes string `json:"languagecodes"`
	// Number of votes for this station. This number is by server and only ever increases.
	// It will never be reset to 0.
	Votes uint64 `json:"votes"`
	// Last time when the stream information was changed in the database
	LastChangeTime time.Time `json:"lastchangetime_iso8601"`
	// The codec of this stream recorded at the last check.
	Codec string `json:"codec"`
	// The bitrate of this stream recorded at the last check.
	Bitrate uint64 `json:"bitrate"`
	// Mark if this stream is using HLS distribution or non-HLS.
	Hls BoolFromlInt `json:"hls"`
	// The current online/offline state of this stream.
	// This is a value calculated from multiple measure points in the internet.
	// The test servers are located in different countries. It is a majority vote.
	LastCheckOk BoolFromlInt `json:"lastcheckok"`
	// The last time when any radio-browser server checked the online state of this stream
	LastCheckTime time.Time `json:"lastchecktime_iso8601"`
	// The last time when the stream was checked for the online status with a positive result
	LastCheckOkTime time.Time `json:"lastcheckoktime_iso8601"`
	// The last time when this server checked the online state and the metadata of this stream.
	LastLocalCheckTime time.Time `json:"lastlocalchecktime_iso8601"`
	// The time of the last click recorded for this stream
	ClickTimestamp *time.Time `json:"clicktimestamp_iso8601,omitempty"`
	// Clicks within the last 24 hours
	ClickCount uint64 `json:"clickcount"`
	// The difference of the clickcounts within the last 2 days.
	// Posivite values mean an increase, negative a decrease of clicks.
	ClickTrend int64 `json:"clicktrend"`
	// 0 means no error, 1 means that there was an ssl error while connecting to the stream url.
	SslError BoolFromlInt `json:"ssl_error"`
	// Latitude on earth where the stream is located.
	GeoLat *float64 `json:"geo_lat,omitempty"`
	// Longitude on earth where the stream is located.
	GeoLong *float64 `json:"geo_long,omitempty"`
	// Is true, if the stream owner does provide extended information as HTTP headers
	// which override the information in the database.
	HasExtendedInfo *bool `json:"has_extended_info,omitempty"`
}

func (bi *BoolFromlInt) UnmarshalJSON(data []byte) error {
	switch string(data) {
	case "1":
		*bi = true
	case "0", "null": // "null" to handle if the field is null in JSON
		*bi = false
	default:
		return fmt.Errorf("boolean from int unmarshal error: invalid input %s", data)
	}
	return nil
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
) ([]Station, error) {

	url := radioBrowser.BaseUrl.JoinPath("/stations")
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

	var stations []Station

	req, err := http.NewRequest("GET", url.String(), nil)
	if err != nil {
		return nil, err
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	result, err := Client.Do(req)
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
	Url RadioGoGoURL `json:"url"`
}

// ClickStation sends a POST request to the RadioBrowser API to increment the click count of a given station.
// It takes a Station struct as input and returns a ClickStationResponse struct and an error.
func (radioBrowser *RadioBrowser) ClickStation(station Station) (ClickStationResponse, error) {

	// POST json/url/stationuuid

	url := radioBrowser.BaseUrl.JoinPath("/url/" + station.StationUuid.String())

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

	result, err := Client.Do(req)
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
