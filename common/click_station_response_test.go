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

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestClickStationResponse_JSONUnmarshal(t *testing.T) {
	t.Run("parses successful response", func(t *testing.T) {
		jsonData := `{
			"ok": true,
			"message": "retrieved station url",
			"stationuuid": "9617a958-0601-11e8-ae97-52543be04c81",
			"name": "Test Station",
			"url": "http://example.com/stream"
		}`

		var response ClickStationResponse
		err := json.Unmarshal([]byte(jsonData), &response)

		assert.NoError(t, err)
		assert.True(t, response.Ok)
		assert.Equal(t, "retrieved station url", response.Message)
		assert.Equal(t, uuid.MustParse("9617a958-0601-11e8-ae97-52543be04c81"), response.StationUuid)
		assert.Equal(t, "Test Station", response.Name)
		assert.Equal(t, "http://example.com/stream", response.Url.URL.String())
	})

	t.Run("parses failed response", func(t *testing.T) {
		jsonData := `{
			"ok": false,
			"message": "station not found",
			"stationuuid": "00000000-0000-0000-0000-000000000000",
			"name": "",
			"url": ""
		}`

		var response ClickStationResponse
		err := json.Unmarshal([]byte(jsonData), &response)

		assert.NoError(t, err)
		assert.False(t, response.Ok)
		assert.Equal(t, "station not found", response.Message)
		assert.Equal(t, uuid.Nil, response.StationUuid)
	})

	t.Run("parses response with empty message", func(t *testing.T) {
		jsonData := `{
			"ok": true,
			"message": "",
			"stationuuid": "9617a958-0601-11e8-ae97-52543be04c81",
			"name": "Test Station",
			"url": "http://example.com/stream"
		}`

		var response ClickStationResponse
		err := json.Unmarshal([]byte(jsonData), &response)

		assert.NoError(t, err)
		assert.True(t, response.Ok)
		assert.Equal(t, "", response.Message)
	})

	t.Run("parses response with complex URL", func(t *testing.T) {
		jsonData := `{
			"ok": true,
			"message": "ok",
			"stationuuid": "9617a958-0601-11e8-ae97-52543be04c81",
			"name": "Test Station",
			"url": "http://example.com:8080/stream?token=abc123&quality=high"
		}`

		var response ClickStationResponse
		err := json.Unmarshal([]byte(jsonData), &response)

		assert.NoError(t, err)
		assert.Contains(t, response.Url.URL.String(), "token=abc123")
		assert.Contains(t, response.Url.URL.String(), "quality=high")
	})

	t.Run("handles malformed JSON", func(t *testing.T) {
		jsonData := `{ok: true, invalid}`

		var response ClickStationResponse
		err := json.Unmarshal([]byte(jsonData), &response)

		assert.Error(t, err)
	})

	t.Run("handles invalid UUID", func(t *testing.T) {
		jsonData := `{
			"ok": true,
			"message": "ok",
			"stationuuid": "not-a-uuid",
			"name": "Test",
			"url": "http://example.com"
		}`

		var response ClickStationResponse
		err := json.Unmarshal([]byte(jsonData), &response)

		assert.Error(t, err)
	})

	t.Run("handles unicode in name and message", func(t *testing.T) {
		jsonData := `{
			"ok": true,
			"message": "成功しました",
			"stationuuid": "9617a958-0601-11e8-ae97-52543be04c81",
			"name": "日本語ラジオ",
			"url": "http://example.com/stream"
		}`

		var response ClickStationResponse
		err := json.Unmarshal([]byte(jsonData), &response)

		assert.NoError(t, err)
		assert.Equal(t, "成功しました", response.Message)
		assert.Equal(t, "日本語ラジオ", response.Name)
	})
}
