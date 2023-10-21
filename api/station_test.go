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
	"radiogogo/common"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStationJSON(t *testing.T) {

	t.Run("parses from JSON", func(t *testing.T) {

		input := `
		[
		{
			"changeuuid": "610cafba-71d8-40fc-bf68-1456ec973b9d",
			"stationuuid": "941ef6f1-0699-4821-95b1-2b678e3ff62e",
			"serveruuid": "8a4a8315-6ff3-4af8-8ee7-24ce0acbaeec",
			"name": "\tBest FM",
			"url": "http://stream.bestfm.sk/128.mp3",
			"url_resolved": "http://stream.bestfm.sk/128.mp3",
			"homepage": "http://bestfm.sk/",
			"favicon": "",
			"tags": "",
			"country": "Slovakia",
			"countrycode": "SK",
			"iso_3166_2": null,
			"state": "",
			"language": "",
			"languagecodes": "",
			"votes": 57,
			"lastchangetime": "2022-11-01 08:40:32",
			"lastchangetime_iso8601": "2022-11-01T08:40:32Z",
			"codec": "MP3",
			"bitrate": 128,
			"hls": 0,
			"lastcheckok": 1,
			"lastchecktime": "2023-10-17 08:46:57",
			"lastchecktime_iso8601": "2023-10-17T08:46:57Z",
			"lastcheckoktime": "2023-10-17 08:46:57",
			"lastcheckoktime_iso8601": "2023-10-17T08:46:57Z",
			"lastlocalchecktime": "2023-10-17 08:46:57",
			"lastlocalchecktime_iso8601": "2023-10-17T08:46:57Z",
			"clicktimestamp": "2023-10-17 11:34:28",
			"clicktimestamp_iso8601": "2023-10-17T11:34:28Z",
			"clickcount": 45,
			"clicktrend": 3,
			"ssl_error": 0,
			"geo_lat": null,
			"geo_long": null,
			"has_extended_info": false
		}
		]
		`
		var stations []common.Station
		err := json.Unmarshal([]byte(input), &stations)

		assert.NoError(t, err)
		assert.Len(t, stations, 1)

	})

}
