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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStationQuery_Render(t *testing.T) {

	tests := []struct {
		query    StationQuery
		expected string
	}{
		{StationQueryByUuid, "By UUID"},
		{StationQueryByName, "By Name"},
		{StationQueryByNameExact, "By Exact Name"},
		{StationQueryByCodec, "By Codec"},
		{StationQueryByCodecExact, "By Exact Codec"},
		{StationQueryByCountry, "By Country"},
		{StationQueryByCountryExact, "By Exact Country"},
		{StationQueryByCountryCodeExact, "By Exact Country Code"},
		{StationQueryByState, "By State"},
		{StationQueryByStateExact, "By Exact State"},
		{StationQueryByLanguage, "By Language"},
		{StationQueryByLanguageExact, "By Exact Language"},
		{StationQueryByTag, "By Tag"},
		{StationQueryByTagExact, "By Exact Tag"},
		{StationQueryAll, "None"},
		{StationQuery("unknown"), "None"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.query.Render())
		})
	}
}

func TestStationQuery_ExampleString(t *testing.T) {

	t.Run("returns non-empty string for ByName", func(t *testing.T) {
		example := StationQueryByName.ExampleString()
		assert.NotEmpty(t, example)
		assert.Contains(t, example, "BBC Radio")
	})

	t.Run("returns non-empty string for ByNameExact", func(t *testing.T) {
		example := StationQueryByNameExact.ExampleString()
		assert.NotEmpty(t, example)
		assert.Contains(t, example, "BBC Radio 1")
	})

	t.Run("returns non-empty string for ByCodec", func(t *testing.T) {
		example := StationQueryByCodec.ExampleString()
		assert.NotEmpty(t, example)
		assert.Contains(t, example, "mp3")
	})

	t.Run("returns non-empty string for ByCountry", func(t *testing.T) {
		example := StationQueryByCountry.ExampleString()
		assert.NotEmpty(t, example)
		assert.Contains(t, example, "Italy")
	})

	t.Run("returns non-empty string for ByLanguage", func(t *testing.T) {
		example := StationQueryByLanguage.ExampleString()
		assert.NotEmpty(t, example)
		assert.Contains(t, example, "Italian")
	})

	t.Run("returns non-empty string for ByTag", func(t *testing.T) {
		example := StationQueryByTag.ExampleString()
		assert.NotEmpty(t, example)
		assert.Contains(t, example, "rock")
	})

	t.Run("returns empty string for unknown query", func(t *testing.T) {
		example := StationQuery("unknown").ExampleString()
		assert.Empty(t, example)
	})

	t.Run("returns empty string for StationQueryAll", func(t *testing.T) {
		example := StationQueryAll.ExampleString()
		assert.Empty(t, example)
	})
}
