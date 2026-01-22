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

func TestRadioGoGoURL_UnmarshalJSON(t *testing.T) {

	t.Run("unmarshals valid URL", func(t *testing.T) {
		var u RadioGoGoURL
		err := u.UnmarshalJSON([]byte(`"https://example.com/stream"`))
		assert.NoError(t, err)
		assert.Equal(t, "https", u.URL.Scheme)
		assert.Equal(t, "example.com", u.URL.Host)
		assert.Equal(t, "/stream", u.URL.Path)
	})

	t.Run("unmarshals URL with port", func(t *testing.T) {
		var u RadioGoGoURL
		err := u.UnmarshalJSON([]byte(`"http://radio.example.com:8080/live"`))
		assert.NoError(t, err)
		assert.Equal(t, "http", u.URL.Scheme)
		assert.Equal(t, "radio.example.com:8080", u.URL.Host)
		assert.Equal(t, "/live", u.URL.Path)
	})

	t.Run("unmarshals empty string as empty URL", func(t *testing.T) {
		var u RadioGoGoURL
		err := u.UnmarshalJSON([]byte(`""`))
		assert.NoError(t, err)
		assert.Empty(t, u.URL.String())
	})

	t.Run("returns error for invalid JSON", func(t *testing.T) {
		var u RadioGoGoURL
		err := u.UnmarshalJSON([]byte(`not-valid-json`))
		assert.Error(t, err)
	})

	t.Run("unmarshals URL with query parameters", func(t *testing.T) {
		var u RadioGoGoURL
		err := u.UnmarshalJSON([]byte(`"https://stream.example.com/play?format=mp3&quality=high"`))
		assert.NoError(t, err)
		assert.Equal(t, "mp3", u.URL.Query().Get("format"))
		assert.Equal(t, "high", u.URL.Query().Get("quality"))
	})
}
