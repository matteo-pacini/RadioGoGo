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

func TestVoteStationResponse_JSONUnmarshal(t *testing.T) {
	t.Run("parses successful vote response", func(t *testing.T) {
		jsonData := `{
			"ok": true,
			"message": "voted for station successfully"
		}`

		var response VoteStationResponse
		err := json.Unmarshal([]byte(jsonData), &response)

		assert.NoError(t, err)
		assert.True(t, response.Ok)
		assert.Equal(t, "voted for station successfully", response.Message)
	})

	t.Run("parses failed vote response (rate limited)", func(t *testing.T) {
		jsonData := `{
			"ok": false,
			"message": "you can only vote once every 10 minutes"
		}`

		var response VoteStationResponse
		err := json.Unmarshal([]byte(jsonData), &response)

		assert.NoError(t, err)
		assert.False(t, response.Ok)
		assert.Contains(t, response.Message, "10 minutes")
	})

	t.Run("parses failed vote response (station not found)", func(t *testing.T) {
		jsonData := `{
			"ok": false,
			"message": "station not found"
		}`

		var response VoteStationResponse
		err := json.Unmarshal([]byte(jsonData), &response)

		assert.NoError(t, err)
		assert.False(t, response.Ok)
		assert.Equal(t, "station not found", response.Message)
	})

	t.Run("parses response with empty message", func(t *testing.T) {
		jsonData := `{
			"ok": true,
			"message": ""
		}`

		var response VoteStationResponse
		err := json.Unmarshal([]byte(jsonData), &response)

		assert.NoError(t, err)
		assert.True(t, response.Ok)
		assert.Equal(t, "", response.Message)
	})

	t.Run("parses minimal response", func(t *testing.T) {
		jsonData := `{"ok": true}`

		var response VoteStationResponse
		err := json.Unmarshal([]byte(jsonData), &response)

		assert.NoError(t, err)
		assert.True(t, response.Ok)
		assert.Equal(t, "", response.Message)
	})

	t.Run("handles malformed JSON", func(t *testing.T) {
		jsonData := `{ok: true}`

		var response VoteStationResponse
		err := json.Unmarshal([]byte(jsonData), &response)

		assert.Error(t, err)
	})

	t.Run("handles empty JSON object", func(t *testing.T) {
		jsonData := `{}`

		var response VoteStationResponse
		err := json.Unmarshal([]byte(jsonData), &response)

		assert.NoError(t, err)
		assert.False(t, response.Ok) // Default value for bool
		assert.Equal(t, "", response.Message)
	})

	t.Run("handles unicode in message", func(t *testing.T) {
		jsonData := `{
			"ok": true,
			"message": "投票が成功しました"
		}`

		var response VoteStationResponse
		err := json.Unmarshal([]byte(jsonData), &response)

		assert.NoError(t, err)
		assert.Equal(t, "投票が成功しました", response.Message)
	})

	t.Run("handles long message", func(t *testing.T) {
		longMessage := "This is a very long message that might be returned by the server in some edge cases when something unexpected happens. It should be handled gracefully without truncation or errors."
		jsonData := `{
			"ok": false,
			"message": "` + longMessage + `"
		}`

		var response VoteStationResponse
		err := json.Unmarshal([]byte(jsonData), &response)

		assert.NoError(t, err)
		assert.Equal(t, longMessage, response.Message)
	})
}

func TestVoteStationResponse_JSONMarshal(t *testing.T) {
	t.Run("marshals to JSON correctly", func(t *testing.T) {
		response := VoteStationResponse{
			Ok:      true,
			Message: "voted successfully",
		}

		data, err := json.Marshal(response)

		assert.NoError(t, err)
		assert.Contains(t, string(data), `"ok":true`)
		assert.Contains(t, string(data), `"message":"voted successfully"`)
	})
}
