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

package models

import (
	"fmt"
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
