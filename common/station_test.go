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

package common

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBoolFromlInt_UnmarshalJSON(t *testing.T) {

	t.Run("unmarshals 1 as true", func(t *testing.T) {
		var b BoolFromlInt
		err := b.UnmarshalJSON([]byte("1"))
		assert.NoError(t, err)
		assert.True(t, bool(b))
	})

	t.Run("unmarshals true as true", func(t *testing.T) {
		var b BoolFromlInt
		err := b.UnmarshalJSON([]byte("true"))
		assert.NoError(t, err)
		assert.True(t, bool(b))
	})

	t.Run("unmarshals 0 as false", func(t *testing.T) {
		var b BoolFromlInt
		err := b.UnmarshalJSON([]byte("0"))
		assert.NoError(t, err)
		assert.False(t, bool(b))
	})

	t.Run("unmarshals false as false", func(t *testing.T) {
		var b BoolFromlInt
		err := b.UnmarshalJSON([]byte("false"))
		assert.NoError(t, err)
		assert.False(t, bool(b))
	})

	t.Run("unmarshals null as false", func(t *testing.T) {
		var b BoolFromlInt
		err := b.UnmarshalJSON([]byte("null"))
		assert.NoError(t, err)
		assert.False(t, bool(b))
	})

	t.Run("returns error for invalid input", func(t *testing.T) {
		var b BoolFromlInt
		err := b.UnmarshalJSON([]byte("invalid"))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "boolean from int unmarshal error")
	})

	t.Run("returns error for number other than 0 or 1", func(t *testing.T) {
		var b BoolFromlInt
		err := b.UnmarshalJSON([]byte("2"))
		assert.Error(t, err)
	})

	t.Run("returns error for negative numbers", func(t *testing.T) {
		var b BoolFromlInt
		err := b.UnmarshalJSON([]byte("-1"))
		assert.Error(t, err)
	})

	t.Run("returns error for floating point numbers", func(t *testing.T) {
		var b BoolFromlInt
		err := b.UnmarshalJSON([]byte("0.5"))
		assert.Error(t, err)
	})

	t.Run("returns error for empty input", func(t *testing.T) {
		var b BoolFromlInt
		err := b.UnmarshalJSON([]byte(""))
		assert.Error(t, err)
	})

	t.Run("returns error for whitespace only", func(t *testing.T) {
		var b BoolFromlInt
		err := b.UnmarshalJSON([]byte("   "))
		assert.Error(t, err)
	})

	t.Run("returns error for string TRUE", func(t *testing.T) {
		var b BoolFromlInt
		err := b.UnmarshalJSON([]byte("TRUE"))
		assert.Error(t, err) // Only lowercase "true" is accepted
	})

	t.Run("returns error for string FALSE", func(t *testing.T) {
		var b BoolFromlInt
		err := b.UnmarshalJSON([]byte("FALSE"))
		assert.Error(t, err) // Only lowercase "false" is accepted
	})

	t.Run("returns error for large numbers", func(t *testing.T) {
		var b BoolFromlInt
		err := b.UnmarshalJSON([]byte("999999999"))
		assert.Error(t, err)
	})
}

func TestStation_JSONUnmarshal(t *testing.T) {
	t.Run("parses full station JSON", func(t *testing.T) {
		jsonData := `{
			"changeuuid": "941ef6f1-0699-4821-95b1-2b678e3ff62e",
			"stationuuid": "16a73a57-5dba-11e8-b0ce-52543be04c81",
			"name": "Test Radio",
			"url": "http://example.com/stream",
			"url_resolved": "http://example.com/stream/resolved",
			"favicon": "http://example.com/favicon.png",
			"tags": "rock,pop,music",
			"countrycode": "US",
			"state": "California",
			"language": "English",
			"languagecodes": "en",
			"votes": 100,
			"lastchangetime_iso8601": "2024-01-15T10:30:00Z",
			"codec": "MP3",
			"bitrate": 128,
			"hls": 0,
			"lastcheckok": 1,
			"lastchecktime_iso8601": "2024-01-15T12:00:00Z",
			"lastcheckoktime_iso8601": "2024-01-15T12:00:00Z",
			"lastlocalchecktime_iso8601": "2024-01-15T11:00:00Z",
			"clickcount": 50,
			"clicktrend": 5,
			"ssl_error": 0
		}`

		var station Station
		err := json.Unmarshal([]byte(jsonData), &station)

		assert.NoError(t, err)
		assert.Equal(t, "Test Radio", station.Name)
		assert.Equal(t, "US", station.CountryCode)
		assert.Equal(t, uint64(128), station.Bitrate)
		assert.Equal(t, uint64(100), station.Votes)
		assert.Equal(t, uint64(50), station.ClickCount)
		assert.True(t, bool(station.LastCheckOk))
		assert.False(t, bool(station.Hls))
		assert.False(t, bool(station.SslError))
	})

	t.Run("parses station with boolean as integers", func(t *testing.T) {
		jsonData := `{
			"stationuuid": "16a73a57-5dba-11e8-b0ce-52543be04c81",
			"name": "Test Radio",
			"hls": 1,
			"lastcheckok": 0,
			"ssl_error": 1
		}`

		var station Station
		err := json.Unmarshal([]byte(jsonData), &station)

		assert.NoError(t, err)
		assert.True(t, bool(station.Hls))
		assert.False(t, bool(station.LastCheckOk))
		assert.True(t, bool(station.SslError))
	})

	t.Run("parses station with unquoted booleans", func(t *testing.T) {
		// BoolFromlInt accepts unquoted true/false (native JSON booleans)
		jsonData := `{
			"stationuuid": "16a73a57-5dba-11e8-b0ce-52543be04c81",
			"name": "Test Radio",
			"hls": true,
			"lastcheckok": false
		}`

		var station Station
		err := json.Unmarshal([]byte(jsonData), &station)

		assert.NoError(t, err)
		assert.True(t, bool(station.Hls))
		assert.False(t, bool(station.LastCheckOk))
	})

	t.Run("rejects quoted string booleans", func(t *testing.T) {
		// BoolFromlInt does NOT accept quoted strings like "true" or "false"
		// The API returns integers (0/1) not strings
		jsonData := `{
			"stationuuid": "16a73a57-5dba-11e8-b0ce-52543be04c81",
			"name": "Test Radio",
			"hls": "true"
		}`

		var station Station
		err := json.Unmarshal([]byte(jsonData), &station)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "boolean from int unmarshal error")
	})

	t.Run("parses station with null click timestamp", func(t *testing.T) {
		jsonData := `{
			"stationuuid": "16a73a57-5dba-11e8-b0ce-52543be04c81",
			"name": "Test Radio",
			"clicktimestamp_iso8601": null
		}`

		var station Station
		err := json.Unmarshal([]byte(jsonData), &station)

		assert.NoError(t, err)
		assert.Nil(t, station.ClickTimestamp)
	})

	t.Run("parses station with optional geo coordinates", func(t *testing.T) {
		jsonData := `{
			"stationuuid": "16a73a57-5dba-11e8-b0ce-52543be04c81",
			"name": "Test Radio",
			"geo_lat": 37.7749,
			"geo_long": -122.4194
		}`

		var station Station
		err := json.Unmarshal([]byte(jsonData), &station)

		assert.NoError(t, err)
		assert.NotNil(t, station.GeoLat)
		assert.NotNil(t, station.GeoLong)
		assert.InDelta(t, 37.7749, *station.GeoLat, 0.0001)
		assert.InDelta(t, -122.4194, *station.GeoLong, 0.0001)
	})

	t.Run("parses station with negative click trend", func(t *testing.T) {
		jsonData := `{
			"stationuuid": "16a73a57-5dba-11e8-b0ce-52543be04c81",
			"name": "Test Radio",
			"clicktrend": -15
		}`

		var station Station
		err := json.Unmarshal([]byte(jsonData), &station)

		assert.NoError(t, err)
		assert.Equal(t, int64(-15), station.ClickTrend)
	})

	t.Run("parses station with zero values", func(t *testing.T) {
		jsonData := `{
			"stationuuid": "00000000-0000-0000-0000-000000000000",
			"name": "",
			"bitrate": 0,
			"votes": 0,
			"clickcount": 0
		}`

		var station Station
		err := json.Unmarshal([]byte(jsonData), &station)

		assert.NoError(t, err)
		assert.Equal(t, "", station.Name)
		assert.Equal(t, uint64(0), station.Bitrate)
		assert.Equal(t, uint64(0), station.Votes)
		assert.Equal(t, uint64(0), station.ClickCount)
	})

	t.Run("handles malformed JSON", func(t *testing.T) {
		jsonData := `{invalid json}`

		var station Station
		err := json.Unmarshal([]byte(jsonData), &station)

		assert.Error(t, err)
	})

	t.Run("handles invalid UUID", func(t *testing.T) {
		jsonData := `{
			"stationuuid": "not-a-valid-uuid",
			"name": "Test Radio"
		}`

		var station Station
		err := json.Unmarshal([]byte(jsonData), &station)

		assert.Error(t, err)
	})

	t.Run("handles unicode in station name", func(t *testing.T) {
		jsonData := `{
			"stationuuid": "16a73a57-5dba-11e8-b0ce-52543be04c81",
			"name": "日本語ラジオ",
			"tags": "日本,音楽"
		}`

		var station Station
		err := json.Unmarshal([]byte(jsonData), &station)

		assert.NoError(t, err)
		assert.Equal(t, "日本語ラジオ", station.Name)
		assert.Equal(t, "日本,音楽", station.Tags)
	})

	t.Run("handles special characters in URL", func(t *testing.T) {
		jsonData := `{
			"stationuuid": "16a73a57-5dba-11e8-b0ce-52543be04c81",
			"name": "Test Radio",
			"url": "http://example.com/stream?param=value&other=123"
		}`

		var station Station
		err := json.Unmarshal([]byte(jsonData), &station)

		assert.NoError(t, err)
		assert.Equal(t, "http://example.com/stream?param=value&other=123", station.Url.URL.String())
	})
}
